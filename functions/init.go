package functions

import (
	"fmt"
	"go/format"
	"os"
	"os/exec"
	"plugin"
	"runtime/debug"
	"strings"
	"time"

	"github.com/pericles-tpt/ownapi/config"
	"github.com/pericles-tpt/ownapi/utility"
	"github.com/pkg/errors"
)

var (
	customFunctionsPath         string
	generatedGoDir              string
	generatedGoFileForReference string
	generatedSOPath             string
)

var (
	tarValidateRetries = 3
	tarRetryDelay      = time.Second * 3

	pl *plugin.Plugin

	funcNames = []string{}
	funcs     = []CustomFunc{}
)

type CustomFunc struct {
	FuncComponentSignature
	F func([]any) ([]any, error)
}

func Init() error {
	for _, t := range typeWhitelist {
		arrayTypeWhitelist = append(arrayTypeWhitelist, fmt.Sprintf("[]%s", t))
	}

	tmpTarDst = config.GetDataDir("go.tar.gz")

	customFunctionsPath = config.GetDataDir("user_functions/source")
	generatedGoDir = fmt.Sprintf("%s/../generated", customFunctionsPath)
	generatedGoFileForReference = fmt.Sprintf("%s/main.go", generatedGoDir)
	generatedSOPath = fmt.Sprintf("%s/main.so", generatedGoDir)

	reload(true)
	return nil
}

func reload(initialLoad bool) bool {
	var (
		err error

		i         int
		tmpGoRoot string
	)
	defer maybePrintErr(&err)
	for i = 0; i < tarValidateRetries; i++ {
		var tmpDirForGo string
		tmpDirForGo, err = os.MkdirTemp("", "ownapi-*")
		if err != nil {
			err = errors.Wrap(err, "failed to create temp dir for go tooling")
			return false
		}
		tmpGoRoot = fmt.Sprintf("%s/go", tmpDirForGo)
		err = os.Mkdir(tmpGoRoot, 0700)
		if err != nil {
			err = errors.Wrapf(err, "failed to create go/ dir in temp path: %s", tmpDirForGo)
			return false
		}
		defer os.RemoveAll(tmpGoRoot)

		err = validateUnpackGoTar(tmpGoRoot)
		if err == nil {
			break
		}

		fmt.Printf("WARN: failed to validate/unpack go tar, err: %v\n", err)
		fmt.Printf("Sleeping for %v...\n", tarRetryDelay)

		utility.SleepLinux(tarRetryDelay)
	}
	if i == tarValidateRetries {
		err = fmt.Errorf("failed to validate and unpack go tar after %d attempts", tarValidateRetries)
		return false
	}

	var (
		filePaths, fileBasenames []string
		filesContents            [][]byte
	)
	filePaths, fileBasenames, filesContents, _, err = getFilesToCompile(customFunctionsPath, nil)
	if err != nil {
		return false
	}

	var (
		components = make([]FileComponents, 0, len(filesContents))
		fileErrors = make([]string, 0, len(filesContents))
	)
	for i, fc := range filesContents {
		c, err := DumbLexer(fc)
		if err != nil {
			fileErrors = append(fileErrors, fmt.Sprintf("\t%s: %s", fileBasenames[i], err.Error()))
		}
		components = append(components, c)
	}
	if len(fileErrors) > 0 {
		err = fmt.Errorf("failed to lex/parse the following files: %s\n", strings.Join(fileErrors, "\n"))
		return false
	}

	var numImports, numVars, numConsts, numPubFuncs, numPrivFuncs int
	for _, c := range components {
		numImports += len(c.Imports)
		numVars += len(c.Vars)
		numConsts += len(c.Consts)
		numPubFuncs += len(c.PublicFunctions)
		numPrivFuncs += len(c.PrivateFunctions)
	}
	combinedComponents := FileComponents{
		Imports:          make([]string, 0, numImports),
		Vars:             make(map[string]string, numVars),
		Consts:           make(map[string]string, numConsts),
		PublicFunctions:  make(map[string]FuncComponent, numPubFuncs),
		PrivateFunctions: make(map[string]FuncComponent, numPrivFuncs),
	}
	for _, c := range components {
		for _, imp := range c.Imports {
			utility.AddIfNotExists(&combinedComponents.Imports, imp)
		}
		utility.AddToMap(combinedComponents.Consts, c.Consts)
		utility.AddToMap(combinedComponents.Vars, c.Vars)
		utility.AddToMap(combinedComponents.PrivateFunctions, c.PrivateFunctions)
		utility.AddToMap(combinedComponents.PublicFunctions, c.PublicFunctions)
	}

	functionNames := make([]string, 0, len(combinedComponents.PublicFunctions))
	functionSignatures := make([]FuncComponentSignature, 0, len(combinedComponents.PublicFunctions))
	for name, pf := range combinedComponents.PublicFunctions {
		functionNames = append(functionNames, name)
		functionSignatures = append(functionSignatures, FuncComponentSignature{
			SigParams:      pf.SigParams,
			SigReturnTypes: pf.SigReturnTypes,
		})
	}

	generatedGoDir := fmt.Sprintf("%s/src/example.com/custom", tmpGoRoot)
	err = os.MkdirAll(generatedGoDir, 0700)
	if err != nil {
		err = errors.Wrapf(err, "failed to create dir for generated go code at: %s", generatedGoDir)
		return false
	}
	generatedGoFile := fmt.Sprintf("%s/main.go", generatedGoDir)

	var (
		tmpGeneratedGoFileForReference = fmt.Sprintf("%s.tmp", generatedGoFileForReference)
		tmpGeneratedSOPath             = fmt.Sprintf("%s.tmp", generatedSOPath)

		reloadPlugin bool
	)
	_, err = os.Stat(generatedGoFileForReference)
	if err == nil {
		err = os.Rename(generatedGoFileForReference, tmpGeneratedGoFileForReference)
		if err != nil {
			err = errors.Wrap(err, "failed to create backup of reference generated go file before regenerating code")
			return false
		}
	}
	_, err = os.Stat(generatedSOPath)
	if err == nil {
		err = os.Rename(generatedSOPath, tmpGeneratedSOPath)
		if err != nil {
			err = errors.Wrap(err, "failed to create backup of generated so file before regenerating code")
			return false
		}
	}
	defer tryRevertCodeGen(tmpGeneratedGoFileForReference, tmpGeneratedSOPath, functionSignatures, functionNames, &err, &reloadPlugin)
	reloadPlugin = initialLoad

	err = RegenerateUserCodeAsSharedObjectGo(combinedComponents, filePaths, []string{generatedGoFile, generatedGoFileForReference})
	if err != nil {
		err = errors.Wrap(err, "failed to generate output go from provided custom_functions")
		return false
	}

	// TODO: Make sure compilation here matches the main binary, otherwise could have problems
	// SOURCE: https://pkg.go.dev/plugin#hdr-Warnings
	goBinPath := fmt.Sprintf("%s/bin/go", tmpGoRoot)
	flagsToUnset := setBuildEnvVars()
	for _, f := range flagsToUnset {
		defer os.Unsetenv(f)
	}
	cmd := exec.Command(goBinPath, "build", "-buildmode=plugin", "-o", generatedSOPath, generatedGoFile)
	var out []byte
	out, err = cmd.CombinedOutput()
	if err != nil {
		err = errors.Wrapf(err, "failed to build plugin, out: %s", out)
		return false
	}

	reloadPlugin, err = reloadPluginAndFuncProps(generatedSOPath, functionSignatures, functionNames)
	return true
}

func maybePrintErr(err *error) {
	if err == nil || *err == nil {
		return
	}
	fmt.Printf("ERROR: Error occurred in auto reload: %s\n", (*err).Error())
}

func tryRevertCodeGen(tmpGenGoPath, tmpSOPath string, functionSignatures []FuncComponentSignature, functionNames []string, err *error, reloadPlugin *bool) {
	if err == nil || *err == nil {
		os.Remove(tmpGenGoPath)
		os.Remove(tmpSOPath)
		return
	}

	recovered := make([]string, 0, 3)
	if _, err1 := os.Stat(tmpGenGoPath); err1 == nil {
		originalGenGoPath := strings.TrimSuffix(tmpGenGoPath, ".tmp")
		err1 = os.Rename(tmpGenGoPath, originalGenGoPath)
		recovered = append(recovered, "old generated go file")
		if err1 != nil {
			recovered = recovered[:len(recovered)-1]
			fmt.Printf("WARN: failed to revert old generated go file after failed recompile, may exist at: %s\n", tmpGenGoPath)
		}
	}
	if _, err1 := os.Stat(tmpSOPath); err1 == nil {
		originalSOPath := strings.TrimSuffix(tmpSOPath, ".tmp")
		err1 = os.Rename(tmpSOPath, originalSOPath)
		recovered = append(recovered, "old so file")
		if err1 != nil {
			recovered = recovered[:len(recovered)-1]
			fmt.Printf("WARN: failed to revert old plugin file after failed recompile, may exist at: %s\n", tmpSOPath)
		} else if reloadPlugin != nil && *reloadPlugin {
			recovered = append(recovered, "reloaded old plugin")
			_, err1 = reloadPluginAndFuncProps(originalSOPath, functionSignatures, functionNames)
			if err1 != nil {
				recovered = recovered[:len(recovered)-1]
				fmt.Printf("WARN: failed to open restored plugin after failed recompile, exists at: %s\n", originalSOPath)
			}
		}
	}
	fmt.Printf("RECOVERED: %s\n", strings.Join(recovered, ", "))
}

func reloadPluginAndFuncProps(pluginPath string, functionSignatures []FuncComponentSignature, functionNames []string) (bool, error) {
	var (
		reloadPlugin bool
		err          error
	)
	pl, err = plugin.Open(pluginPath)
	if err != nil {
		return reloadPlugin, errors.Wrapf(err, "failed to open generated plugin at path: %s", pluginPath)
	}
	reloadPlugin = true

	// Populate local function properties
	funcNames = make([]string, len(functionSignatures))
	funcs = make([]CustomFunc, len(functionSignatures))
	for i, fs := range functionSignatures {
		name := functionNames[i]
		maybeFnc, err := pl.Lookup(name)
		if err != nil {
			return reloadPlugin, errors.Wrapf(err, "failed to find function in plugin with name '%s'", name)
		}

		var (
			fnc func([]any) ([]any, error)
			ok  bool
		)
		if fnc, ok = maybeFnc.(func([]any) ([]any, error)); !ok {
			return reloadPlugin, fmt.Errorf("failed to assert function with name '%s', as expected type `func([]any) ([]any, error)`", name)
		}

		funcNames[i] = name
		funcs[i] = CustomFunc{
			FuncComponentSignature: fs,
			F:                      fnc,
		}
	}

	return false, nil
}

func setBuildEnvVars() []string {
	bi, _ := debug.ReadBuildInfo()
	envs := make([]string, 0, len(bi.Settings))
	for _, s := range bi.Settings {
		if strings.HasPrefix(s.Key, "GO") || strings.HasPrefix(s.Key, "CGO") {
			os.Setenv(s.Key, s.Value)
			envs = append(envs, s.Key)
		}
	}
	return envs
}

func getFilesToCompile(root string, basenameModMap *map[string]time.Time) ([]string, []string, [][]byte, *[]bool, error) {
	var (
		filePaths     = []string{}
		fileBasenames = []string{}
		filesContents = [][]byte{}
		isNew         *[]bool
	)
	dirents, err := os.ReadDir(root)
	if err != nil {
		return filePaths, fileBasenames, filesContents, isNew, errors.Wrapf(err, "failed to read `%s`", root)
	}
	filePaths = make([]string, 0, len(dirents))
	fileBasenames = make([]string, 0, len(dirents))
	filesContents = make([][]byte, 0, len(dirents))
	if basenameModMap != nil {
		in := make([]bool, 0, len(dirents))
		isNew = &in
	}

	var (
		parents = []struct {
			lastChild int
			path      string
		}{
			{
				lastChild: len(dirents) - 1,
				path:      root,
			},
		}
		currParentIdx = 0
	)

	for i := 0; i < len(dirents); i++ {
		if i > parents[currParentIdx].lastChild {
			currParentIdx++
		}
		currParentPath := parents[currParentIdx].path
		currParentLastChild := parents[currParentIdx].lastChild

		var (
			de   = dirents[i]
			name = de.Name()
			path = fmt.Sprintf("%s/%s", currParentPath, name)
		)
		if de.IsDir() {
			nestedDirents, err := os.ReadDir(path)
			if err != nil {
				return filePaths, fileBasenames, filesContents, isNew, errors.Wrapf(err, "failed to read dir in root `%s`", path)
			}

			// NOTE: Only support 1 level of file nesting from root, currently
			nestedFiles := make([]os.DirEntry, 0, len(nestedDirents))
			for _, de := range nestedDirents {
				if de.Type().IsRegular() {
					nestedFiles = append(nestedFiles, de)
				}
			}
			dirents = append(dirents, nestedFiles...)

			parents = append(parents, struct {
				lastChild int
				path      string
			}{
				lastChild: currParentLastChild + len(nestedDirents),
				path:      path,
			})
			continue
		}

		if de.Type().IsRegular() && strings.HasSuffix(name, ".go") && name != "main.go" {
			if basenameModMap != nil {
				info, err := de.Info()
				if err != nil {
					fmt.Printf("WARN: Failed to read file: %s\n", name)
					continue
				}
				currLastModified := info.ModTime()

				(*isNew) = append((*isNew), false)
				var (
					prevLastModified time.Time
					exists           bool
				)
				if prevLastModified, exists = (*basenameModMap)[name]; !exists {
					(*isNew)[len(*isNew)-1] = true
				}
				isModified := prevLastModified != currLastModified
				funcsModified[name] = currLastModified

				if !isModified {
					continue
				}
			}

			var contents []byte
			contents, err := os.ReadFile(path)
			if err != nil {
				return filePaths, fileBasenames, filesContents, isNew, errors.Wrapf(err, "failed to read file '%s'", path)
			}

			// Format with `gofmt`, simplifies parsing a lot
			contents, err = format.Source(contents)
			if err != nil {
				return filePaths, fileBasenames, filesContents, isNew, errors.Wrapf(err, "failed to format go file '%s', likely invalid", path)
			}

			filesContents = append(filesContents, contents)
			filePaths = append(filePaths, path)
			fileBasenames = append(fileBasenames, name)
		}
	}

	return filePaths, fileBasenames, filesContents, isNew, nil
}

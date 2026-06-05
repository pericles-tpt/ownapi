package functions

import (
	"fmt"
	"go/format"
	"os"
	"os/exec"
	"plugin"
	"runtime/debug"
	"strings"

	"github.com/pericles-tpt/ownapi/utility"
	"github.com/pkg/errors"
)

const (
	customFunctionsPath = "./user_functions"
)

var (
	generatedGoDir              = fmt.Sprintf("%s/generated", customFunctionsPath)
	generatedGoFileForReference = fmt.Sprintf("%s/main.go", generatedGoDir)
	generatedSOPath             = fmt.Sprintf("%s/main.so", generatedGoDir)

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
	return Reload()
}

func Reload() error {
	tmpDirForGo, err := os.MkdirTemp("", "ownapi-*")
	if err != nil {
		return errors.Wrap(err, "failed to create temp dir for go tooling")
	}
	tmpGoRoot := fmt.Sprintf("%s/go", tmpDirForGo)
	err = os.Mkdir(tmpGoRoot, 0700)
	if err != nil {
		return errors.Wrapf(err, "failed to create go/ dir in temp path: %s", tmpDirForGo)
	}
	defer os.RemoveAll(tmpGoRoot)
	err = validateUnpackGoTar(tmpGoRoot)
	if err != nil {
		return errors.Wrap(err, "failed to check go binary")
	}

	dirents, err := os.ReadDir(customFunctionsPath)
	if err != nil {
		return errors.Wrapf(err, "failed to read `%s`", customFunctionsPath)
	}
	var (
		filePaths     = make([]string, 0, len(dirents))
		fileBasenames = make([]string, 0, len(dirents))
		filesContents = make([][]byte, 0, len(dirents))
	)
	for _, de := range dirents {
		var (
			name     = de.Name()
			filepath = fmt.Sprintf("%s/%s", customFunctionsPath, name)
		)
		if de.Type().IsRegular() && strings.HasSuffix(name, ".go") && name != "main.go" {
			var contents []byte
			contents, err = os.ReadFile(filepath)
			if err != nil {
				return errors.Wrapf(err, "failed to read file '%s'", filename)
			}

			// Format with `gofmt`, simplifies parsing a lot
			contents, err = format.Source(contents)
			if err != nil {
				return errors.Wrapf(err, "failed to format go file '%s', likely invalid", filename)
			}

			filesContents = append(filesContents, contents)
			filePaths = append(filePaths, filepath)
			fileBasenames = append(fileBasenames, name)
		}
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
		return fmt.Errorf("failed to lex/parse the following files: %s\n", strings.Join(fileErrors, "\n"))
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
		return errors.Wrapf(err, "failed to create dir for generated go code at: %s", generatedGoDir)
	}
	generatedGoFile := fmt.Sprintf("%s/main.go", generatedGoDir)

	err = RegenerateUserCodeAsSharedObjectGo(combinedComponents, filePaths, []string{generatedGoFile, generatedGoFileForReference})
	if err != nil {
		return errors.Wrap(err, "failed to generate output go from provided custom_functions")
	}

	// TODO: Make sure compilation here matches the main binary, otherwise could have problems
	// SOURCE: https://pkg.go.dev/plugin#hdr-Warnings
	goBinPath := fmt.Sprintf("%s/bin/go", tmpGoRoot)
	flagsToUnset := setBuildEnvVars()
	for _, f := range flagsToUnset {
		defer os.Unsetenv(f)
	}
	cmd := exec.Command(goBinPath, "build", "-buildmode=plugin", "-o", generatedSOPath, generatedGoFile)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, "failed to build plugin, out: %s", out)
	}

	pl, err = plugin.Open(generatedSOPath)
	if err != nil {
		return errors.Wrap(err, "failed to open generated plugin")
	}

	// Populate local function properties
	funcNames = make([]string, len(functionSignatures))
	funcs = make([]CustomFunc, len(functionSignatures))
	for i, fs := range functionSignatures {
		name := functionNames[i]
		maybeFnc, err := pl.Lookup(name)
		if err != nil {
			return errors.Wrapf(err, "failed to find function in plugin with name '%s'", name)
		}

		var (
			fnc func([]any) ([]any, error)
			ok  bool
		)
		if fnc, ok = maybeFnc.(func([]any) ([]any, error)); !ok {
			return fmt.Errorf("failed to assert function with name '%s', as expected type `func([]any) ([]any, error)`", name)
		}

		funcNames[i] = name
		funcs[i] = CustomFunc{
			FuncComponentSignature: fs,
			F:                      fnc,
		}
	}

	return nil
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

package functions

import (
	"fmt"
	"os"
	"strings"

	"github.com/pericles-tpt/ownapi/utility"
)

var (
	importWhitelist = []string{"fmt", "strings", "math", "time", "errors"}
	typeWhitelist   = []string{"string", "int", "uint", "rune", "byte", "int8", "int16", "int32", "int64", "uint8", "uint16", "uint32", "uint64", "float32", "float64", "bool", "time.Time", "error"}
	// Populated on Init()
	arrayTypeWhitelist = []string{}
)

func RegenerateUserCodeAsSharedObjectGo(components FileComponents, outPath string) error {
	err := validateComponents(components)
	if err != nil {
		return err
	}
	err = generateFiles(components, outPath)
	if err != nil {
		return err
	}

	return nil
}

func validateComponents(components FileComponents) error {
	// Validate imports
	invalidImports := make([]string, 0, len(components.Imports))
	for _, imp := range components.Imports {
		if _, ok := utility.Contains(imp, importWhitelist); !ok {
			invalidImports = append(invalidImports, imp)
		}
	}

	// Validate functions
	invalidParamReturns := make(map[string][2][]string)
	for funcName, funcComponents := range components.PrivateFunctions {
		var (
			paramProperties = funcComponents.SigParams
			returnTypes     = funcComponents.SigReturnTypes

			invalidParams  = make([]string, 0, len(paramProperties.Names))
			invalidReturns = make([]string, 0, len(returnTypes))
		)
		for i, pt := range paramProperties.Types {
			if valid, _ := isTypeValid(pt); !valid {
				invalidParams = append(invalidParams, paramProperties.Names[i])
			}
		}
		for _, r := range returnTypes {
			if valid, _ := isTypeValid(r); !valid {
				invalidReturns = append(invalidReturns, r)
			}
		}
		if len(invalidParams) > 0 || len(invalidReturns) > 0 {
			invalidParamReturns[funcName] = [2][]string{
				invalidParams,
				invalidReturns,
			}
		}
	}
	for funcName, funcComponents := range components.PublicFunctions {
		var (
			paramProperties = funcComponents.SigParams
			returnTypes     = funcComponents.SigReturnTypes

			invalidParams  = make([]string, 0, len(paramProperties.Names))
			invalidReturns = make([]string, 0, len(returnTypes))
		)
		for i, pt := range paramProperties.Types {
			if valid, _ := isTypeValid(pt); !valid {
				invalidParams = append(invalidParams, paramProperties.Names[i])
			}
		}
		for _, r := range returnTypes {
			if valid, _ := isTypeValid(r); !valid {
				invalidReturns = append(invalidReturns, r)
			}
		}
		if len(invalidParams) > 0 || len(invalidReturns) > 0 {
			invalidParamReturns[funcName] = [2][]string{
				invalidParams,
				invalidReturns,
			}
		}
	}

	// Generate errors (if any)
	errMsgParts := make([]string, 0, 2)
	if len(invalidImports) > 0 {
		errMsgParts = append(errMsgParts, fmt.Sprintf("invalid imports: %s", strings.Join(invalidImports, ", ")))
	}
	if len(invalidParamReturns) > 0 {
		invalidParamReturnsMessages := make([]string, 0, len(invalidParamReturns)*2)
		for funcName, paramReturns := range invalidParamReturns {
			if len(paramReturns[0]) > 0 {
				invalidParamReturnsMessages = append(invalidParamReturnsMessages, fmt.Sprintf("invalid param types in func `%s`: %s", funcName, strings.Join(paramReturns[0], ", ")))
			}
			if len(paramReturns[1]) > 0 {
				invalidParamReturnsMessages = append(invalidParamReturnsMessages, fmt.Sprintf("invalid return types in func `%s`: %s", funcName, strings.Join(paramReturns[1], ", ")))
			}
		}
		errMsgParts = append(errMsgParts, strings.Join(invalidParamReturnsMessages, "\n"))
	}
	if len(errMsgParts) > 0 {
		return fmt.Errorf("issues found: %s", strings.Join(errMsgParts, "\n"))
	}
	return nil
}

func generateFiles(components FileComponents, outPath string) error {
	sb := strings.Builder{}

	sb.WriteString("package main\n")

	for i, imp := range components.Imports {
		components.Imports[i] = fmt.Sprintf(`"%s"`, imp)
	}
	sb.WriteString(fmt.Sprintf("import (%s)\n", string(strings.Join(components.Imports, "\n"))))

	sb.WriteString("const (\n")
	for k, v := range components.Consts {
		sb.WriteString(fmt.Sprintf("%s %s\n", k, v))
	}
	sb.WriteString(")\n")

	sb.WriteString("var (\n")
	for k, v := range components.Vars {
		sb.WriteString(fmt.Sprintf("%s %s\n", k, v))
	}
	sb.WriteString(")\n")

	pubFuncLookupEntries := make([]string, 0, len(components.PublicFunctions))
	for name, pubFunc := range components.PublicFunctions {
		fs, err := generateFunction(name, pubFunc, true)
		if err != nil {
			return err
		}
		pubFuncLookupEntries = append(pubFuncLookupEntries, fmt.Sprintf(`"%s": %s`, name, name))
		sb.WriteString(fs)
	}
	for name, privFunc := range components.PrivateFunctions {
		fs, err := generateFunction(name, privFunc, false)
		if err != nil {
			return err
		}
		sb.WriteString(fs)
	}

	err := os.WriteFile(outPath, []byte(sb.String()), 0666)
	if err != nil {
		return err
	}

	return nil
}

func generateFunction(name string, f FuncComponent, isPublic bool) (string, error) {
	signature := fmt.Sprintf("func %s(args []any) ([]any, error) {", name)
	if !isPublic {
		signature = fmt.Sprintf("func %s(%s) (%s) {", name, getParamListString(f.SigParams), strings.Join(f.SigReturnTypes, ","))
	}

	var (
		paramNames, paramTypes = f.SigParams.Names, f.SigParams.Types
		assertions             = []string{}
	)
	if isPublic {
		assertions = make([]string, 0, len(f.SigParams.Names)+1)
		assertions = append(assertions, fmt.Sprintf("if len(args) != %d {\n return []any{}, fmt.Errorf(\"invalid number of args provided, exp: %d, got %%d\", len(args)) \n}\n", len(paramNames), len(paramNames)))
		for i, pt := range paramTypes {
			assertions = append(assertions, fmt.Sprintf("var %s %s\nif %s, ok = args[%d].(%s); !ok {\nreturn []any{}, fmt.Errorf(\"failed to parse arg %d with name '%s', expected type: %s\")\n}\n", paramNames[i], pt, paramNames[i], i, pt, i, paramNames[i], pt))
		}
	}

	// IF original return types contains an error, reorder return lines to put it at the end
	var (
		vars          string
		returnTypes   = f.SigReturnTypes
		skipReturnIdx = -1
	)
	if isPublic && len(paramNames) > 0 {
		vars = "var ok bool"
		for i := len(returnTypes) - 1; i > -1; i-- {
			if returnTypes[i] == "error" {
				skipReturnIdx = i
				break
			}
		}
	}

	// Reformat return replacements, replace placeholders in body
	bodyString := string(f.Body)
	for i, rv := range f.BodyReturnValues {
		errVal := "nil"

		vals := customReturnValuesSplit(rv, len(returnTypes))
		newVals := make([]string, 0, len(vals))
		for j, v := range vals {
			if j == skipReturnIdx {
				errVal = v
			} else {
				newVals = append(newVals, v)
			}
		}

		reformatted := strings.Join(newVals, ",")
		if isPublic {
			reformatted = fmt.Sprintf("[]any{%s}, %s", strings.Join(newVals, ","), errVal)
		}
		bodyString = strings.Replace(bodyString, getReturnPlaceholder(i), reformatted, 1)
	}

	return fmt.Sprintf("%s\n%s\n%s\n%s\n}\n", signature, vars, strings.Join(assertions, "\n"), bodyString), nil
}

func getArgNameTypes(val []rune) ([]string, []string, error) {
	var (
		args  = strings.Split(string(val), ",")
		names = make([]string, 0, len(args))
		types = make([]string, 0, len(args))
	)
	for _, a := range args {
		a = strings.TrimSpace(a)
		// Either of the form `a int`, or `int`
		var (
			maybeNameTypeParts = strings.Split(a, " ")
			t                  = maybeNameTypeParts[0]
		)
		if len(maybeNameTypeParts) > 1 {
			names = append(names, maybeNameTypeParts[0])
			t = strings.Join(maybeNameTypeParts[1:], " ")
		}
		t = strings.TrimSpace(t)
		if len(t) > 0 {
			types = append(types, strings.TrimSpace(t))
		}
	}

	if len(names) != len(types) {
		return names, types, fmt.Errorf("number of len(names) != len(types): %d != %d", len(names), len(types))
	}
	return names, types, nil
}

func getArgTypes(val []rune) []string {
	var (
		args  = strings.Split(string(val), ",")
		types = make([]string, 0, len(args))
	)
	for _, a := range args {
		a = strings.TrimSpace(a)
		// Either of the form `a int`, or `int`
		var (
			maybeNameTypeParts = strings.Split(a, " ")
			t                  = maybeNameTypeParts[0]
		)
		if len(maybeNameTypeParts) > 1 {
			t = strings.Join(maybeNameTypeParts[1:], " ")
		}
		t = strings.TrimSpace(t)
		if len(t) > 0 {
			types = append(types, strings.TrimSpace(t))
		}
	}

	return types
}

func isTypeValid(t string) (bool, bool) {
	_, isValid := utility.Contains(t, typeWhitelist)
	_, isValidArray := utility.Contains(t, arrayTypeWhitelist)
	return isValid || isValidArray, isValidArray
}

// customReturnValuesSplit, used instead of regular Split() method since return values
//
// can include function calls with comma separated arguments, this accounts for that
func customReturnValuesSplit(returnValues string, count int) []string {
	var (
		curvies = 0
		arg     = make([]rune, 0, len(returnValues))
		ret     = make([]string, 0, count)
	)
	for _, r := range returnValues {
		isComma := r == rune(',')
		if curvies != 0 || !isComma {
			arg = append(arg, r)
		} else if isComma {
			ret = append(ret, string(arg))
			arg = arg[:0]
		}

		if r == rune('(') {
			curvies -= 1
		} else if r == rune(')') {
			curvies += 1
		}
	}
	if len(arg) > 0 {
		ret = append(ret, string(arg))
	}

	return ret
}

func getParamListString(params Properties) string {
	if len(params.Names) != len(params.Types) {
		return "INVALID_PARAMS"
	}
	props := make([]string, 0, len(params.Names))
	for i, n := range params.Names {
		props = append(props, fmt.Sprintf("%s %s", n, params.Types[i]))
	}
	return strings.Join(props, ", ")
}

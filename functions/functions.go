package functions

import (
	"fmt"
	"go/format"
	"os"
	"strings"
	"unicode"

	"github.com/pericles-tpt/ownapi/utility"
	"github.com/pkg/errors"
)

var (
	forbiddenKeywords = []string{"struct", "type"}
)

func DumbLexer(filename string) ([]string, map[string]string, map[string][3][]rune, map[string][3][]rune, error) {
	var (
		imports   = make([]string, 0, 20)
		vars      = map[string]string{}
		pubFuncs  = map[string][3][]rune{}
		privFuncs = map[string][3][]rune{}

		err error
	)

	contents, err := os.ReadFile(filename)
	if err != nil {
		return imports, vars, pubFuncs, privFuncs, errors.Wrapf(err, "failed to read file '%s'", filename)
	}

	// Format with `gofmt`, simplifies parsing a lot
	contents, err = format.Source(contents)
	if err != nil {
		return imports, vars, pubFuncs, privFuncs, errors.Wrapf(err, "failed to format go file '%s', likely invalid", filename)
	}

	// TODO: Should use this std parser to replace my custom code below
	// fset := token.NewFileSet()
	// _, err = parser.ParseFile(fset, "", contents, parser.AllErrors)
	// if err != nil {
	// 	return res, errors.Wrapf(err, "invalid go file '%s'", filename)
	// }

	runes := []rune(string(contents))
	var (
		ri           rune
		buf          = make([]rune, 0, 50)
		newlineCount int

		singleLineComment, multiLineComment bool
		skipWord                            bool
	)
	for i := 0; i < len(runes); i++ {
		ri = runes[i]
		if !unicode.IsSpace(ri) {
			buf = append(buf, ri)
			multilineEnd := multiLineComment && (ri == rune('*') && (i+1) < len(runes) && runes[i+1] == rune('/'))
			if multiLineComment && multilineEnd {
				multiLineComment = false
				buf = buf[:0]
				i++
			}
			continue
		}

		if ri == rune('\n') {
			newlineCount++
			if singleLineComment {
				buf = buf[:0]
			}
			singleLineComment = false
		}
		if skipWord {
			buf = buf[:0]
			skipWord = false
			continue
		}

		// handle word
		var (
			word = string(buf)
			nlc  int
		)

		switch word {
		case "//": // gut
			singleLineComment = true
		case "/*": // gut
			multiLineComment = true
		case "package": // gut
			skipWord = true
		case "import": // gut
			var newImports []string
			i, nlc, newImports = extractImports(i, runes)
			newlineCount += nlc
			imports = append(imports, newImports...)
		case "const": // gut
			fallthrough
		case "var": // gut
			i, nlc, err = extractVars(i, runes, vars)
			newlineCount += nlc
			if err != nil {
				return imports, vars, pubFuncs, privFuncs, errors.Wrap(err, "failed to extract const/var block")
			}
		case "func": // nacht gut
			var (
				key      string
				val      [3][]rune
				isPublic bool
			)

			i, nlc, key, val, isPublic = extractFunc(i, runes)
			newlineCount += nlc
			privFuncs[key] = val
			if isPublic {
				pubFuncs[key] = val
				delete(privFuncs, key)
			}
		// case "goto": // Not valid outside of functions, safe to ignore
		// case "return": // As above
		case "type": // gut
			return imports, vars, pubFuncs, privFuncs, fmt.Errorf("detected disallowed keyword: %s, please refactor to remove it", word)
		default:
			// Do nothing, for now
		}

		buf = buf[:0]
	}

	// TODO: Figure out why newline count is wrong

	return imports, vars, pubFuncs, privFuncs, nil
}

func extractVars(i int, runes []rune, vars map[string]string) (int, int, error) {
	var (
		j   int
		nlc int
		rj  rune

		key                 string
		endRune             = rune('\n')
		buf                 = make([]rune, 0, 50)
		captureUntilNewline bool
		captureUntilCurly   bool
	)
	for j = i; j < len(runes) && rj != endRune; j++ {
		rj = runes[j]
		if rj == rune('(') {
			endRune = rune(')')
		} else if rj == rune('=') {
			captureUntilNewline = true
		} else if rj == rune('\n') {
			nlc++
			if captureUntilNewline {
				err := assignIfNotForbidden(key, string(buf), vars)
				if err != nil {
					return j, nlc, err
				}
				captureUntilCurly = len(buf) > 0 && buf[len(buf)-1] == rune('{')
				captureUntilNewline = false
			}
		} else if unicode.IsSpace(rj) {
			if (!captureUntilNewline && !captureUntilCurly) && len(buf) > 0 {
				err := assignIfNotForbidden(key, string(buf), vars)
				if err != nil {
					return j, nlc, err
				}
				newKey := string(buf)
				if len(key) > 0 {
					newKey = ""
				}
				key = newKey
				buf = buf[:0]
			}
		} else if rj != endRune {
			if rj == rune(',') && !captureUntilCurly && len(key) == 0 {
				return j, nlc, errors.New("invalid multi-variable assignment detected")
			}
			buf = append(buf, rj)
			captureUntilCurly = captureUntilCurly && rj != rune('}')
		}
	}

	if len(buf) > 0 {
		err := assignIfNotForbidden(key, string(buf), vars)
		if err != nil {
			return j, nlc, err
		}
	}
	delete(vars, "")

	return j, nlc, nil
}

func extractImports(i int, runes []rune) (int, int, []string) {
	var (
		j   int
		nlc int
		rj  rune

		imports = make([]string, 0, 20)
		endRune = rune('\n')
		buf     = make([]rune, 0, 50)
	)
	for j = i; j < len(runes) && rj != endRune; j++ {
		rj = runes[j]
		if unicode.IsSpace(rj) {
			if rj == rune('\n') {
				nlc++
			}
			if len(buf) > 0 {
				imports = append(imports, strings.Trim(string(buf), `"`))
				buf = buf[:0]
			}
			continue
		}

		if rj != rune('(') {
			buf = append(buf, rj)
		} else {
			endRune = rune(')')
		}
	}

	if endRune == rune('\n') {
		nlc++
	}

	last := string(buf)
	if len(last) > 0 && last != ")" {
		imports = append(imports, strings.Trim(last, `"`))
	}

	return j, nlc, imports
}

func extractFunc(i int, runes []rune) (int, int, string, [3][]rune, bool) {
	var (
		j   = i
		nlc int
		rj  rune

		nameBuf   = make([]rune, 0, 100)
		paramBuf  = make([]rune, 0, 100)
		returnBuf = make([]rune, 0, 100)
		bodyBuf   = make([]rune, 0, 1000)
	)

	// Collect name
	for j = i; j < len(runes) && rj != rune('('); j++ {
		rj = maybeIncrementNewlineCount(runes, j, &nlc)
		if !unicode.IsSpace(rj) {
			nameBuf = append(nameBuf, rj)
		}
	}
	nameBuf = removeBrackets(nameBuf, nil)

	// Collect params
	var maybeParamsEnd bool
	for ; j < len(runes); j++ {
		rj = maybeIncrementNewlineCount(runes, j, &nlc)
		if maybeParamsEnd {
			if unicode.IsSpace(rj) {
				continue
			} else if rj == rune('{') || rj == rune('(') {
				break
			}
		}
		maybeParamsEnd = rj == rune(')')
		paramBuf = append(paramBuf, rj)
	}
	paramBuf = removeBrackets(paramBuf, nil)

	// Collect returns
	var (
		maybeReturnsEnd bool
		skipReturns     bool
	)
	for ; j < len(runes); j++ {
		rj = maybeIncrementNewlineCount(runes, j, &nlc)
		if !unicode.IsSpace(rj) {
			skipReturns = rj == rune('{')
			break
		}
	}
	if !skipReturns {
		for ; j < len(runes); j++ {
			rj = maybeIncrementNewlineCount(runes, j, &nlc)

			if maybeReturnsEnd {
				if unicode.IsSpace(rj) {
					continue
				} else if rj == rune('{') || rj == rune('(') {
					break
				}
			}
			maybeReturnsEnd = rj == rune(')')
			returnBuf = append(returnBuf, rj)
		}
	}
	returnBuf = removeBrackets(returnBuf, nil)

	// Collect body
	for ; j < len(runes) && rj != rune('{'); j++ {
		rj = maybeIncrementNewlineCount(runes, j, &nlc)
	}
	j++
	curlies := 1
	for ; j < len(runes) && curlies != 0; j++ {
		rj = maybeIncrementNewlineCount(runes, j, &nlc)
		if rj == rune('{') {
			curlies++
		} else if rj == rune('}') {
			curlies--
		}
		bodyBuf = append(bodyBuf, rj)
	}
	bodyBuf = removeBrackets(bodyBuf, &[2]rune{'{', '}'})

	isPublic := unicode.IsUpper(nameBuf[0])
	return j, nlc, string(nameBuf), [3][]rune{paramBuf, returnBuf, bodyBuf}, isPublic
}

func assignIfNotForbidden(key, value string, m map[string]string) error {
	i, containsForbiddenKeyword := utility.SubstringsInTarget(value, forbiddenKeywords)
	if containsForbiddenKeyword {
		return fmt.Errorf("forbidden keyword found in value - v: '%v', keyword: '%v'", value, forbiddenKeywords[i])
	}
	m[key] = value
	return nil
}

func removeBrackets(buf []rune, startEndSet *[2]rune) []rune {
	if len(buf) == 0 {
		return buf
	}

	var (
		startEnds = [2]rune{'(', ')'}
		start     = 0
		end       = len(buf) - 1
	)
	if startEndSet != nil {
		startEnds = *startEndSet
	}

	if buf[0] == startEnds[0] || buf[0] == startEnds[1] {
		start = 1
	}
	if buf[len(buf)-1] == startEnds[0] || buf[len(buf)-1] == startEnds[1] {
		end = len(buf) - 1
	}
	if end == 0 {
		return buf[start:]
	}
	return buf[start:end]
}

func maybeIncrementNewlineCount(runes []rune, i int, newlineCount *int) rune {
	ret := runes[i]
	if ret == rune('\n') {
		*newlineCount++
	}
	return ret
}

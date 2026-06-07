package functions

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/pericles-tpt/ownapi/utility"
	"github.com/pkg/errors"
)

type FileComponents struct {
	Imports          []string
	Vars             map[string]string
	Consts           map[string]string
	PublicFunctions  map[string]FuncComponent
	PrivateFunctions map[string]FuncComponent
}

type FuncComponent struct {
	FuncComponentSignature
	Comment          string
	Body             string
	BodyReturnValues []string
}

type FuncComponentSignature struct {
	SigParams      Properties
	SigReturnTypes []string
}

type Properties struct {
	Names []string
	Types []string
}

const ILLEGAL_CHAR_GO = "🟦"

var forbiddenKeywords = []string{"struct"}

func DumbLexer(contents []byte) (FileComponents, error) {
	var (
		ret = FileComponents{
			Imports:          make([]string, 0, 20),
			Vars:             map[string]string{},
			Consts:           map[string]string{},
			PublicFunctions:  map[string]FuncComponent{},
			PrivateFunctions: map[string]FuncComponent{},
		}

		err error
	)

	runes := []rune(string(contents))
	var (
		ri           rune
		buf          = make([]rune, 0, 50)
		commentBuf   = make([]rune, 0, 100)
		newlineCount int

		singleLineComment, multiLineComment bool
		resetOnNextComment                  bool
		skipWord                            bool
	)
	for i := 0; i < len(runes); i++ {
		var isNewline bool
		ri, isNewline = getRuneMaybeIncrementNewlineCount(runes, i, &newlineCount)
		if singleLineComment || multiLineComment {
			commentBuf = append(commentBuf, ri)
			if isNewline {
				if multiLineComment {
					multiLineComment = (i-2 < 0 || (runes[i-2] != rune('*') || runes[i-1] != rune('/')))
				}
				nextLineIsComment := i+2 < len(runes) && (multiLineComment || string(runes[i+1:i+3]) == "//" || string(runes[i+1:i+3]) == "/*")
				resetOnNextComment = !nextLineIsComment
				singleLineComment = false
			}
		} else if !unicode.IsSpace(ri) {
			buf = append(buf, ri)
			continue
		}

		if skipWord {
			buf = buf[:0]
			skipWord = false
			continue
		}

		var (
			word = string(buf)
			nlc  int
		)

		switch word {
		case "//":
			if resetOnNextComment {
				commentBuf = commentBuf[:0]
				resetOnNextComment = false
			}
			commentBuf = append(commentBuf, []rune{'/', '/', ri}...)
			singleLineComment = true
		case "/*":
			if resetOnNextComment {
				commentBuf = commentBuf[:0]
				resetOnNextComment = false
			}
			commentBuf = append(commentBuf, []rune{'/', '*', ri}...)
			multiLineComment = true
		case "package":
			skipWord = true
		case "import":
			var newImports []string
			i, nlc, newImports = extractImports(i, runes)
			newlineCount += nlc
			ret.Imports = append(ret.Imports, newImports...)
		case "const":
			fallthrough
		case "var":
			if word == "const" {
				i, nlc, err = extractVars(i, runes, ret.Consts)
			} else {
				i, nlc, err = extractVars(i, runes, ret.Vars)
			}
			newlineCount += nlc
			if err != nil {
				return ret, errors.Wrap(err, "failed to extract const/var block")
			}
		case "func":
			var (
				key      string
				val      FuncComponent
				isPublic bool
			)

			i, nlc, key, val, isPublic, err = extractFunc(i, runes, commentBuf)
			if err != nil {
				return ret, errors.Wrapf(err, "failed to extract func with name '%s'", key)
			}
			newlineCount += nlc
			ret.PrivateFunctions[key] = val
			if isPublic {
				ret.PublicFunctions[key] = val
				delete(ret.PrivateFunctions, key)
			}
		// case "goto": // Not valid outside of functions, safe to ignore
		// case "return": // As above
		case "type":
			// TODO: Implement `extractType`
		default:
			// Do nothing, for now
		}

		buf = buf[:0]
	}

	return ret, nil
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

		nextAssignHasEquals bool
	)
	for j = i; j < len(runes) && rj != endRune; j++ {
		var isNewline bool
		rj, isNewline = getRuneMaybeIncrementNewlineCount(runes, j, &nlc)
		if rj == rune('(') {
			endRune = rune(')')
		} else if rj == rune('=') {
			captureUntilNewline = true
			nextAssignHasEquals = true
		} else if isNewline {
			if captureUntilNewline {
				err := assignIfNotForbidden(key, string(buf), vars, nextAssignHasEquals)
				if err != nil {
					return j, nlc, err
				}
				captureUntilCurly = len(buf) > 0 && buf[len(buf)-1] == rune('{')
				captureUntilNewline = false
			}
		} else if unicode.IsSpace(rj) {
			if captureUntilNewline {
				buf = append(buf, rj)
			} else if !captureUntilCurly && len(buf) > 0 {
				err := assignIfNotForbidden(key, string(buf), vars, nextAssignHasEquals)
				if err != nil {
					return j, nlc, err
				}
				newKey := string(buf)
				if len(key) > 0 {
					newKey = ""
					nextAssignHasEquals = false
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
		err := assignIfNotForbidden(key, string(buf), vars, nextAssignHasEquals)
		if err != nil {
			return j, nlc, err
		}
	}
	delete(vars, "")

	// NOTE: Need to do this in all `extract` functions since if they end on a
	// 		 newline it'll be skipped in the next iteration of the calling context
	// 		 screwing up the line count
	return j - 1, nlc, nil
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
		rj, _ = getRuneMaybeIncrementNewlineCount(runes, j, &nlc)
		if unicode.IsSpace(rj) {
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

	last := string(buf)
	if len(last) > 0 && last != ")" {
		imports = append(imports, strings.Trim(last, `"`))
	}

	return j - 1, nlc, imports
}

func extractFunc(i int, runes []rune, commentBuf []rune) (int, int, string, FuncComponent, bool, error) {
	var (
		j        = i
		nlc      int
		name     string
		rj       rune
		isPublic bool

		val = FuncComponent{
			BodyReturnValues: make([]string, 0, 5),
		}
	)

	if len(commentBuf) > 0 {
		val.Comment = string(commentBuf)
	}

	// Collect name
	nameBuf := make([]rune, 0, 100)
	for j = i; j < len(runes) && rj != rune('('); j++ {
		rj, _ = getRuneMaybeIncrementNewlineCount(runes, j, &nlc)
		if !unicode.IsSpace(rj) {
			nameBuf = append(nameBuf, rj)
		}
	}
	name = string(removeBrackets(nameBuf, nil))

	// Collect params
	var (
		paramsBuf     = make([]rune, 0, 100)
		curvies   int = 1
	)
	for ; j < len(runes) && curvies != 0; j++ {
		rj, _ = getRuneMaybeIncrementNewlineCount(runes, j, &nlc)
		if rj == rune('(') {
			curvies++
		} else if rj == rune(')') {
			curvies--
		}
		paramsBuf = append(paramsBuf, rj)
	}
	paramsBuf = removeBrackets(paramsBuf, nil)
	names, types, err := getArgNameTypes(paramsBuf)
	if err != nil {
		return j, nlc, name, val, isPublic, errors.Wrap(err, "invalid provided func params")
	}
	val.SigParams = Properties{
		Names: names,
		Types: types,
	}

	// Collect returns
	var countCurvies bool
	for ; j < len(runes); j++ {
		rj, _ = getRuneMaybeIncrementNewlineCount(runes, j, &nlc)
		if !unicode.IsSpace(rj) {
			countCurvies = rj == rune('(')
			break
		}
	}
	if countCurvies {
		manualIncrement(runes, &j, &nlc)
		curvies = 1
	}
	var returnsBuf = make([]rune, 0, 100)
	for ; j < len(runes) && ((countCurvies && curvies != 0) || !countCurvies); j++ {
		rj, _ = getRuneMaybeIncrementNewlineCount(runes, j, &nlc)
		if !unicode.IsSpace(rj) {
			if countCurvies {
				if rj == rune('(') {
					curvies++
				} else if rj == rune(')') {
					curvies--
				}
			} else if rj == rune('{') {
				break
			}
		}
		returnsBuf = append(returnsBuf, rj)
	}
	if countCurvies {
		returnsBuf = removeBrackets(returnsBuf, nil)
	}
	types = getArgTypes(returnsBuf)

	val.SigReturnTypes = make([]string, len(types))
	copy(val.SigReturnTypes, types)

	// Collect body
	for ; j < len(runes) && rj != rune('{'); j++ {
		rj, _ = getRuneMaybeIncrementNewlineCount(runes, j, &nlc)
	}
	manualIncrement(runes, &j, &nlc)

	var (
		bodyBuf = make([]rune, 0, 1000)

		curlies     = 1
		wordBuf     = make([]rune, 0, 50)
		returnCount int
	)
	for ; j < len(runes) && curlies != 0; j++ {
		rj, _ = getRuneMaybeIncrementNewlineCount(runes, j, &nlc)

		if unicode.IsSpace(rj) {
			w := string(wordBuf)
			if w == "return" {
				var (
					k       int
					rk      rune
					endRune *rune = nil

					countCommas        bool
					foundFirstNonSpace bool

					returnValuesBuf = make([]rune, 0, 100)
				)
				for k = j + 1; k < len(runes) && (endRune == nil || rk != *endRune); k++ {
					var currIsNewline bool
					rk, currIsNewline = getRuneMaybeIncrementNewlineCount(runes, k, &nlc)

					foundFirstNonSpace = unicode.IsSpace(rk)
					returnHasBrackets := endRune == nil && foundFirstNonSpace && rk == rune('(')
					if returnHasBrackets {
						ne := rune(')')
						endRune = &ne
					}
					countCommas = !returnHasBrackets

					// Below handles cases like:
					// A: 	return 2, struct{ hello int }{
					// 			hello: 32,
					// 		},
					//		3
					// Note the open curly followed by the newline and comma followed by newline
					if countCommas && currIsNewline && (runes[k-1] != rune(',') && runes[k-1] != rune('{')) {
						break
					}
					returnValuesBuf = append(returnValuesBuf, rk)
				}
				j = k

				returnPlaceholder := getReturnPlaceholder(returnCount)
				bodyBuf = append(bodyBuf, []rune(fmt.Sprintf(" %s\n", returnPlaceholder))...)
				val.BodyReturnValues = append(val.BodyReturnValues, string(returnValuesBuf))
				returnCount++
			}

			wordBuf = wordBuf[:0]
		} else {
			wordBuf = append(wordBuf, rj)
		}

		if rj == rune('{') {
			curlies++
		} else if rj == rune('}') {
			curlies--
		}
		bodyBuf = append(bodyBuf, rj)
	}
	bodyBuf = removeBrackets(bodyBuf, &[2]rune{'{', '}'})
	val.Body = string(bodyBuf)

	isPublic = unicode.IsUpper(rune(name[0]))
	return j - 1, nlc, name, val, isPublic, nil
}

func assignIfNotForbidden(key, value string, m map[string]string, assignHasEquals bool) error {
	keyWithEquals := fmt.Sprintf("%s = ", key)
	if assignHasEquals {
		key = keyWithEquals
	}

	i, containsForbiddenKeyword := utility.SubstringsInTarget(value, forbiddenKeywords)
	if containsForbiddenKeyword {
		return fmt.Errorf("forbidden keyword found in value - v: '%v', keyword: '%v'", value, forbiddenKeywords[i])
	}
	m[key] = value
	return nil
}

func getReturnPlaceholder(idx int) string {
	return fmt.Sprintf("%s%d", ILLEGAL_CHAR_GO, idx)
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

func getRuneMaybeIncrementNewlineCount(runes []rune, i int, newlineCount *int) (rune, bool) {
	var (
		ret       = runes[i]
		isNewline = ret == rune('\n')
	)
	if isNewline {
		*newlineCount++
	}
	return ret, isNewline
}

func manualIncrement(runes []rune, i *int, newlineCount *int) {
	if i == nil || newlineCount == nil || *i >= len(runes) {
		return
	}
	if *i < len(runes) && runes[*i] == rune('\n') {
		*newlineCount++
	}
	*i++
}

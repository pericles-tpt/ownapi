package utility

import (
	"fmt"
	"strings"
)

func AnyToString(arr []any) []string {
	ret := make([]string, len(arr))
	for i, el := range arr {
		ret[i] = fmt.Sprint(el)
	}
	return ret
}

func SubstringsInTarget(target string, subs []string) (int, bool) {
	for i, s := range subs {
		if strings.Contains(target, s) {
			return i, true
		}
	}
	return -1, false
}

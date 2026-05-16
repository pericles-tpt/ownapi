package utility

import "fmt"

func AnyToString(arr []any) []string {
	ret := make([]string, len(arr))
	for i, el := range arr {
		ret[i] = fmt.Sprint(el)
	}
	return ret
}

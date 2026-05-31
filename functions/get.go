package functions

import "fmt"

func GetFunc(name string) (CustomFunc, error) {
	for i, n := range funcNames {
		if n == name {
			return funcs[i], nil
		}
	}
	return CustomFunc{}, fmt.Errorf("failed to find custom func with name '%s'", name)
}

package functions

import "fmt"

func GetFunc(name string) (func([]any) ([]any, error), error) {
	var (
		ret func([]any) ([]any, error)
		ok  bool
	)
	if ret, ok = funcMap[name]; ok {
		return ret, nil
	}
	return ret, fmt.Errorf("failed to find func with name '%s'", name)
}

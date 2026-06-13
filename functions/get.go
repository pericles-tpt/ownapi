package functions

import "fmt"

func GetFunc(name string) (func([]any) ([]any, error), error) {
	f, err := pl.Lookup(name)
	if err != nil {
		return nil, fmt.Errorf("failed to find func matching name '%s'", name)
	}

	if f, ok := f.(func([]any) ([]any, error)); ok {
		return f, nil
	}
	return nil, fmt.Errorf("func matching name '%s', doesn't have `func([]any) ([]any, error)` signature", name)
}

func GetFuncSignature(name string) (FuncComponentSignature, error) {
	for i, f := range funcNames {
		if f == name {
			if i >= len(funcs) {
				return FuncComponentSignature{}, fmt.Errorf("func matching name '%s' has index out of range of func signatures", name)
			}
			return funcs[i], nil
		}
	}
	return FuncComponentSignature{}, fmt.Errorf("failed to find func signature matching name '%s'", name)
}

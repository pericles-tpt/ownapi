package utility

import (
	"fmt"
)

var supportedTypes = []string{"int", "float64", "string", "bool"}

func ValidateType(val any, expType string) (bool, error) {
	var valid bool

	switch expType {
	case "int":
		_, valid = val.(int)
	case "float64":
		_, valid = val.(float64)
	case "string":
		_, valid = val.(string)
	case "bool":
		_, valid = val.(bool)
	default:
		return valid, fmt.Errorf("invalid type provided: %s", expType)
	}
	return valid, nil
}

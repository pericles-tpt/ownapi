package utility

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/pkg/errors"
)

func GetTypeFromMap[T any](m map[string]any, key string) (T, bool, error) {
	var (
		val       any
		found, ok bool

		ret T
	)
	if val, found = m[key]; !found {
		return ret, found, nil
	}

	if ret, ok = val.(T); !ok {
		var (
			expType, gotType = reflect.TypeOf(ret), reflect.TypeOf(val)
		)
		return ret, found, fmt.Errorf("invalid type at key '%s' - exp: %s, got: %s", key, expType, gotType)
	}

	return ret, found, nil
}

// OverrideTypeFromMap, relies on JSON tags in type T to override its original values from corresponding
//
// keys in the map, it DOESN'T do any type validation until the last `Unmarshal`
func OverrideTypeFromJSONMap[T any](original T, overrides map[string]any) (T, error) {
	if overrides == nil {
		return original, nil
	}

	originalBytes, err := json.Marshal(original)
	if err != nil {
		return original, errors.Wrapf(err, "failed to marshal 'original' type: %s", reflect.TypeOf(original))
	}

	updatedMap := map[string]any{}
	err = json.Unmarshal(originalBytes, &updatedMap)
	if err != nil {
		return original, errors.Wrap(err, "failed to unmarshal original to a map")
	}
	var modified bool
	for k := range updatedMap {
		expKey := fmt.Sprintf("input:%s", k)
		if newV, ok := overrides[expKey]; ok {
			updatedMap[k] = newV
			modified = true
		}
	}
	if !modified {
		return original, nil
	}

	originalWithOverridesBytes, err := json.Marshal(updatedMap)
	if err != nil {
		return original, errors.Wrap(err, "failed to marshal 'original with overrides' to bytes, this shouldn't happen...")
	}
	var originalWithOverrides T
	err = json.Unmarshal(originalWithOverridesBytes, &originalWithOverrides)
	if err != nil {
		return original, errors.Wrapf(err, "failed to unmarshal 'original with overrides' map to type `%s`", reflect.TypeOf(originalWithOverrides))
	}

	return originalWithOverrides, nil
}

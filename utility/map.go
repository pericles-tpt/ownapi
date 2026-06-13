package utility

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

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
	for k, v := range updatedMap {
		// Key matches
		var expKey = fmt.Sprintf("input:%s", k)
		if newV, ok := overrides[expKey]; ok {
			updatedMap[k] = newV
			modified = true
		}

		// Types matches (if string)
		switch t := v.(type) {
		case string:
			updatedMap[k] = replacePlaceholders(t, overrides)
			modified = true
		case []any:
			var (
				i    int
				newT = make([]any, 0, len(t))
			)
			for i = 0; i < len(t); i++ {
				var (
					s  string
					ok bool
				)
				if s, ok = t[i].(string); !ok {
					break
				}
				maybeNewS := replacePlaceholders(s, overrides)
				newT = append(newT, maybeNewS)
			}

			if i == len(t) {
				updatedMap[k] = newT
				modified = true
			}
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

func AddToMap[T comparable, V any](dst map[T]V, src map[T]V) {
	for k, v := range src {
		dst[k] = v
	}
}

// TODO: Improve this, it could definitely be faster, but it works at least
func replacePlaceholders(s string, placeholders map[string]any) string {
	for k, v := range placeholders {
		if vs, ok := v.(string); ok {
			var (
				newS       = vs
				exactMatch = s == k
			)

			if !exactMatch {
				newS = strings.ReplaceAll(s, k, vs)
			}
			s = newS
		}
	}
	return s
}

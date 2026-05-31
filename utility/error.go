package utility

import (
	"errors"
	"fmt"
	"strings"
)

func ErrOnAnyMatch[T comparable](arrs [][]T, errMsgPrefixes []string, match T) error {
	if len(arrs) != len(errMsgPrefixes) {
		return fmt.Errorf("unequal number of arrays and errMsgPrefixes: %d != %d", len(arrs), len(errMsgPrefixes))
	}

	invalidIndices := make([][]int, len(arrs))
	for i, arr := range arrs {
		invalidIndices[i] = make([]int, 0, len(arr))
	}
	for i, arr := range arrs {
		for j, el := range arr {
			if el == match {
				invalidIndices[i] = append(invalidIndices[i], j)
			}
		}
	}

	matchErrors := make([]string, 0, len(arrs))
	for i, invalid := range invalidIndices {
		if len(invalid) > 0 {
			matchErrors = append(matchErrors, fmt.Sprintf("%s: %v", errMsgPrefixes[i], invalid))
		}
	}
	if len(matchErrors) > 0 {
		return errors.New(strings.Join(matchErrors, ", "))
	}
	return nil
}

func ErrOnAnyMismatch[T comparable](a [][]T, b [][]T, errMsgPrefixes []string) error {
	if len(a) != len(b) {
		return fmt.Errorf("unequal number of arrays in a and b: %d != %d", len(a), len(b))
	}
	if len(a) != len(errMsgPrefixes) {
		return fmt.Errorf("unequal number of arrays a/b and errMsgPrefixes: %d != %d", len(a), len(errMsgPrefixes))
	}

	invalidIndices := make([][]int, len(a))
	for i, arr := range a {
		invalidIndices[i] = make([]int, 0, len(arr))
	}
	for i, arrA := range a {
		arrB := b[i]
		for j, elA := range arrA {
			elB := arrB[j]
			if elA != elB {
				invalidIndices[i] = append(invalidIndices[i], j)
			}
		}
	}

	mismatchErrors := make([]string, 0, len(a))
	for i, invalid := range invalidIndices {
		if len(invalid) > 0 {
			mismatchErrors = append(mismatchErrors, fmt.Sprintf("%s: %v", errMsgPrefixes[i], invalid))
		}
	}
	if len(mismatchErrors) > 0 {
		return errors.New(strings.Join(mismatchErrors, ", "))
	}
	return nil
}

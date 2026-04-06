package utility

func Contains[T comparable](target T, arr []T) (int, bool) {
	for i, el := range arr {
		if el == target {
			return i, true
		}
	}
	return -1, false
}

func ContainsAll[T comparable](targetArr []T, validSet []T) ([]T, []int, []T) {
	var (
		foundVals   = make([]T, 0, len(targetArr))
		foundIdxs   = make([]int, 0, len(targetArr))
		missingVals = make([]T, 0, len(targetArr))
	)
	for _, t := range targetArr {
		var found bool
		for j, el := range validSet {
			if t == el {
				foundVals = append(foundVals, t)
				foundIdxs = append(foundIdxs, j)
				found = true
				break
			}
		}
		if !found {
			missingVals = append(missingVals, t)
		}
	}
	return foundVals, foundIdxs, missingVals
}

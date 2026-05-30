package utility

func AddIfNotExists[T comparable](arr *[]T, elem T) {
	if arr == nil {
		return
	}
	for _, el := range *arr {
		if el == elem {
			return
		}
	}
	(*arr) = append((*arr), elem)
}

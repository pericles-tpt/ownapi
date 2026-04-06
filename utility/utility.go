package utility

func Optional[T any](val T) *T {
	return &val
}

func OptionalAny[T any](val T) *any {
	var anyVal any = val
	return &anyVal
}

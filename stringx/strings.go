package stringx

func ToString[T ~string](v T) string {
	return string(v)
}

func FromString[T ~string](v string) T {
	return T(v)
}

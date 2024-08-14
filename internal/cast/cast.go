package cast

import "unsafe"

func FromStrings[T ~string](input []string) []T {
	return *(*[]T)(unsafe.Pointer(&input))
}

func ToStrings[T ~string](input []T) []string {
	return *(*[]string)(unsafe.Pointer(&input))
}

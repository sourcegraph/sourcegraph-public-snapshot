package codehost_testing

func reverse[T any](src []T) []T {
	reversed := make([]T, 0, len(src))
	for i := len(src) - 1; i >= 0; i-- {
		reversed = append(reversed, src[i])
	}
	return reversed
}

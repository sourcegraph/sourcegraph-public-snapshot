package codehost_testing

func boolp(v bool) *bool {
	return &v
}

func strp(v string) *string {
	return &v
}

func reverse[T any](src []T) []T {
	reversed := make([]T, 0, len(src))
	for i := len(src) - 1; i >= 0; i-- {
		reversed = append(reversed, src[i])
	}
	return reversed
}

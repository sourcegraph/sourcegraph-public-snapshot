package command

// Flatten combines string values and (non-recursive) string slice values
// into a single string slice.
func Flatten(values ...any) []string {
	union := make([]string, 0, len(values))
	for _, value := range values {
		switch v := value.(type) {
		case string:
			union = append(union, v)
		case []string:
			union = append(union, v...)
		}
	}

	return union
}

// Intersperse returns a slice following the pattern `flag, v1, flag, v2, ...`.
func Intersperse(flag string, values []string) []string {
	interspersed := make([]string, 0, len(values))
	for _, v := range values {
		interspersed = append(interspersed, flag, v)
	}

	return interspersed
}

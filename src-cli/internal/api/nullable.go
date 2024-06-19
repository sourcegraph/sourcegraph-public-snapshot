package api

// NullInt returns a nullable int for use in a GraphQL variable, where -1 is
// treated as a nil value.
func NullInt(n int) *int {
	if n == -1 {
		return nil
	}
	return &n
}

// NullString returns a nullable string for use in a GraphQL variable, where ""
// is treated as a nil value.
func NullString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

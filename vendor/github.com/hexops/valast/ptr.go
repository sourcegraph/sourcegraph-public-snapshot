package valast

// Ptr returns a pointer to the given value.
func Ptr[V any](v V) *V {
	return &v
}

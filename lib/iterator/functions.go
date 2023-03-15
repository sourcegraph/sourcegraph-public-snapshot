package iterator

// From is a convenience function to create an iterator from the slice s.
//
// Note: this function keeps a reference to s, so do not mutate it.
func From[T any](s []T) *Iterator[T] {
	done := false
	return New(func() ([]T, error) {
		if done {
			return nil, nil
		}
		done = true
		return s, nil
	})
}

// Collect transforms the iterator it into a slice. It returns the slice and
// the value of Err.
func Collect[T any](it *Iterator[T]) ([]T, error) {
	var s []T
	for it.Next() {
		s = append(s, it.Current())
	}
	return s, it.Err()
}

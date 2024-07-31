package collections

type IterFunc[T any] func() (T, bool)

// ForEach returns the total number of iterations.
func (i IterFunc[T]) ForEach(do func(int, T)) int {
	for nIter := 0; ; nIter++ {
		val, ok := i()
		if !ok {
			return nIter
		}
		do(nIter, val)
	}
}

func MapIter[T, U any](ts []T, f func(T) U) IterFunc[U] {
	i := 0
	return func() (U, bool) {
		if i == len(ts) {
			var zero U
			return zero, false
		}
		u := f(ts[i])
		i++
		return u, true
	}
}

package slices

func Map[S, T any](list []S, f func(S) T) []T {
	ret := make([]T, len(list))
	for i, e := range list {
		ret[i] = f(e)
	}
	return ret
}

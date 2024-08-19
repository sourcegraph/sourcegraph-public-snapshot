package slices

import (
	"sort"

	"github.com/life4/genesis/constraints"
)

// All returns true if f returns true for all items.
func All[S ~[]T, T any](items S, f func(el T) bool) bool {
	for _, el := range items {
		if !f(el) {
			return false
		}
	}
	return true
}

// Any returns true if f returns true for any item in items.
func Any[S ~[]T, T any](items S, f func(el T) bool) bool {
	for _, el := range items {
		if f(el) {
			return true
		}
	}
	return false
}

// ChunkBy splits items on every element for which f returns a new value.
func ChunkBy[S ~[]T, T comparable, G comparable](items S, f func(el T) G) []S {
	chunks := make([]S, 0)
	if len(items) == 0 {
		return chunks
	}

	chunk := make([]T, 0)
	prev := f(items[0])
	chunk = append(chunk, items[0])

	for _, el := range items[1:] {
		curr := f(el)
		if curr != prev {
			chunks = append(chunks, chunk)
			chunk = make([]T, 0)
			prev = curr
		}
		chunk = append(chunk, el)
	}
	if len(chunk) > 0 {
		chunks = append(chunks, chunk)
	}
	return chunks
}

// CountBy returns how many times f returns true.
func CountBy[S ~[]T, T any](items S, f func(el T) bool) int {
	count := 0
	for _, el := range items {
		if f(el) {
			count++
		}
	}
	return count
}

// DedupBy returns a copy of items, but without consecutive elements
// for which f returns the same result.
func DedupBy[S ~[]T, T any, G comparable](items S, f func(el T) G) S {
	result := make(S, 0, len(items))
	if len(items) == 0 {
		return result
	}

	prev := f(items[0])
	result = append(result, items[0])
	for _, el := range items[1:] {
		curr := f(el)
		if curr != prev {
			result = append(result, el)
			prev = curr
		}
	}
	return result
}

// DropWhile drops elements from the start of items while f returns true.
func DropWhile[S ~[]T, T any](items S, f func(el T) bool) S {
	for i, el := range items {
		if !f(el) {
			return items[i:]
		}
	}
	return []T{}
}

// Each calls f for each item in items.
func Each[S ~[]T, T any](items S, f func(el T)) {
	for _, el := range items {
		f(el)
	}
}

// EachErr calls f for each element in items until f returns an error.
func EachErr[S ~[]E, E any](items S, f func(el E) error) error {
	var err error
	for _, el := range items {
		err = f(el)
		if err != nil {
			return err
		}
	}
	return err
}

// EqualBy returns true if the cmp function returns true for all element pairs
// in the two slices.
//
// If the slices are different lengths, false is returned.
func EqualBy[S1 ~[]E1, S2 ~[]E2, E1, E2 any](s1 S1, s2 S2, eq func(E1, E2) bool) bool {
	if len(s1) != len(s2) {
		return false
	}
	for i, v1 := range s1 {
		if !eq(v1, s2[i]) {
			return false
		}
	}
	return true
}

// Filter returns a slice containing only items where f returns true.
func Filter[S ~[]T, T any](items S, f func(el T) bool) S {
	result := make([]T, 0, len(items))
	for _, el := range items {
		if f(el) {
			result = append(result, el)
		}
	}
	return result
}

// Find returns the first element for which f returns true.
func Find[S ~[]T, T any](items S, f func(el T) bool) (T, error) {
	for _, el := range items {
		if f(el) {
			return el, nil
		}
	}
	var tmp T
	return tmp, ErrNotFound
}

// FindIndex returns the index of the first element for which f returns true.
// Returns -1 if no matching element is found.
func FindIndex[S ~[]T, T any](items S, f func(el T) bool) int {
	for i, el := range items {
		if f(el) {
			return i
		}
	}
	return -1
}

// GroupBy groups items by the value returned by f.
func GroupBy[S ~[]T, T any, K comparable](items S, f func(el T) K) map[K]S {
	result := make(map[K]S)
	for _, el := range items {
		key := f(el)
		result[key] = append(result[key], el)
	}
	return result
}

// IndexBy returns the first index in items for which f returns true.
func IndexBy[S []T, T comparable](items S, f func(T) bool) (int, error) {
	for i, val := range items {
		if f(val) {
			return i, nil
		}
	}
	return 0, ErrNotFound
}

// Map applies f to all elements in items and returns a slice of the results.
func Map[S ~[]T, T any, G any](items S, f func(el T) G) []G {
	result := make([]G, 0, len(items))
	for _, el := range items {
		result = append(result, f(el))
	}
	return result
}

// MapFilter applies f to all elements in items, and returns a filtered slice of the results.
//
// f returns two values: the mapped value, and a boolean indicating whether the
// result should be included in the filtered slice.
func MapFilter[S ~[]T, T any, G any](items S, f func(el T) (G, bool)) []G {
	result := make([]G, 0, len(items))
	for _, el := range items {
		r, b := f(el)
		if b {
			result = append(result, r)
		}
	}
	return result
}

// Partition splits items into two slices based on if f returns true or false.
//
// The first returned slice contains the items for which f returned true.
// The second returned slice contains the remainder. Order is preserved in both
// slices.
func Partition[S ~[]T, T any](items S, f func(el T) bool) (S, S) {
	good := make(S, 0)
	bad := make(S, 0)
	for _, item := range items {
		if f(item) {
			good = append(good, item)
		} else {
			bad = append(bad, item)
		}
	}
	return good, bad
}

// Reduce applies f to acc and each element in items, reducing the slice to a
// single value.
func Reduce[S ~[]T, T any, G any](items S, acc G, f func(el T, acc G) G) G {
	for _, el := range items {
		acc = f(el, acc)
	}
	return acc
}

// ReduceWhile is like [Reduce], but stops when f returns an error.
func ReduceWhile[S ~[]T, T any, G any](items S, acc G, f func(el T, acc G) (G, error)) (G, error) {
	var err error
	for _, el := range items {
		acc, err = f(el, acc)
		if err != nil {
			return acc, err
		}
	}
	return acc, nil
}

// Reject returns a slice containing only items where f returns false.
// It is the opposite of [Filter].
func Reject[S ~[]T, T any](items S, f func(el T) bool) S {
	notF := func(el T) bool {
		return !f(el)
	}
	return Filter(items, notF)
}

// Scan applies f to acc and each element in items, feeding the output of the
// last call into the next call, and returning a slice of the results.
//
// Scan is like [Reduce], but it returns a slice of the results instead of a
// single value.
func Scan[S ~[]T, T any, G any](items S, acc G, f func(el T, acc G) G) []G {
	result := make([]G, 0, len(items))
	for _, el := range items {
		acc = f(el, acc)
		result = append(result, acc)
	}
	return result
}

// SortBy sorts items using the values returned by f.
//
// The function might be called more than once for the same element.
// It is expected to be fast and always produce the same result.
//
// The sort is stable. If two elements have the same ordering key,
// they are not swapped.
func SortBy[S ~[]T, T any, K constraints.Ordered](items S, f func(el T) K) S {
	if len(items) <= 1 {
		return items
	}
	less := func(i int, j int) bool {
		return f(items[i]) < f(items[j])
	}
	sort.SliceStable(items, less)
	return items
}

// TakeWhile takes elements from items while f returns true.
func TakeWhile[S ~[]T, T any](items S, f func(el T) bool) S {
	result := make(S, 0, len(items))
	for _, el := range items {
		if !f(el) {
			return result
		}
		result = append(result, el)
	}
	return result
}

// ToMapGroupedBy is an alias for [GroupBy].
//
// DEPRECATED. Use [GroupBy] instead.
func ToMapGroupedBy[V any, T comparable](items []V, keyExtractor func(V) T) map[T][]V {
	return GroupBy(items, keyExtractor)
}

package slices

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/life4/genesis/constraints"
)

// Choice chooses a random element from the slice.
// If seed is zero, UNIX timestamp will be used.
func Choice[S ~[]T, T any](items S, seed int64) (T, error) {
	if len(items) == 0 {
		var tmp T
		return tmp, ErrEmpty
	}

	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	r := rand.New(rand.NewSource(seed))
	i := r.Intn(len(items))
	return items[i], nil
}

// ChunkEvery splits items into groups of length count.
//
// If items can't be split evenly, the final chunk will contain the remaining elements.
func ChunkEvery[S ~[]T, T any](items S, count int) ([]S, error) {
	chunks := make([]S, 0)
	if count <= 0 {
		return chunks, ErrNegativeValue
	}
	chunk := make([]T, 0, count)
	for i, el := range items {
		chunk = append(chunk, el)
		if (i+1)%count == 0 {
			chunks = append(chunks, chunk)
			chunk = make([]T, 0, count)
		}
	}
	if len(chunk) > 0 {
		chunks = append(chunks, chunk)
	}
	return chunks, nil
}

// Contains returns true if el is in items.
func Contains[S ~[]T, T comparable](items S, el T) bool {
	for _, val := range items {
		if val == el {
			return true
		}
	}
	return false
}

// Copy creates a copy of the given slice.
func Copy[S ~[]T, T any](items S) S {
	if items == nil {
		return nil
	}
	var res S
	return append(res, items...)
}

// Count return the count of el occurrences in items.
func Count[S ~[]T, T comparable](items S, el T) int {
	count := 0
	for _, val := range items {
		if val == el {
			count++
		}
	}
	return count
}

// Cycle is an infinite loop over items.
func Cycle[S ~[]T, T any](items S) chan T {
	c := make(chan T)
	go func() {
		defer close(c)
		if len(items) == 0 {
			return
		}
		for {
			for _, val := range items {
				c <- val
			}
		}
	}()
	return c
}

// Dedup returns items without consecutive duplicated elements.
func Dedup[S ~[]T, T comparable](items S) S {
	if len(items) == 0 {
		return items
	}

	result := make(S, 1, len(items))
	prev := items[0]
	result[0] = prev
	for _, el := range items[1:] {
		if el != prev {
			result = append(result, el)
			prev = el
		}
	}
	return result
}

// Delete deletes the first occurrence of the element from items.
func Delete[S ~[]T, T comparable](items S, element T) S {
	if len(items) == 0 {
		return items
	}

	result := make([]T, 0, len(items))
	deleted := false
	for _, el := range items {
		if !deleted && el == element {
			deleted = true
			continue
		}
		result = append(result, el)
	}
	return result
}

// DeleteAll deletes all occurrences of the element from items.
func DeleteAll[S ~[]T, T comparable](items S, element T) S {
	if len(items) == 0 {
		return items
	}

	result := make([]T, 0, len(items))
	for _, el := range items {
		if el == element {
			continue
		}
		result = append(result, el)
	}
	return result

}

// DeleteAt returns items without the elements in indices.
func DeleteAt[S ~[]T, T any](items S, indices ...int) (S, error) {
	if len(indices) == 0 || len(items) == 0 {
		return items, nil
	}

	for _, index := range indices {
		if index >= len(items) {
			return items, ErrOutOfRange
		}
	}

	result := make([]T, 0, len(items)-1)
	for i, el := range items {
		add := true
		for _, index := range indices {
			if i == index {
				add = false
				break
			}
		}
		if add {
			result = append(result, el)
		}
	}
	return result, nil
}

// DropEvery returns a slice with every nth element in items dropped,
// starting with the first element.
func DropEvery[S ~[]T, T any](items S, nth int, from int) (S, error) {
	if nth <= 0 {
		return items, ErrNonPositiveValue
	}
	if from < 0 {
		return items, ErrNegativeValue
	}
	result := make([]T, 0, len(items)/nth)
	for i, el := range items {
		if (i+from)%nth != 0 {
			result = append(result, el)
		}
	}
	return result, nil
}

// DropZero returns a slice with every default value removed.
//
// For example, for a slice of pointers it will drop nils,
// for a slice of ints it will drop zero,
// and for a slice of strings it will drop empty strings.
func DropZero[S ~[]T, T comparable](items S) S {
	result := make(S, 0)
	var zero T
	for _, item := range items {
		if item != zero {
			result = append(result, item)
		}
	}
	return result
}

// EndsWith returns true if items ends with the given suffix slice.
//
// If suffix is empty, it returns true.
func EndsWith[S ~[]T, T comparable](items S, suffix S) bool {
	if len(suffix) > len(items) {
		return false
	}
	start := len(items) - len(suffix)
	for i, el := range suffix {
		if el != items[start+i] {
			return false
		}
	}
	return true
}

// Equal returns true if the slices are equal.
func Equal[S1 ~[]T, S2 ~[]T, T comparable](items S1, other S2) bool {
	if len(items) != len(other) {
		return false
	}
	for i, el := range other {
		if items[i] != el {
			return false
		}
	}
	return true
}

// Grow increases the slice capacity to fit at least n more elements.
//
// So, for len(slice)=8 and n=2, the result will have a capacity of at least 10.
//
// The function can be used to reduce allocations when inserting more elements
// into an existing slice.
//
// If the slice already has sufficient capacity, this slice is returned unmodified.
func Grow[S ~[]T, T any](items S, n int) S {
	return append(items, make(S, n)...)[:len(items)]
}

// Index returns the index of the first occurrence of item in items.
func Index[S ~[]T, T comparable](items S, item T) (int, error) {
	for i, val := range items {
		if val == item {
			return i, nil
		}
	}
	return 0, ErrNotFound
}

// InsertAt returns items with item inserted at the given index.
//
// If index is beyond the length of items, ErrOutOfRange is returned.
// If index is negative, ErrNegativeValue is returned.
func InsertAt[S ~[]T, T any](items S, index int, item T) (S, error) {
	result := make([]T, 0, len(items)+1)

	// insert at the end
	if index == len(items) {
		result = append(result, items...)
		result = append(result, item)
		return result, nil
	}

	if index > len(items) {
		return items, ErrOutOfRange
	}
	if index < 0 {
		return items, ErrNegativeValue
	}

	for i, el := range items {
		if i == index {
			result = append(result, item)
		}
		result = append(result, el)
	}
	return result, nil
}

// Intersperse inserts el between each element of items.
func Intersperse[S ~[]T, T any](items S, el T) S {
	if len(items) == 0 {
		return items
	}
	result := make([]T, 0, len(items)*2-1)
	result = append(result, items[0])
	for _, val := range items[1:] {
		result = append(result, el, val)
	}
	return result
}

// Join concatenates elements of items to create a single string.
func Join[S ~[]T, T any](items S, sep string) string {
	strs := make([]string, 0, len(items))
	for _, el := range items {
		strs = append(strs, fmt.Sprintf("%v", el))
	}
	return strings.Join(strs, sep)
}

// Last returns the last element from items.
func Last[S ~[]T, T any](items S) (T, error) {
	if len(items) == 0 {
		var tmp T
		return tmp, ErrEmpty
	}
	return items[len(items)-1], nil
}

// Max returns the maximal element in items.
func Max[S ~[]T, T constraints.Ordered](items S) (T, error) {
	if len(items) == 0 {
		var tmp T
		return tmp, ErrEmpty
	}

	max := items[0]
	for _, el := range items[1:] {
		if el > max {
			max = el
		}
	}
	return max, nil
}

// Min returns the minimal element from items.
func Min[S ~[]T, T constraints.Ordered](items S) (T, error) {
	if len(items) == 0 {
		var tmp T
		return tmp, ErrEmpty
	}

	min := items[0]
	for _, el := range items[1:] {
		if el < min {
			min = el
		}
	}
	return min, nil
}

// Permutations returns successive size-length permutations of elements from items.
func Permutations[S ~[]T, T any](items S, size int) chan S {
	c := make(chan S, 1)
	go func() {
		if len(items) > 0 {
			permutations(items, c, size, []T{}, items)
		}
		close(c)
	}()
	return c
}

// permutations is a core implementation for Permutations.
func permutations[S ~[]T, T any](items S, c chan S, size int, left []T, right []T) {
	if len(left) == size || len(right) == 0 {
		c <- left
		return
	}

	for i, el := range right {
		newLeft := make([]T, 0, len(left)+1)
		newLeft = append(newLeft, left...)
		newLeft = append(newLeft, el)

		newRight := make([]T, 0, len(right)-1)
		for j, other := range right {
			if j != i {
				newRight = append(newRight, other)
			}
		}
		permutations(items, c, size, newLeft, newRight)
	}
}

// Prepend returns the target slice with the given items added at the beginning.
func Prepend[S ~[]T, T any](target S, items ...T) S {
	return Concat(items, target)
}

// Product returns the cartesian product of items to itself, repeat times.
func Product[S ~[]T, T any](items S, repeat int) chan []T {
	c := make(chan []T, 1)
	go func() {
		defer close(c)
		if repeat < 1 {
			return
		}
		product(items, c, repeat, []T{}, 0)
	}()
	return c
}

// product is a core implementation for [Product]
func product[S ~[]T, T any](items S, c chan []T, repeat int, left []T, pos int) {
	// iterate over the last array
	if pos == repeat-1 {
		for _, el := range items {
			result := make([]T, 0, len(left)+1)
			result = append(result, left...)
			result = append(result, el)
			c <- result
		}
		return
	}

	for _, el := range items {
		result := make([]T, 0, len(left)+1)
		result = append(result, left...)
		result = append(result, el)
		product(items, c, repeat, result, pos+1)
	}
}

// Repeat repeats items n times.
func Repeat[S ~[]T, T any](items S, n int) S {
	result := make([]T, 0, len(items)*n)
	for i := 0; i < n; i++ {
		result = append(result, items...)
	}
	return result
}

// Replace replaces elements in items from start to end with the given item.
//
// The item with the end index is not replaced.
//
// Result:
//
//   - If start or end are negative, [ErrNegativeValue] is returned.
//   - If start is greater or equal to end, [ErrOutOfRange] is returned.
//   - If start or end is bigger than the slice len, [ErrOutOfRange] is returned.
func Replace[S ~[]T, T comparable, I constraints.Integer](items S, start, end I, item T) (S, error) {
	if start < 0 || end < 0 {
		return items, ErrNegativeValue
	}
	if start >= end {
		return items, ErrOutOfRange
	}
	l := I(len(items))
	if start > l || end > l {
		return items, ErrOutOfRange
	}
	result := make(S, 0, l)
	result = append(result, items[:start]...)
	for i := start; i < end; i++ {
		result = append(result, item)
	}
	result = append(result, items[end:]...)
	return result, nil
}

// Reverse returns items in reversed order.
func Reverse[S ~[]T, T any](items S) S {
	if len(items) <= 1 {
		return items
	}
	result := make([]T, 0, len(items))
	for i := len(items) - 1; i >= 0; i-- {
		result = append(result, items[i])
	}
	return result
}

// Same returns true if all elements in items are the same.
func Same[S ~[]T, T comparable](items S) bool {
	if len(items) <= 1 {
		return true
	}
	for i := 0; i < len(items)-1; i++ {
		if items[i] != items[i+1] {
			return false
		}
	}
	return true
}

// Shrink removes unused capacity from the slice.
//
// In other words, the returned slice has capacity equal to length.
func Shrink[S ~[]T, T any](items S) S {
	return items[:len(items):len(items)]
}

// Shuffle in random order the given elements.
//
// This is an in-place operation, it modifies the passed slice.
func Shuffle[S ~[]T, T any](items S, seed int64) {
	if len(items) <= 1 {
		return
	}
	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	r := rand.New(rand.NewSource(seed))
	swap := func(i, j int) {
		items[i], items[j] = items[j], items[i]
	}
	r.Shuffle(len(items), swap)
}

// Sort returns a sorted copy of items.
func Sort[S ~[]T, T constraints.Ordered](items S) S {
	if len(items) <= 1 {
		return items
	}
	less := func(i int, j int) bool {
		return items[i] < items[j]
	}
	sort.SliceStable(items, less)
	return items
}

// Sorted returns true if items is sorted.
func Sorted[S ~[]T, T constraints.Ordered](items S) bool {
	l := len(items)
	if l <= 1 {
		return true
	}
	for i := 1; i < l; i++ {
		if items[i-1] > items[i] {
			return false
		}
	}
	return true
}

// SortedUnique returns true if items is sorted and all elements are unique.
func SortedUnique[S ~[]T, T constraints.Ordered](items S) bool {
	l := len(items)
	if l <= 1 {
		return true
	}
	for i := 1; i < l; i++ {
		if items[i-1] >= items[i] {
			return false
		}
	}
	return true
}

// Split splits items by sep.
func Split[S ~[]T, T comparable](items S, sep T) []S {
	result := make([]S, 0)
	curr := make([]T, 0)
	for _, el := range items {
		if el == sep {
			result = append(result, curr)
			curr = make([]T, 0)
		} else {
			curr = append(curr, el)
		}
	}
	result = append(result, curr)
	return result
}

// StartsWith returns true if items starts with prefix.
// If prefix is empty, StartsWith returns true.
func StartsWith[S ~[]T, T comparable](items S, prefix S) bool {
	if len(prefix) > len(items) {
		return false
	}
	for i, el := range prefix {
		if el != items[i] {
			return false
		}
	}
	return true
}

// Sum return sum of all elements in items.
func Sum[S ~[]T, T constraints.Ordered](items S) T {
	var sum T
	for _, el := range items {
		sum += el
	}
	return sum
}

// TakeEvery returns a slice containing every nth element in items.
func TakeEvery[S ~[]T, T any](items S, nth int, from int) (S, error) {
	if nth <= 0 {
		return items, ErrNonPositiveValue
	}
	if from < 0 {
		return items, ErrNegativeValue
	}
	result := make(S, 0, len(items))
	for i, el := range items {
		if (i+from)%nth == 0 {
			result = append(result, el)
		}
	}
	return result, nil
}

// TakeRandom returns a slice of count random elements from items.
func TakeRandom[S ~[]T, T any](items S, count int, seed int64) (S, error) {
	if count > len(items) {
		return nil, ErrOutOfRange
	}
	if count <= 0 {
		return nil, ErrNonPositiveValue
	}

	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	r := rand.New(rand.NewSource(seed))
	swap := func(i, j int) {
		items[i], items[j] = items[j], items[i]
	}
	r.Shuffle(len(items), swap)
	return items[:count], nil
}

// ToChannel returns a channel with elements from items.
func ToChannel[S ~[]T, T any](items S) chan T {
	c := make(chan T)
	go func() {
		for _, el := range items {
			c <- el
		}
		close(c)
	}()
	return c
}

// ToKeys returns a map where the keys are the elements in items and all values
// are set to val.
func ToKeys[S ~[]K, K comparable, V any](items S, val V) map[K]V {
	if items == nil {
		return nil
	}
	result := make(map[K]V)
	for _, item := range items {
		result[item] = val
	}
	return result
}

// ToMap converts the given slice into a map where the keys are indices and the
// values are elements from items.
func ToMap[S ~[]V, V any](items S) map[int]V {
	if items == nil {
		return nil
	}
	result := make(map[int]V)
	for index, item := range items {
		result[index] = item
	}
	return result
}

// Uniq returns items with only the first occurrence of each unique element.
func Uniq[S ~[]T, T comparable](items S) S {
	if len(items) <= 1 {
		return items
	}
	added := make(map[T]struct{})
	nothing := struct{}{}
	result := make([]T, 0, len(items))
	for _, el := range items {
		_, exists := added[el]
		if !exists {
			result = append(result, el)
			added[el] = nothing
		}
	}
	return result

}

// Unique returns true if each element in items appears only once.
func Unique[S ~[]T, T comparable](items S) bool {
	seen := make(map[T]struct{})
	for _, item := range items {
		_, found := seen[item]
		if found {
			return false
		}
		seen[item] = struct{}{}
	}
	return true
}

// Window makes a sliding window for items.
func Window[S ~[]T, T any](items S, size int) ([]S, error) {
	if size <= 0 {
		return nil, ErrNonPositiveValue
	}
	result := make([]S, 0, len(items)/size)
	for i := 0; i <= len(items)-size; i++ {
		chunk := items[i : i+size]
		result = append(result, chunk)
	}
	return result, nil
}

// Without returns items without the given elements.
func Without[S ~[]T, T comparable](items S, elements ...T) S {
	result := make(S, 0, len(items))
	for _, el := range items {
		allowed := true
		for _, other := range elements {
			if el == other {
				allowed = false
			}
		}
		if allowed {
			result = append(result, el)
		}
	}
	return result
}

// Wrap wraps item in a slice of the same type.
func Wrap[T any](item T) []T {
	return []T{item}
}

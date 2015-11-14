package base

import "sort"

// Store mapping of string -> set of strings
type StringStringMapSet map[string]map[string]bool

func MakeStringStringMapSet() StringStringMapSet {
	return make(StringStringMapSet)
}

// Returns true if the value is new to the set; false if the
// value already existed.
func (m StringStringMapSet) Insert(key, value string) bool {
	set, exists := m[key]
	if !exists {
		set = make(map[string]bool)
		m[key] = set
	}
	set[value] = true
	return !exists
}

// Return the data type a map of strings to the array of values,
// in case-sensitive sorted order
func (m StringStringMapSet) ToMapSlice() map[string][]string {
	q := make(map[string][]string, len(m))
	for key, set := range m {
		arr := make([]string, 0, len(set))
		for value, _ := range set {
			arr = append(arr, value)
		}
		sort.Strings(arr)
		q[key] = arr
	}
	return q
}

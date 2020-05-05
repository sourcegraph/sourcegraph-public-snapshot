package datastructures

import "sort"

// TODO(efritz) - these need to be IDs not strings

// IDSet is a set of string identifiers.
type IDSet map[string]struct{}

// Add adds the given element to the set.
func (set IDSet) Add(id string) {
	set[id] = struct{}{}
}

// AddAll adds the contents of the given set to this set.
func (set IDSet) AddAll(other IDSet) {
	for k := range other {
		set.Add(k)
	}
}

// Contains determines if the given element is a member of this set.
func (set IDSet) Contains(id string) bool {
	_, ok := set[id]
	return ok
}

// Keys returns the sorted contents of this set.
func (set IDSet) Keys() []string {
	var keys []string
	for k := range set {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	return keys
}

// Choose deterministically returns a key from this set. If the set is
// empty, a false-valued flag is returned.
func (set IDSet) Choose() (string, bool) {
	if len(set) == 0 {
		return "", false
	}

	return set.Keys()[0], true
}

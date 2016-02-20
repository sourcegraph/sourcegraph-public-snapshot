package base

import (
	"sort"
)

// Sorted int vector, in ascending order.  Duplicates are allowed.
type SortedIntVector []int

func (s SortedIntVector) Len() int           { return len(s) }
func (s SortedIntVector) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s SortedIntVector) Less(i, j int) bool { return s[i] < s[j] }

func (s SortedIntVector) Add(x int) SortedIntVector {
	// Fast path for sorted insert
	if len(s) == 0 || s[len(s)-1] < x {
		return append(s, x)
	}

	i := sort.Search(len(s), func(i int) bool { return s[i] >= x })
	s = append(s, x)
	copy(s[i+1:], s[i:])
	s[i] = x

	return s
}

func (s SortedIntVector) Remove(x int) SortedIntVector {
	i := sort.Search(len(s), func(i int) bool { return s[i] >= x })
	if i < len(s) && s[i] == x {
		copy(s[i:], s[i+1:])
		s = s[:len(s)-1]
	}
	return s
}

func (s SortedIntVector) Contains(x int) bool {
	i := sort.Search(len(s), func(i int) bool { return s[i] >= x })
	return i < len(s) && s[i] == x
}

// Sorted int64 vector, in ascending order.  Duplicates are allowed.
type SortedInt64Vector []int64

func (s SortedInt64Vector) Len() int           { return len(s) }
func (s SortedInt64Vector) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s SortedInt64Vector) Less(i, j int) bool { return s[i] < s[j] }

func (s SortedInt64Vector) Add(x int64) SortedInt64Vector {
	// Fast path for sorted insert
	if len(s) == 0 || s[len(s)-1] < x {
		return append(s, x)
	}

	i := sort.Search(len(s), func(i int) bool { return s[i] >= x })
	s = append(s, x)
	copy(s[i+1:], s[i:])
	s[i] = x

	return s
}

func (s SortedInt64Vector) Remove(x int64) SortedInt64Vector {
	i := sort.Search(len(s), func(i int) bool { return s[i] >= x })
	if i < len(s) && s[i] == x {
		copy(s[i:], s[i+1:])
		s = s[:len(s)-1]
	}
	return s
}

func (s SortedInt64Vector) Contains(x int64) bool {
	i := sort.Search(len(s), func(i int) bool { return s[i] >= x })
	return i < len(s) && s[i] == x
}

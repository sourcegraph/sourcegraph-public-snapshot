package collections

import orderedmap "github.com/wk8/go-ordered-map/v2"

// OrderedSet keeps track of values in insertion order.
type OrderedSet[T comparable] orderedmap.OrderedMap[T, struct{}]

// NewOrderedSet creates a OrderedSet[T] with the given values.
// T must be a comparable type (implementing sort.Interface or == operator).
func NewOrderedSet[T comparable](values ...T) *OrderedSet[T] {
	s := OrderedSet[T](*orderedmap.New[T, struct{}]())
	s.Add(values...)
	return &s
}

func (s *OrderedSet[T]) impl() *orderedmap.OrderedMap[T, struct{}] {
	return (*orderedmap.OrderedMap[T, struct{}])(s)
}

func (s *OrderedSet[T]) Add(values ...T) {
	for _, v := range values {
		s.impl().Set(v, struct{}{})
	}
}

func (s *OrderedSet[T]) Remove(values ...T) {
	for _, v := range values {
		s.impl().Delete(v)
	}
}

func (s *OrderedSet[T]) Has(value T) bool {
	_, found := s.impl().Get(value)
	return found
}

// Values returns a slice with all the values in the set.
// The values are returned in insertion order.
func (s *OrderedSet[T]) Values() []T {
	out := make([]T, 0, s.impl().Len())
	for x := s.impl().Oldest(); x != nil; x = x.Next() {
		out = append(out, x.Key)
	}
	return out
}

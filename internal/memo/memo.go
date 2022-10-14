package memo

import "sync"

// MemoizedConstructor wraps a function returning a value and an error,
// memoizing its result. Multiple calls to Init will result in the
// underlying constructor being called once. All callers will receive
// the same return values.
type MemoizedConstructor[T any] struct {
	ctor  func() (T, error)
	value T
	err   error
	once  sync.Once
}

// NewMemoizedConstructor memoizes the given constructor
func NewMemoizedConstructor[T any](ctor func() (T, error)) *MemoizedConstructor[T] {
	return &MemoizedConstructor[T]{ctor: ctor}
}

// Init ensures that the given constructor has been called exactly
// once, then returns the constructor's result value and error.
func (m *MemoizedConstructor[T]) Init() (T, error) {
	m.once.Do(func() { m.value, m.err = m.ctor() })
	return m.value, m.err
}

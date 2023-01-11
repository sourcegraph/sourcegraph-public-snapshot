package memo

import "sync"

// MemoizedConstructorWithArg wraps a function returning taking a
// single argument value and returning a value and an error, memoizing
// its result. Multiple calls to Init will result in the underlying
// constructor being called once. The arguments to the call will be the
// first call to occur. All callers will receive the same return values.
type MemoizedConstructorWithArg[A, T any] struct {
	ctor  func(A) (T, error)
	value T
	err   error
	once  sync.Once
}

// NewMemoizedConstructor memoizes the given constructor
func NewMemoizedConstructorWithArg[A, T any](ctor func(A) (T, error)) *MemoizedConstructorWithArg[A, T] {
	return &MemoizedConstructorWithArg[A, T]{ctor: ctor}
}

// Init ensures that the given constructor has been called exactly
// once, then returns the constructor's result value and error.
func (m *MemoizedConstructorWithArg[A, T]) Init(arg A) (T, error) {
	m.once.Do(func() { m.value, m.err = m.ctor(arg) })
	return m.value, m.err
}

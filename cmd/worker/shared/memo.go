package shared

import "sync"

// MemoizedConstructor wraps a function returning a value and an error
// and memoizes its result. Multiple calls to Init will result in the
// underlying constructor being called once. All callers will receive
// the same return values.
type MemoizedConstructor struct {
	ctor  func() (interface{}, error)
	value interface{}
	err   error
	once  sync.Once
}

// NewMemoizedConstructor memoizes the given constructor
func NewMemoizedConstructor(ctor func() (interface{}, error)) *MemoizedConstructor {
	return &MemoizedConstructor{ctor: ctor}
}

// Init ensures that the given constructor has been called exactly
// once, then returns the constructor's result value and error.
func (m *MemoizedConstructor) Init() (interface{}, error) {
	m.once.Do(func() { m.value, m.err = m.ctor() })
	return m.value, m.err
}

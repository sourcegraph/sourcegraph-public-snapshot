package atomicvalue

import "sync"

// Value manages an atomic value.
type Value struct {
	mu    sync.RWMutex
	value interface{}
}

// Get returns the current value.
func (v *Value) Get() interface{} {
	v.mu.RLock()
	cpy := v.value
	v.mu.RUnlock()
	return cpy
}

// Set changes the value to the result of f. The mutex is held for the entire
// duration of the function call.
func (v *Value) Set(f func() interface{}) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.value = f()
}

// New returns a new Value.
func New() *Value {
	return &Value{}
}

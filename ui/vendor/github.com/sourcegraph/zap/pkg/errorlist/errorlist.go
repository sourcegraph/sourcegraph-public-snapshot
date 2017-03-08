// package errorlist contains a concurrency safe error list helper.
package errorlist

import (
	"fmt"
	"sync"
)

// Errors is a concurrency safe error list helper. Note: Do not pass around by
// value, since it contains a mutex.
type Errors struct {
	mu     sync.Mutex
	errors []error
}

// Add adds err to the list of errors. It is safe to call it from concurrent
// goroutines.
func (e *Errors) Add(err error) {
	e.mu.Lock()
	e.errors = append(e.errors, err)
	e.mu.Unlock()
}

// Error returns the combined errors. If there are no errors nil is returned.
func (e *Errors) Error() error {
	e.mu.Lock()
	errs := e.errors
	e.mu.Unlock()
	switch len(errs) {
	case 0:
		return nil
	case 1:
		return errs[0]
	default:
		return fmt.Errorf("%s [and %d more errors]", errs[0], len(errs)-1)
	}
}

package buildserver

import (
	"fmt"
	"sync"
)

type errorList struct {
	mu     sync.Mutex
	errors []error
}

// add adds err to the list of errors. It is safe to call it from
// concurrent goroutines.
func (e *errorList) add(err error) {
	e.mu.Lock()
	e.errors = append(e.errors, err)
	e.mu.Unlock()
}

// errors returns the list of errors as a single error. It is NOT safe
// to call from concurrent goroutines.
func (e *errorList) error() error {
	switch len(e.errors) {
	case 0:
		return nil
	case 1:
		return e.errors[0]
	default:
		return fmt.Errorf("%s [and %d more errors]", e.errors[0], len(e.errors)-1)
	}
}

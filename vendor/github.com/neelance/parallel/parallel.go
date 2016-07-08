// Package parallel can be used for easily running goroutines with a limit on concurrency. It also
// helps with collecting errors.
//
// It is intended as a replacement for github.com/rogpeppe/rog-go/parallel and avoids messing up CPU
// and memory profiles by not adding its own stack frame.
package parallel

import (
	"fmt"
	"sync"
)

// Run represents a number of functions running concurrently.
type Run struct {
	sem chan struct{}
	wg  sync.WaitGroup

	errsMutex sync.Mutex
	errs      []error
}

// NewRun returns a new parallel instance. It will allow up to maxPar locks concurrently.
func NewRun(maxPar int) *Run {
	return &Run{
		sem: make(chan struct{}, maxPar),
	}
}

// Acquire acquires a lock. Up to maxPar (see NewRun) locks can be held concurrently. If no lock is
// available, Acquire blocks until another goroutine calls Release.
//
// Caution: Call this before starting a new goroutine, not inside of it (race condition with Wait).
func (r *Run) Acquire() {
	r.sem <- struct{}{}
	r.wg.Add(1)
}

// Release releases a lock and allows another lock to be aquired by Acquire.
func (r *Run) Release() {
	<-r.sem
	r.wg.Done()
}

// Error stores an error to be returned by Wait.
func (r *Run) Error(err error) {
	r.errsMutex.Lock()
	r.errs = append(r.errs, err)
	r.errsMutex.Unlock()
}

// Wait waits for all locks to be released. If any errors were encountered, it returns an
// Errors value describing all the errors in arbitrary order.
func (r *Run) Wait() error {
	r.wg.Wait()
	if len(r.errs) != 0 {
		return Errors(r.errs)
	}
	return nil
}

// Errors holds any errors encountered during the parallel run.
type Errors []error

func (errs Errors) Error() string {
	switch len(errs) {
	case 0:
		return "no error"
	case 1:
		return errs[0].Error()
	default:
		return fmt.Sprintf("%s (and %d more)", errs[0].Error(), len(errs)-1)
	}
}

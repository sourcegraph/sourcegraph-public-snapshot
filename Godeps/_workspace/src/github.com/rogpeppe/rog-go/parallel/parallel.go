// The parallel package provides a way of running functions
// concurrently while limiting the maximum number
// running at once.
package parallel

import (
	"fmt"
	"sync"
)

// Run represents a number of functions running concurrently.
type Run struct {
	limiter chan struct{}
	done    chan error
	err     chan error
	wg      sync.WaitGroup
}

// Errors holds any errors encountered during
// the parallel run.
type Errors []error

func (errs Errors) Error() string {
	switch len(errs) {
	case 0:
		return "no error"
	case 1:
		return errs[0].Error()
	}
	return fmt.Sprintf("%s (and %d more)", errs[0].Error(), len(errs)-1)
}

// NewRun returns a new parallel instance.  It will run up to maxPar
// functions concurrently.
func NewRun(maxPar int) *Run {
	r := &Run{
		limiter: make(chan struct{}, maxPar),
		done:    make(chan error),
		err:     make(chan error),
	}
	go func() {
		var errs Errors
		for e := range r.done {
			errs = append(errs, e)
		}
		// TODO sort errors by original order of Do request?
		if len(errs) > 0 {
			r.err <- errs
		} else {
			r.err <- nil
		}
	}()
	return r
}

// Do requests that r run f concurrently.  If there are already the maximum
// number of functions running concurrently, it will block until one of
// them has completed. Do may itself be called concurrently.
func (r *Run) Do(f func() error) {
	r.limiter <- struct{}{}
	r.wg.Add(1)
	go func() {
		defer func() {
			r.wg.Done()
			<-r.limiter
		}()
		if err := f(); err != nil {
			r.done <- err
		}
	}()
}

// Wait marks the parallel instance as complete and waits for all the
// functions to complete.  If any errors were encountered, it returns an
// Errors value describing all the errors in arbitrary order.
func (r *Run) Wait() error {
	r.wg.Wait()
	close(r.done)
	return <-r.err
}

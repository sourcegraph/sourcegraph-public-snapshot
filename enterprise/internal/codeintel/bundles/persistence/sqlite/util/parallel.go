package util

import (
	"sync"

	"github.com/hashicorp/go-multierror"
)

// InvokeAll invokes each of the given functions in a different goroutine and blocks
// until all goroutines have finished. The return value is the multierror composed of
// error values from each invocation.
func InvokeAll(fns ...func() error) (err error) {
	var wg sync.WaitGroup
	errs := make(chan error, len(fns))

	for _, fn := range fns {
		wg.Add(1)

		go func(fn func() error) {
			defer wg.Done()

			if err := fn(); err != nil {
				errs <- err
			}
		}(fn)
	}

	wg.Wait()
	close(errs)

	for e := range errs {
		err = multierror.Append(err, e)
	}
	return err
}

// InvokeN invokes n copies of the given function in different goroutines. See InvokeAll
// for additional notes on semantics.
func InvokeN(n int, f func() error) error {
	fns := make([]func() error, n)
	for i := 0; i < n; i++ {
		fns[i] = f
	}

	return InvokeAll(fns...)
}

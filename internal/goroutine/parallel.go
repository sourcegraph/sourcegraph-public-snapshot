package goroutine

import (
	"sync"

	"github.com/hashicorp/go-multierror"
)

// Parallel calls each of the given functions in a goroutine. This method
// blocks until all goroutines have unblocked. The errors from each of the
// function invocations will be combined into a single error.
func Parallel(fns ...func() error) (err error) {
	var wg sync.WaitGroup
	wg.Add(len(fns))
	errs := make(chan error, len(fns))

	for _, f := range fns {
		go func(f func() error) {
			defer wg.Done()
			errs <- f()
		}(f)
	}

	wg.Wait()
	close(errs)

	for retValue := range errs {
		if retValue == nil {
			continue
		}

		if err == nil {
			err = retValue
		} else {
			err = multierror.Append(err, retValue)
		}
	}

	return err
}

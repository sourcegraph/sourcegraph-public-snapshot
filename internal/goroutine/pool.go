package goroutine

import (
	"runtime"
	"sync"

	"github.com/hashicorp/go-multierror"
)

// Pool worker is a function invoked by RunWorkers that sends
// any errors that occur during execution down a shared channel.
type PoolWorker func(errs chan<- error)

// SimplePoolWorker converts a function returning a single error
// value into a PoolWorker.
func SimplePoolWorker(fn func() error) PoolWorker {
	return func(errs chan<- error) {
		if err := fn(); err != nil {
			errs <- err
		}
	}
}

// RunWorkers invokes the given worker a number of times proportional
// to the maximum number of CPUs that can be executing simultaneously.
func RunWorkers(worker PoolWorker) error {
	return RunWorkersN(runtime.GOMAXPROCS(0), worker)
}

// RunWorkersN invokes the given worker n times and collects the
// errors from each invocation.
func RunWorkersN(n int, worker PoolWorker) (err error) {
	errs := make(chan error, n)

	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() { worker(errs); wg.Done() }()
	}

	go func() {
		wg.Wait()
		close(errs)
	}()

	for e := range errs {
		if err == nil {
			err = e
		} else {
			err = multierror.Append(err, e)
		}
	}

	return err
}

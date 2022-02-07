package goroutine

import (
	"runtime"
	"sync"

	"github.com/sourcegraph/sourcegraph/lib/errors"
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
			err = errors.Append(err, e)
		}
	}

	return err
}

// RunWorkersOverStrings invokes the given worker once for each of the
// given string values. The worker function will receive the index as well
// as the string value as parameters. Workers will be invoked in a number
// of concurrent routines proportional to the maximum number of CPUs that
// can be executing simultaneously.
func RunWorkersOverStrings(values []string, worker func(index int, value string) error) error {
	return RunWorkersOverStringsN(runtime.GOMAXPROCS(0), values, worker)
}

// RunWorkersOverStrings invokes the given worker once for each of the
// given string values. The worker function will receive the index as well
// as the string value as parameters. Workers will be invoked in n concurrent
// routines.
func RunWorkersOverStringsN(n int, values []string, worker func(index int, value string) error) error {
	return RunWorkersN(n, indexedStringWorker(loadIndexedStringChannel(values), worker))
}

type indexedString struct {
	Index int
	Value string
}

func loadIndexedStringChannel(values []string) <-chan indexedString {
	ch := make(chan indexedString, len(values))
	defer close(ch)

	for i, value := range values {
		ch <- indexedString{Index: i, Value: value}
	}

	return ch
}

func indexedStringWorker(ch <-chan indexedString, worker func(index int, value string) error) PoolWorker {
	return func(errs chan<- error) {
		for value := range ch {
			if err := worker(value.Index, value.Value); err != nil {
				errs <- err
			}
		}
	}
}

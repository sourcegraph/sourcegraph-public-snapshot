package uploadstore

import (
	"runtime"
	"sync"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// poolWorker is a function invoked by RunWorkers that sends
// any errors that occur during execution down a shared channel.
type poolWorker func(errs chan<- error)

// runWorkersN invokes the given worker n times and collects the
// errors from each invocation.
func runWorkersN(n int, worker poolWorker) (err error) {
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
	return runWorkersOverStringsN(runtime.GOMAXPROCS(0), values, worker)
}

// RunWorkersOverStrings invokes the given worker once for each of the
// given string values. The worker function will receive the index as well
// as the string value as parameters. Workers will be invoked in n concurrent
// routines.
func runWorkersOverStringsN(n int, values []string, worker func(index int, value string) error) error {
	return runWorkersN(n, indexedStringWorker(loadIndexedStringChannel(values), worker))
}

type indexedString struct {
	index int
	value string
}

func loadIndexedStringChannel(values []string) <-chan indexedString {
	ch := make(chan indexedString, len(values))
	defer close(ch)

	for i, value := range values {
		ch <- indexedString{index: i, value: value}
	}

	return ch
}

func indexedStringWorker(ch <-chan indexedString, worker func(index int, value string) error) poolWorker {
	return func(errs chan<- error) {
		for value := range ch {
			if err := worker(value.index, value.value); err != nil {
				errs <- err
			}
		}
	}
}

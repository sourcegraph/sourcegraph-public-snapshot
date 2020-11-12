package uploadstore

import (
	"runtime"
	"sync"

	"github.com/hashicorp/go-multierror"
)

type indexedString struct {
	Index int
	Value string
}

func invokeParallel(values []string, f func(index int, value string) error) (err error) {
	ch := make(chan indexedString, len(values))
	for i, value := range values {
		ch <- indexedString{Index: i, Value: value}
	}
	close(ch)

	var wg sync.WaitGroup
	errs := make(chan error, len(values))

	for i := 0; i < runtime.GOMAXPROCS(0); i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for value := range ch {
				if err := f(value.Index, value.Value); err != nil {
					errs <- err
				}
			}
		}()
	}

	wg.Wait()
	close(errs)

	for e := range errs {
		if err == nil {
			err = e
		} else {
			err = multierror.Append(err, e)
		}
	}

	return err
}

package uploadstore

import "github.com/sourcegraph/sourcegraph/internal/goroutine"

// TODO - move to goroutine package

func invokeParallel(values []string, f func(index int, value string) error) error {
	ch := loadIndexedStringChannel(values)

	return goroutine.RunWorkers(func(errs chan<- error) {
		for value := range ch {
			if err := f(value.Index, value.Value); err != nil {
				errs <- err
			}
		}
	})
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

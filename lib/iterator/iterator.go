package iterator

import "fmt"

// New returns an Iterator for next.
//
// next is a function which is repeatedly called until no items are returned
// or there is a non-nil error. These items are returned one by one via Next
// and Current.
func New[T any](next func() ([]T, error)) *Iterator[T] {
	return &Iterator[T]{next: next}
}

// Iterator provides a convenient interface for iterating over items which are
// fetched in batches and can error. In particular this is designed for
// pagination.
//
// Iterating stops as soon as the underlying next function returns no items.
// If an error is returned then next won't be called again and Err will return
// a non-nil error.
type Iterator[T any] struct {
	items []T
	err   error
	done  bool

	next func() ([]T, error)
}

// Next advances the iterator to the next item, which will then be available
// from Current. It returns false when the iterator stops, either due to the
// end of the input or an error occurred. After Next returns false Err() will
// return the error occurred or nil if none.
func (it *Iterator[T]) Next() bool {
	if len(it.items) > 1 {
		it.items = it.items[1:]
		return true
	}

	// done is true if we shouldn't call it.next again.
	if it.done {
		it.items = nil // "consume" the last item when err != nil
		return false
	}

	it.items, it.err = it.next()
	if len(it.items) == 0 || it.err != nil {
		it.done = true
	}

	return len(it.items) > 0
}

// Current returns the latest item advanced by Next. Note: this will panic if
// Next returned false or if Next was never called.
func (it *Iterator[T]) Current() T {
	if len(it.items) == 0 {
		if it.done {
			panic(fmt.Sprintf("%T.Current() called after Next() returned false", it))
		} else {
			panic(fmt.Sprintf("%T.Current() called before first call to Next()", it))
		}
	}
	return it.items[0]
}

// Err returns the first non-nil error encountered by Next.
func (it *Iterator[T]) Err() error {
	return it.err
}

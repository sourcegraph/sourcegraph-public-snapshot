package iterator

func New[T any](next func() ([]T, error)) *Iterator[T] {
	return &Iterator[T]{next: next}
}

type Iterator[T any] struct {
	items []T
	err   error
	done  bool

	// next is a function which is repeatedly called until no items are
	// returned or there is a non-nil error. These items are returned one by
	// one via Next and Current.
	next func() ([]T, error)
}

func (it *Iterator[T]) Next() bool {
	if it.done {
		return false
	}

	if len(it.items) > 1 {
		it.items = it.items[1:]
		return true
	}

	it.items, it.err = it.next()
	if len(it.items) == 0 || it.err != nil {
		it.done = true
	}

	return !it.done
}

func (it *Iterator[T]) Current() T {
	return it.items[0]
}

func (it *Iterator[T]) Err() error {
	return it.err
}

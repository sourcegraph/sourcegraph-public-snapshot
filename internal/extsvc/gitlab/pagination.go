package gitlab

import (
	"encoding/json"
	"sync"
	"sync/atomic"
)

type PaginatedResult[T any] struct {
	mu       sync.Mutex
	cache    []T
	current  int
	complete atomic.Bool

	nextPage func() ([]T, error)
}

var _ json.Marshaler = &PaginatedResult[struct{}]{}

func newPaginatedResult[T any](nextPage func() ([]T, error)) (*PaginatedResult[T], error) {
	pr := &PaginatedResult[T]{
		nextPage: nextPage,
	}

	if err := pr.getNextPage(); err != nil {
		return nil, err
	}
	return pr, nil
}

func (pr *PaginatedResult[T]) All() ([]T, error) {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	if err := pr.fullyHydrate(); err != nil {
		return nil, err
	}

	return pr.cache, nil
}

func (pr *PaginatedResult[T]) Next() (t T, ok bool, err error) {
	if pr.complete.Load() {
		return
	}

	pr.mu.Lock()
	defer pr.mu.Unlock()

	if pr.current >= len(pr.cache) {
		if err = pr.getNextPage(); err != nil {
			return
		}
	}

	if pr.current >= len(pr.cache) {
		return
	}

	t = pr.cache[pr.current]
	ok = true
	pr.current += 1

	return
}

// TODO: keep?
func (pr *PaginatedResult[T]) Reset() {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	pr.current = 0
}

func (pr *PaginatedResult[T]) MarshalJSON() ([]byte, error) {
	all, err := pr.All()
	if err != nil {
		return nil, err
	}

	return json.Marshal(all)
}

func (pr *PaginatedResult[T]) mustAll() []T {
	all, err := pr.All()
	if err != nil {
		panic(err)
	}

	return all
}

func (pr *PaginatedResult[T]) fullyHydrate() error {
	for {
		if pr.complete.Load() {
			return nil
		}

		if err := pr.getNextPage(); err != nil {
			return err
		}
	}
}

func (pr *PaginatedResult[T]) getNextPage() error {
	if pr.complete.Load() {
		return nil
	}

	page, err := pr.nextPage()
	if err != nil {
		return err
	}

	if len(page) == 0 {
		pr.complete.Store(true)
		return nil
	}

	pr.cache = append(pr.cache, page...)
	return nil
}

func NewMockPaginatedResult[T any](results []T, err error) (*PaginatedResult[T], error) {
	i := 0
	return newPaginatedResult(func() ([]T, error) {
		if i >= len(results) {
			return nil, err
		}

		page := []T{results[i]}
		i += 1

		return page, err
	})
}

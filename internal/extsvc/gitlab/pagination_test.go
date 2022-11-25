package gitlab

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestPaginatedResult(t *testing.T) {
	t.Run("empty result set", func(t *testing.T) {
		pr, err := newPaginatedResult(func() ([]struct{}, error) {
			return nil, nil
		})
		assert.NoError(t, err)
		assert.NotNil(t, pr)
		assert.True(t, pr.complete.Load())

		have, ok, err := pr.Next()
		assert.NoError(t, err)
		assert.Zero(t, have)
		assert.False(t, ok)
	})

	t.Run("immediate error", func(t *testing.T) {
		want := errors.New("foo")
		pr, err := newPaginatedResult(func() ([]struct{}, error) {
			return nil, want
		})
		assert.ErrorIs(t, err, want)
		assert.Nil(t, pr)
	})

	t.Run("belated error", func(t *testing.T) {
		want := errors.New("foo")
		page := 0
		pr, err := newPaginatedResult(func() ([]int, error) {
			if page > 1 {
				return nil, want
			}

			page += 1
			return []int{1, 2, 3}, nil
		})
		assert.NoError(t, err)
		assert.NotNil(t, pr)

		success := 0
		for {
			var ok bool
			_, ok, err = pr.Next()
			if err != nil {
				assert.False(t, ok)
				break
			}
			assert.True(t, ok)
			success += 1
		}

		assert.ErrorIs(t, err, want)
		assert.EqualValues(t, 6, success)
	})

	t.Run("success", func(t *testing.T) {
		for _, concurrency := range []int{1, 10, 100, 1000} {
			t.Run(fmt.Sprintf("concurrency %d", concurrency), func(t *testing.T) {
				ns := numberSequence{chunkSize: 100, pages: 100}
				pr, err := newPaginatedResult(func() ([]int, error) {
					return ns.nextPage()
				})
				assert.NoError(t, err)
				assert.NotNil(t, pr)

				if concurrency > 1 {
					assertParallel(t, pr, ns.expected(), concurrency)
				} else {
					// We want to defer back to assertSync here because it can
					// also check the order results are returned in.
					assertSync(t, pr, ns.expected())
				}
			})
		}
	})
}

type numberSequence struct {
	chunkSize   int
	currentPage int
	pages       int
}

func (ns *numberSequence) expected() []int {
	all := make([]int, ns.chunkSize*ns.pages)
	for i := 0; i < ns.chunkSize*ns.pages; i++ {
		all[i] = i
	}

	return all
}

func (ns *numberSequence) nextPage() ([]int, error) {
	if ns.currentPage >= ns.pages {
		return nil, nil
	}

	page := make([]int, ns.chunkSize)
	for i := 0; i < ns.chunkSize; i++ {
		page[i] = i + ns.chunkSize*ns.currentPage
	}

	ns.currentPage += 1
	return page, nil
}

func assertParallel[T any](t *testing.T, pr *PaginatedResult[T], want []T, concurrency int) {
	t.Helper()

	var (
		seen   []T
		recvWg sync.WaitGroup
		sendWg sync.WaitGroup
	)

	c := make(chan T, concurrency)
	recvWg.Add(1)
	go func() {
		defer recvWg.Done()
		for value := range c {
			seen = append(seen, value)
		}
	}()

	for i := 0; i < concurrency; i++ {
		sendWg.Add(1)
		go func() {
			defer sendWg.Done()

			for {
				next, ok, err := pr.Next()
				assert.NoError(t, err)

				if !ok {
					return
				}
				c <- next
			}
		}()
	}

	sendWg.Wait()
	close(c)
	recvWg.Wait()

	// We can't guarantee that the order results were added to seen is the same
	// as the order they were received in, due to the vagaries of parallel
	// execution, so we'll just ensure we at least visited everything once.
	assert.ElementsMatch(t, seen, want)
}

func assertSync[T any](t *testing.T, pr *PaginatedResult[T], want []T) {
	t.Helper()

	seen := make([]T, 0, len(want))
	for next, ok, err := pr.Next(); ok; next, ok, err = pr.Next() {
		assert.NoError(t, err)
		seen = append(seen, next)
	}

	assert.Equal(t, seen, want)
}

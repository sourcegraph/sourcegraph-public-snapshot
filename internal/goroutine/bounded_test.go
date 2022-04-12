package goroutine

import (
	"sync"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestBounded(t *testing.T) {
	var (
		mu      sync.Mutex
		total   int
		current int
		errS    string
	)

	size := 4
	b := NewBounded(size)
	for i := 0; i < 20; i++ {
		b.Go(func() error {
			mu.Lock()
			current++
			total++
			if current > size {
				errS = "too many running"
			}
			mu.Unlock()

			time.Sleep(time.Millisecond)

			mu.Lock()
			current--
			mu.Unlock()
			return nil
		})
	}

	if err := b.Wait(); err != nil {
		t.Fatal("Wait:", err)
	}

	if errS != "" {
		t.Fatal(errS)
	}
}

func TestBounded_error(t *testing.T) {
	boom := errors.New("boom")
	b := NewBounded(4)
	for i := 0; i < 20; i++ {
		if i%5 == 0 {
			b.Go(func() error { return boom })
		} else {
			b.Go(func() error { return nil })
		}
	}

	err := b.Wait()
	if err != boom {
		t.Fatal("unexpected error", err)
	}
}

package watchdog

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/derision-test/glock"
)

func TestWatchDog(t *testing.T) {
	ticker := glock.NewMockTicker(5 * time.Minute)
	var count uint32
	expectedCount := 100
	var wg sync.WaitGroup

	mockCallback := func() {
		atomic.AddUint32(&count, 1)
		wg.Done()
	}

	w := &WatchDog{
		ticker:   ticker,
		callback: mockCallback,
		done:     make(chan struct{}, 1),
	}

	go w.Start()
	for i := 0; i < expectedCount; i++ {
		wg.Add(1)
		ticker.BlockingAdvance(5 * time.Minute)
	}
	wg.Wait()
	w.Stop()

	if count != uint32(expectedCount) {
		t.Errorf("expected mock callback to be called %d times, got %d", expectedCount, count)
	}
}

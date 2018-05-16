package openidconnect

import (
	"sync"
	"testing"
	"time"
)

func TestProviderCache(t *testing.T) {
	calls := 0
	mockNewProvider = func(string) (*provider, error) {
		calls++
		return nil, nil
	}
	defer func() { mockNewProvider = nil }()

	defer func(d time.Duration) { cacheTTLOK = d }(cacheTTLOK)
	cacheTTLOK = 250 * time.Millisecond

	cache.get("")
	deadline := time.Now().Add(cacheTTLOK + 25*time.Millisecond) // slightly more than 1 TTL
	var wg sync.WaitGroup
	for time.Now().Before(deadline) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cache.get("")
		}()
		time.Sleep(1 * time.Millisecond)
	}
	wg.Wait()
	if want := 2; calls != want {
		t.Errorf("got %d calls, want %d", calls, want)
	}
}

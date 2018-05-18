package saml

import (
	"sync"
	"testing"
	"time"

	"github.com/crewjam/saml/samlsp"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestProviderCache1(t *testing.T) {
	calls := 0
	mockGetServiceProvider1 = func(*schema.SAMLAuthProvider) (*samlsp.Middleware, error) {
		calls++
		return nil, nil
	}
	defer func() { mockGetServiceProvider1 = nil }()

	defer func(d time.Duration) { cache1TTLOK = d }(cache1TTLOK)
	cache1TTLOK = 250 * time.Millisecond

	cache1.get(schema.SAMLAuthProvider{})
	deadline := time.Now().Add(cache1TTLOK + 25*time.Millisecond) // slightly more than 1 TTL
	var wg sync.WaitGroup
	for time.Now().Before(deadline) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cache1.get(schema.SAMLAuthProvider{})
		}()
		time.Sleep(1 * time.Millisecond)
	}
	wg.Wait()
	if want := 2; calls != want {
		t.Errorf("got %d calls, want %d", calls, want)
	}
}

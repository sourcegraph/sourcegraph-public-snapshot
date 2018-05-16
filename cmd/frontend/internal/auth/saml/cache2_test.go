package saml

import (
	"sync"
	"testing"
	"time"

	saml2 "github.com/russellhaering/gosaml2"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestProviderCache2(t *testing.T) {
	calls := 0
	mockGetServiceProvider2 = func(*schema.SAMLAuthProvider) (*saml2.SAMLServiceProvider, error) {
		calls++
		return nil, nil
	}
	defer func() { mockGetServiceProvider2 = nil }()

	defer func(d time.Duration) { cache2TTLOK = d }(cache2TTLOK)
	cache2TTLOK = 250 * time.Millisecond

	cache2.get(schema.SAMLAuthProvider{})
	deadline := time.Now().Add(cache2TTLOK + 25*time.Millisecond) // slightly more than 1 TTL
	var wg sync.WaitGroup
	for time.Now().Before(deadline) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cache2.get(schema.SAMLAuthProvider{})
		}()
		time.Sleep(1 * time.Millisecond)
	}
	wg.Wait()
	if want := 2; calls != want {
		t.Errorf("got %d calls, want %d", calls, want)
	}
}

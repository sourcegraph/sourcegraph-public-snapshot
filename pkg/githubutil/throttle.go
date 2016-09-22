package githubutil

import (
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

// throttledTransport ensures there is only one request made to the GitHub API
// at a time per client. This helps ensure we are in accordance with the GitHub
// API best practices.
//
// See https://developer.github.com/guides/best-practices-for-integrators/#dealing-with-rate-limits
type throttledTransport struct {
	Transport http.RoundTripper
	Throttle  sync.Locker // A lock that limits concurrent requests (if nil, requests will not be throttled)
}

func (t *throttledTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	if t.Throttle == nil {
		return t.Transport.RoundTrip(r)
	}
	t.Throttle.Lock()
	defer t.Throttle.Unlock()
	return t.Transport.RoundTrip(r)
}

// gaugedMutex is a mutex that reports the amount of outstanding lock acquirers
// to prometheus.
type gaugedMutex struct {
	lock    sync.Mutex
	counter prometheus.Gauge
}

func newGaugedMutex(g prometheus.Gauge) *gaugedMutex {
	return &gaugedMutex{
		lock:    sync.Mutex{},
		counter: g,
	}
}

func (g *gaugedMutex) Lock() {
	g.counter.Inc()
	g.lock.Lock()
}

func (g *gaugedMutex) Unlock() {
	g.lock.Unlock()
	g.counter.Dec()
}

package backend

import (
	"context"
	"time"

	"github.com/google/zoekt"
	"github.com/google/zoekt/query"
	"github.com/prometheus/client_golang/prometheus"
)

var requestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Namespace: "src",
	Subsystem: "zoekt",
	Name:      "request_duration_seconds",
	Help:      "Time (in seconds) spent on request.",
	Buckets:   prometheus.DefBuckets,
}, []string{"category", "code"})

func init() {
	prometheus.MustRegister(requestDuration)
}

type meteredSearcher struct {
	zoekt.Searcher
}

func NewMeteredSearcher(z zoekt.Searcher) zoekt.Searcher {
	return &meteredSearcher{z}
}

func (m *meteredSearcher) Search(ctx context.Context, q query.Q, opts *zoekt.SearchOptions) (*zoekt.SearchResult, error) {
	start := time.Now()
	zsr, err := m.Searcher.Search(ctx, q, opts)
	d := time.Since(start)

	code := "200"
	if err != nil {
		code = "error"
	}

	// TODO(uwedeportivo): host label for horizontally scaled zoekt case
	requestDuration.WithLabelValues("Search", code).Observe(d.Seconds())
	return zsr, err
}

func (m *meteredSearcher) List(ctx context.Context, q query.Q) (*zoekt.RepoList, error) {
	start := time.Now()
	zrl, err := m.Searcher.List(ctx, q)
	d := time.Since(start)

	code := "200"
	if err != nil {
		code = "error"
	}

	requestDuration.WithLabelValues("List", code).Observe(d.Seconds())
	return zrl, err
}

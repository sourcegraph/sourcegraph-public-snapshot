package backend

import (
	"context"
	"time"

	"github.com/google/zoekt"
	"github.com/google/zoekt/query"
	"github.com/opentracing/opentracing-go/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

var requestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "src_zoekt_request_duration_seconds",
	Help:    "Time (in seconds) spent on request.",
	Buckets: prometheus.DefBuckets,
}, []string{"hostname", "category", "code"})

func init() {
	prometheus.MustRegister(requestDuration)
}

type meteredSearcher struct {
	zoekt.Searcher

	hostname string
}

func NewMeteredSearcher(hostname string, z zoekt.Searcher) zoekt.Searcher {
	return &meteredSearcher{
		Searcher: z,
		hostname: hostname,
	}
}

func (m *meteredSearcher) Search(ctx context.Context, q query.Q, opts *zoekt.SearchOptions) (*zoekt.SearchResult, error) {
	start := time.Now()

	cat := "SearchAll"
	if m.hostname != "" {
		cat = "Search"
	}

	tr, ctx := trace.New(ctx, "zoekt."+cat, queryString(q))
	tr.LogFields(
		log.String("hostname", m.hostname),
		log.Object("options", opts),
	)

	zsr, err := m.Searcher.Search(ctx, q, opts)

	code := "200"
	if err != nil {
		code = "error"
	}

	requestDuration.WithLabelValues(m.hostname, cat, code).Observe(time.Since(start).Seconds())

	tr.SetError(err)
	if zsr != nil {
		tr.LogFields(
			log.Int("filematches", len(zsr.Files)),
			log.Object("stats", &zsr.Stats),
		)
	}
	tr.Finish()

	return zsr, err
}

func (m *meteredSearcher) List(ctx context.Context, q query.Q) (*zoekt.RepoList, error) {
	start := time.Now()

	cat := "ListAll"
	if m.hostname != "" {
		cat = "List"
	}

	tr, ctx := trace.New(ctx, "zoekt."+cat, queryString(q))
	tr.LogFields(log.String("hostname", m.hostname))

	zsl, err := m.Searcher.List(ctx, q)

	code := "200"
	if err != nil {
		code = "error"
	}

	requestDuration.WithLabelValues(m.hostname, cat, code).Observe(time.Since(start).Seconds())

	tr.SetError(err)
	if zsl != nil {
		tr.LogFields(log.Int("repos", len(zsl.Repos)))
	}
	tr.Finish()

	return zsl, err
}

func queryString(q query.Q) string {
	if q == nil {
		return "<nil>"
	}
	return q.String()
}

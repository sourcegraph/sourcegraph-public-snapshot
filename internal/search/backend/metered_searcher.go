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
	tr.LogFields(log.String("hostname", m.hostname))
	if opts != nil {
		tr.LogFields(
			log.Bool("opts.EstimateDocCount", opts.EstimateDocCount),
			log.Bool("opts.Whole", opts.Whole),
			log.Int("opts.ShardMaxMatchCount", opts.ShardMaxMatchCount),
			log.Int("opts.TotalMaxMatchCount", opts.TotalMaxMatchCount),
			log.Int("opts.ShardMaxImportantMatch", opts.ShardMaxImportantMatch),
			log.Int("opts.TotalMaxImportantMatch", opts.TotalMaxImportantMatch),
			log.Int64("opts.MaxWallTimeMS", opts.MaxWallTime.Milliseconds()),
			log.Int("opts.MaxDocDisplayCount", opts.MaxDocDisplayCount),
		)
	}

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
			log.Int64("stats.ContentBytesLoaded", zsr.Stats.ContentBytesLoaded),
			log.Int64("stats.IndexBytesLoaded", zsr.Stats.IndexBytesLoaded),
			log.Int("stats.Crashes", zsr.Stats.Crashes),
			log.Int64("stats.DurationMS", zsr.Stats.Duration.Milliseconds()),
			log.Int("stats.FileCount", zsr.Stats.FileCount),
			log.Int("stats.ShardFilesConsidered", zsr.Stats.ShardFilesConsidered),
			log.Int("stats.FilesConsidered", zsr.Stats.FilesConsidered),
			log.Int("stats.FilesLoaded", zsr.Stats.FilesLoaded),
			log.Int("stats.FilesSkipped", zsr.Stats.FilesSkipped),
			log.Int("stats.ShardsSkipped", zsr.Stats.ShardsSkipped),
			log.Int("stats.MatchCount", zsr.Stats.MatchCount),
			log.Int("stats.NgramMatches", zsr.Stats.NgramMatches),
			log.Int64("stats.WaitMS", zsr.Stats.Wait.Milliseconds()),
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

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

	var cat string
	var tags []trace.Tag
	if m.hostname == "" {
		cat = "SearchAll"
	} else {
		cat = "Search"
		tags = []trace.Tag{
			{Key: "span.kind", Value: "client"},
			{Key: "peer.address", Value: m.hostname},
			{Key: "peer.service", Value: "zoekt"},
		}
	}

	tr, ctx := trace.New(ctx, "zoekt."+cat, queryString(q), tags...)
	if opts != nil {
		tr.LogFields(
			log.Bool("opts.estimate_doc_count", opts.EstimateDocCount),
			log.Bool("opts.whole", opts.Whole),
			log.Int("opts.shard_max_match_count", opts.ShardMaxMatchCount),
			log.Int("opts.total_max_match_count", opts.TotalMaxMatchCount),
			log.Int("opts.shard_max_important_match", opts.ShardMaxImportantMatch),
			log.Int("opts.total_max_important_match", opts.TotalMaxImportantMatch),
			log.Int64("opts.max_wall_time_ms", opts.MaxWallTime.Milliseconds()),
			log.Int("opts.max_doc_display_count", opts.MaxDocDisplayCount),
		)
	}

	zsr, err := m.Searcher.Search(ctx, q, opts)

	code := "200"
	if err != nil {
		code = "error"
	}

	d := time.Since(start)
	requestDuration.WithLabelValues(m.hostname, cat, code).Observe(d.Seconds())

	tr.SetError(err)
	if zsr != nil {
		tr.LogFields(
			log.Int("filematches", len(zsr.Files)),
			log.Int64("rpc.latency_ms", (d-zsr.Stats.Duration-zsr.Stats.Wait).Milliseconds()),
			log.Int64("stats.content_bytes_loaded", zsr.Stats.ContentBytesLoaded),
			log.Int64("stats.index_bytes_loaded", zsr.Stats.IndexBytesLoaded),
			log.Int("stats.crashes", zsr.Stats.Crashes),
			log.Int64("stats.duration_ms", zsr.Stats.Duration.Milliseconds()),
			log.Int("stats.file_count", zsr.Stats.FileCount),
			log.Int("stats.shard_files_considered", zsr.Stats.ShardFilesConsidered),
			log.Int("stats.files_considered", zsr.Stats.FilesConsidered),
			log.Int("stats.files_loaded", zsr.Stats.FilesLoaded),
			log.Int("stats.files_skipped", zsr.Stats.FilesSkipped),
			log.Int("stats.shards_skipped", zsr.Stats.ShardsSkipped),
			log.Int("stats.match_count", zsr.Stats.MatchCount),
			log.Int("stats.ngram_matches", zsr.Stats.NgramMatches),
			log.Int64("stats.wait_ms", zsr.Stats.Wait.Milliseconds()),
		)
	}
	tr.Finish()

	return zsr, err
}

func (m *meteredSearcher) List(ctx context.Context, q query.Q) (*zoekt.RepoList, error) {
	start := time.Now()

	var cat string
	var tags []trace.Tag
	if m.hostname == "" {
		cat = "ListAll"
	} else {
		cat = "List"
		tags = []trace.Tag{
			{Key: "span.kind", Value: "client"},
			{Key: "peer.address", Value: m.hostname},
			{Key: "peer.service", Value: "zoekt"},
		}
	}

	tr, ctx := trace.New(ctx, "zoekt."+cat, queryString(q), tags...)

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

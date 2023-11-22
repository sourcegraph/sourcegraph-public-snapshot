package backend

import (
	"context"
	"sync"
	"time"

	"github.com/keegancsmith/rpc"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	sglog "github.com/sourcegraph/log"
	"github.com/sourcegraph/zoekt"
	"github.com/sourcegraph/zoekt/query"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/honey"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

var requestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "src_zoekt_request_duration_seconds",
	Help:    "Time (in seconds) spent on request.",
	Buckets: prometheus.DefBuckets,
}, []string{"hostname", "category", "code"})

type meteredSearcher struct {
	zoekt.Streamer

	hostname string
	log      sglog.Logger
}

func NewMeteredSearcher(hostname string, z zoekt.Streamer) zoekt.Streamer {
	return &meteredSearcher{
		Streamer: z,
		hostname: hostname,
		log:      sglog.Scoped("meteredSearcher"),
	}
}

func (m *meteredSearcher) StreamSearch(ctx context.Context, q query.Q, opts *zoekt.SearchOptions, c zoekt.Sender) (err error) {
	start := time.Now()

	// isLeaf is true if this is a zoekt.Searcher which does a network
	// call. False if we are an aggregator. We use this to decide if we need
	// to add RPC tracing and adjust how we record metrics.
	isLeaf := m.hostname != ""

	var cat string
	attrs := []attribute.KeyValue{
		attribute.String("query", queryString(q)),
	}
	if !isLeaf {
		cat = "SearchAll"
	} else {
		cat = "Search"
		attrs = append(attrs,
			attribute.String("span.kind", "client"),
			attribute.String("peer.address", m.hostname),
			attribute.String("peer.service", "zoekt"),
		)
	}

	event := honey.NoopEvent()
	if honey.Enabled() && cat == "SearchAll" {
		event = honey.NewEvent("search-zoekt")
		event.AddField("category", cat)
		event.AddField("actor", actor.FromContext(ctx).UIDString())
		event.AddAttributes(attrs)
	}

	tr, ctx := trace.New(ctx, "zoekt."+cat, attrs...)
	defer func() {
		tr.SetErrorIfNotContext(err)
		tr.End()
	}()
	if opts != nil {
		fields := []attribute.KeyValue{
			attribute.Bool("opts.estimate_doc_count", opts.EstimateDocCount),
			attribute.Bool("opts.whole", opts.Whole),
			attribute.Int("opts.shard_max_match_count", opts.ShardMaxMatchCount),
			attribute.Int("opts.shard_repo_max_match_count", opts.ShardRepoMaxMatchCount),
			attribute.Int("opts.total_max_match_count", opts.TotalMaxMatchCount),
			attribute.Int64("opts.max_wall_time_ms", opts.MaxWallTime.Milliseconds()),
			attribute.Int64("opts.flush_wall_time_ms", opts.FlushWallTime.Milliseconds()),
			attribute.Int("opts.max_doc_display_count", opts.MaxDocDisplayCount),
		}
		tr.AddEvent("begin", fields...)
		event.AddAttributes(fields)
	}

	// We wrap our queries in GobCache, this gives us a convenient way to find
	// out the marshalled size of the query.
	if gobCache, ok := q.(*query.GobCache); ok {
		b, _ := gobCache.GobEncode()
		tr.SetAttributes(attribute.Int("query.size", len(b)))
		event.AddField("query.size", len(b))
	}

	// Instrument the RPC layer
	var writeRequestStart, writeRequestDone time.Time
	if isLeaf {
		ctx = rpc.WithClientTrace(ctx, &rpc.ClientTrace{
			WriteRequestStart: func() {
				tr.SetAttributes(attribute.String("event", "rpc.write_request_start"))
				writeRequestStart = time.Now()
			},

			WriteRequestDone: func(err error) {
				fields := []attribute.KeyValue{}
				if err != nil {
					fields = append(fields, attribute.String("rpc.write_request.error", err.Error()))
				}
				tr.AddEvent("rpc.write_request_done", fields...)
				writeRequestDone = time.Now()
			},
		})
	}

	var (
		code  = "200" // final code to record
		first sync.Once
	)

	mu := sync.Mutex{}
	statsAgg := &zoekt.Stats{}
	nFilesMatches := 0
	nEvents := 0
	var totalSendTimeMs int64

	err = m.Streamer.StreamSearch(ctx, q, opts, ZoektStreamFunc(func(zsr *zoekt.SearchResult) {
		first.Do(func() {
			if isLeaf {
				if !writeRequestStart.IsZero() {
					tr.SetAttributes(
						attribute.Int64("rpc.queue_latency_ms", writeRequestStart.Sub(start).Milliseconds()),
						attribute.Int64("rpc.write_duration_ms", writeRequestDone.Sub(writeRequestStart).Milliseconds()),
					)
				}
				tr.SetAttributes(
					attribute.Int64("stream.latency_ms", time.Since(start).Milliseconds()),
				)
			}
		})

		if zsr != nil {
			mu.Lock()
			statsAgg.Add(zsr.Stats)
			nFilesMatches += len(zsr.Files)
			nEvents++
			mu.Unlock()

			startSend := time.Now()
			c.Send(zsr)
			sendTimeMs := time.Since(startSend).Milliseconds()

			mu.Lock()
			totalSendTimeMs += sendTimeMs
			mu.Unlock()
		}
	}))

	if err != nil {
		code = "error"
	}

	fields := []attribute.KeyValue{
		attribute.Int("filematches", nFilesMatches),
		attribute.Int("events", nEvents),
		attribute.Int64("stream.total_send_time_ms", totalSendTimeMs),

		// Zoekt stats.
		attribute.Int64("stats.content_bytes_loaded", statsAgg.ContentBytesLoaded),
		attribute.Int64("stats.index_bytes_loaded", statsAgg.IndexBytesLoaded),
		attribute.Int("stats.crashes", statsAgg.Crashes),
		attribute.Int("stats.file_count", statsAgg.FileCount),
		attribute.Int("stats.files_considered", statsAgg.FilesConsidered),
		attribute.Int("stats.files_loaded", statsAgg.FilesLoaded),
		attribute.Int("stats.files_skipped", statsAgg.FilesSkipped),
		attribute.Int("stats.match_count", statsAgg.MatchCount),
		attribute.Int("stats.ngram_lookups", statsAgg.NgramLookups),
		attribute.Int("stats.ngram_matches", statsAgg.NgramMatches),
		attribute.Int("stats.shard_files_considered", statsAgg.ShardFilesConsidered),
		attribute.Int("stats.shards_scanned", statsAgg.ShardsScanned),
		attribute.Int("stats.shards_skipped", statsAgg.ShardsSkipped),
		attribute.Int("stats.shards_skipped_filter", statsAgg.ShardsSkippedFilter),
		attribute.Int64("stats.wait_ms", statsAgg.Wait.Milliseconds()),
		attribute.Int64("stats.match_tree_construction_ms", statsAgg.MatchTreeConstruction.Milliseconds()),
		attribute.Int64("stats.match_tree_search_ms", statsAgg.MatchTreeSearch.Milliseconds()),
		attribute.Int("stats.regexps_considered", statsAgg.RegexpsConsidered),
		attribute.String("stats.flush_reason", statsAgg.FlushReason.String()),
	}
	tr.AddEvent("done", fields...)
	event.AddField("duration_ms", time.Since(start).Milliseconds())
	if err != nil {
		event.AddField("error", err.Error())
	}
	event.AddAttributes(fields)
	event.Send()

	// Record total duration of stream
	requestDuration.WithLabelValues(m.hostname, cat, code).Observe(time.Since(start).Seconds())

	return err
}

func (m *meteredSearcher) Search(ctx context.Context, q query.Q, opts *zoekt.SearchOptions) (*zoekt.SearchResult, error) {
	return AggregateStreamSearch(ctx, m.StreamSearch, q, opts)
}

func (m *meteredSearcher) List(ctx context.Context, q query.Q, opts *zoekt.ListOptions) (_ *zoekt.RepoList, err error) {
	start := time.Now()

	var cat string
	var attrs []attribute.KeyValue

	if m.hostname == "" {
		cat = "ListAll"
	} else {
		cat = listCategory(opts)
		attrs = []attribute.KeyValue{
			attribute.String("span.kind", "client"),
			attribute.String("peer.address", m.hostname),
			attribute.String("peer.service", "zoekt"),
		}
	}

	qStr := queryString(q)

	tr, ctx := trace.New(ctx, "zoekt."+cat, attrs...)
	tr.SetAttributes(
		attribute.Stringer("opts", opts),
		attribute.String("query", qStr),
	)
	defer tr.EndWithErr(&err)

	event := honey.NoopEvent()
	if honey.Enabled() && cat == "ListAll" {
		event = honey.NewEvent("search-zoekt")
		event.AddField("category", cat)
		event.AddField("query", qStr)
		event.AddAttributes(attrs)
	}

	zsl, err := m.Streamer.List(ctx, q, opts)

	code := "200"
	if err != nil {
		code = "error"
	}

	event.AddField("duration_ms", time.Since(start).Milliseconds())
	if zsl != nil {
		// the fields are mutually exclusive so we can just add them
		event.AddField("repos", len(zsl.Repos)+len(zsl.ReposMap))
		event.AddField("stats.crashes", zsl.Crashes)
	}
	if err != nil {
		event.AddField("error", err.Error())
	}
	event.Send()

	requestDuration.WithLabelValues(m.hostname, cat, code).Observe(time.Since(start).Seconds())

	if zsl != nil {
		tr.SetAttributes(attribute.Int("repos", len(zsl.Repos)))
	}

	return zsl, err
}

func (m *meteredSearcher) String() string {
	return "MeteredSearcher{" + m.Streamer.String() + "}"
}

func queryString(q query.Q) string {
	if q == nil {
		return "<nil>"
	}
	return q.String()
}

func listCategory(opts *zoekt.ListOptions) string {
	field, err := opts.GetField()
	if err != nil {
		return "ListMisconfigured"
	}

	switch field {
	case zoekt.RepoListFieldRepos:
		return "List"
	case zoekt.RepoListFieldReposMap:
		return "ListReposMap"
	default:
		return "ListUnknown"
	}
}

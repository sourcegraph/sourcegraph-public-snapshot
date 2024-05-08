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
	"github.com/sourcegraph/sourcegraph/lib/errors"
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

	if opts != nil {
		attrs = append(attrs, filterDefaultValue(
			attribute.Bool("opts.estimate_doc_count", opts.EstimateDocCount),
			attribute.Bool("opts.whole", opts.Whole),
			attribute.Int("opts.shard_max_match_count", opts.ShardMaxMatchCount),
			attribute.Int("opts.shard_repo_max_match_count", opts.ShardRepoMaxMatchCount),
			attribute.Int("opts.total_max_match_count", opts.TotalMaxMatchCount),
			attribute.Int64("opts.max_wall_time_ms", opts.MaxWallTime.Milliseconds()),
			attribute.Int64("opts.flush_wall_time_ms", opts.FlushWallTime.Milliseconds()),
			attribute.Int("opts.max_doc_display_count", opts.MaxDocDisplayCount),
			attribute.Int("opts.max_match_display_count", opts.MaxMatchDisplayCount),
			attribute.Int("opts.context_lines", opts.NumContextLines),
			attribute.Bool("opts.chunk_matches", opts.ChunkMatches),
			attribute.Bool("opts.use_document_ranks", opts.UseDocumentRanks),
			attribute.Float64("opts.document_ranks_weight", opts.DocumentRanksWeight),
			attribute.Bool("opts.use_bm25_scoring", opts.UseBM25Scoring),
			attribute.Bool("opts.debug_score", opts.DebugScore),
		)...)
	}

	event := honey.NonSendingEvent()
	if honey.Enabled() && cat == "SearchAll" {
		event = honey.NewEvent("search-zoekt")
		event.AddAttributes([]attribute.KeyValue{
			attribute.String("category", cat),
			attribute.Int("actor", int(actor.FromContext(ctx).UID)),
		})
		event.AddAttributes(attrs)
	}

	tr, ctx := trace.New(ctx, "zoekt."+cat, attrs...)
	defer tr.EndWithErrIfNotContext(&err)

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

	var first sync.Once

	mu := sync.Mutex{}
	statsAgg := &zoekt.Stats{}
	nFilesMatches := 0
	nEvents := 0
	var totalSendTimeMs int64

	err = m.Streamer.StreamSearch(ctx, q, opts, ZoektStreamFunc(func(zsr *zoekt.SearchResult) {
		first.Do(func() {
			latency := attribute.Int64("stream.latency_ms", time.Since(start).Milliseconds())
			tr.SetAttributes(latency)
			event.AddAttributes([]attribute.KeyValue{latency})

			// Only leafs do RPC
			if isLeaf {
				if !writeRequestStart.IsZero() {
					tr.SetAttributes(
						attribute.Int64("rpc.queue_latency_ms", writeRequestStart.Sub(start).Milliseconds()),
						attribute.Int64("rpc.write_duration_ms", writeRequestDone.Sub(writeRequestStart).Milliseconds()),
					)
				}
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

	code, maybeErrorStr := codeAndErrorStr(err)

	fields := []attribute.KeyValue{
		attribute.Int("filematches", nFilesMatches),
		attribute.Int("events", nEvents),
		attribute.Int64("stream.total_send_time_ms", totalSendTimeMs),
		attribute.String("code", code),
	}

	// Zoekt stats, filter out default values to aid readability.
	fields = append(fields, filterDefaultValue(
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
	)...)
	tr.AddEvent("done", fields...)
	event.AddField("duration_ms", time.Since(start).Milliseconds())
	if maybeErrorStr != "" {
		event.AddField("error", maybeErrorStr)
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

	// isLeaf is true if this is a zoekt.Searcher which does a network
	// call. False if we are an aggregator. We use this to decide if we need
	// to add RPC tracing and adjust how we record metrics.
	isLeaf := m.hostname != ""

	// Note: we do not log opts in telemetry directly. It currently only has 1
	// field, and that is covered by the listCategory function.

	var cat string
	attrs := []attribute.KeyValue{
		attribute.String("query", queryString(q)),
	}
	if !isLeaf {
		cat = "ListAll"
	} else {
		cat = listCategory(opts)
		attrs = append(attrs,
			attribute.String("span.kind", "client"),
			attribute.String("peer.address", m.hostname),
			attribute.String("peer.service", "zoekt"),
		)
	}

	event := honey.NonSendingEvent()
	if honey.Enabled() && cat == "ListAll" {
		event = honey.NewEvent("search-zoekt")
		event.AddAttributes([]attribute.KeyValue{
			attribute.String("category", cat),
			attribute.Int("actor", int(actor.FromContext(ctx).UID)),
		})
		event.AddAttributes(attrs)
	}

	tr, ctx := trace.New(ctx, "zoekt."+cat, attrs...)
	defer tr.EndWithErrIfNotContext(&err)

	zsl, err := m.Streamer.List(ctx, q, opts)

	code, maybeErrorStr := codeAndErrorStr(err)

	fields := []attribute.KeyValue{
		attribute.String("code", code),
	}

	if zsl != nil {
		fields = []attribute.KeyValue{
			// the fields are mutually exclusive so we can just add them
			attribute.Int("repos", len(zsl.Repos)+len(zsl.ReposMap)),
			attribute.Int("stats.crashes", zsl.Crashes),
		}
	}

	tr.AddEvent("done", fields...)

	event.AddAttributes(fields)
	event.AddField("duration_ms", time.Since(start).Milliseconds())
	if maybeErrorStr != "" {
		event.AddField("error", maybeErrorStr)
	}
	event.Send()

	requestDuration.WithLabelValues(m.hostname, cat, code).Observe(time.Since(start).Seconds())

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

// filterDefaultValue removes values which are the default. This is used to
// reduce the amount of options we send over + make it easier for a human to
// eyeball.
func filterDefaultValue(attrs ...attribute.KeyValue) []attribute.KeyValue {
	filtered := attrs[:0]

	for _, kv := range attrs {
		isDefault := false

		// We do not handle the slice types
		switch kv.Value.Type() {
		case attribute.BOOL:
			isDefault = !kv.Value.AsBool()
		case attribute.INT64:
			isDefault = kv.Value.AsInt64() == 0
		case attribute.FLOAT64:
			isDefault = kv.Value.AsFloat64() == 0
		case attribute.STRING:
			isDefault = kv.Value.Emit() == ""
		}

		if !isDefault {
			filtered = append(filtered, kv)
		}
	}

	return filtered
}

func codeAndErrorStr(err error) (code, maybeErrStr string) {
	if err == nil {
		return "200", ""
	}
	// Canceled is a not an error due to reaching limits.
	if errors.Is(err, context.Canceled) {
		return "canceled", ""
	}
	// DeadlineExceeded is not an error either, but rather us hitting a
	// timeout.
	if errors.Is(err, context.DeadlineExceeded) {
		return "timeout", ""
	}

	return "error", err.Error()
}

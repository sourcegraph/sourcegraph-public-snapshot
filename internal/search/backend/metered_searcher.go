package backend

import (
	"context"
	"sync"
	"time"

	"github.com/keegancsmith/rpc"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	sglog "github.com/sourcegraph/log"
	"github.com/sourcegraph/zoekt"
	"github.com/sourcegraph/zoekt/query"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/honey"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
	"github.com/sourcegraph/sourcegraph/internal/trace/policy"
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
		log:      sglog.Scoped("meteredSearcher", "wraps zoekt.Streamer with observability"),
	}
}

func (m *meteredSearcher) StreamSearch(ctx context.Context, q query.Q, opts *zoekt.SearchOptions, c zoekt.Sender) (err error) {
	start := time.Now()

	// isLeaf is true if this is a zoekt.Searcher which does a network
	// call. False if we are an aggregator. We use this to decide if we need
	// to add RPC tracing and adjust how we record metrics.
	isLeaf := m.hostname != ""

	var cat string
	var tags []trace.Tag
	if !isLeaf {
		cat = "SearchAll"
	} else {
		cat = "Search"
		tags = []trace.Tag{
			{Key: "span.kind", Value: "client"},
			{Key: "peer.address", Value: m.hostname},
			{Key: "peer.service", Value: "zoekt"},
		}
	}

	qStr := queryString(q)

	event := honey.NoopEvent()
	if honey.Enabled() && cat == "SearchAll" {
		event = honey.NewEvent("search-zoekt")
		event.AddField("category", cat)
		event.AddField("query", qStr)
		event.AddField("actor", actor.FromContext(ctx).UIDString())
		for _, t := range tags {
			event.AddField(t.Key, t.Value)
		}
	}

	tr, ctx := trace.New(ctx, "zoekt."+cat, qStr, tags...)
	defer func() {
		tr.SetErrorIfNotContext(err)
		tr.Finish()
	}()
	if opts != nil {
		fields := []log.Field{
			log.Bool("opts.estimate_doc_count", opts.EstimateDocCount),
			log.Bool("opts.whole", opts.Whole),
			log.Int("opts.shard_max_match_count", opts.ShardMaxMatchCount),
			log.Int("opts.shard_repo_max_match_count", opts.ShardRepoMaxMatchCount),
			log.Int("opts.total_max_match_count", opts.TotalMaxMatchCount),
			log.Int64("opts.max_wall_time_ms", opts.MaxWallTime.Milliseconds()),
			log.Int64("opts.flush_wall_time_ms", opts.FlushWallTime.Milliseconds()),
			log.Int("opts.max_doc_display_count", opts.MaxDocDisplayCount),
			log.Bool("opts.use_document_ranks", opts.UseDocumentRanks),
		}
		tr.LogFields(fields...)
		event.AddLogFields(fields)
	}

	// We wrap our queries in GobCache, this gives us a convenient way to find
	// out the marshalled size of the query.
	if gobCache, ok := q.(*query.GobCache); ok {
		b, _ := gobCache.GobEncode()
		tr.LogFields(log.Int("query.size", len(b)))
		event.AddField("query.size", len(b))
	}

	if isLeaf && opts != nil && policy.ShouldTrace(ctx) {
		// Replace any existing spanContext with a new one, given we've done additional tracing
		spanContext := make(map[string]string)
		if span := opentracing.SpanFromContext(ctx); span == nil {
			m.log.Warn("ctx does not have a trace span associated with it")
		} else if err := ot.GetTracer(ctx).Inject(span.Context(), opentracing.TextMap, opentracing.TextMapCarrier(spanContext)); err == nil {
			newOpts := *opts
			newOpts.SpanContext = spanContext
			opts = &newOpts
		} else {
			m.log.Warn("error injecting new span context into map", sglog.Error(err))
		}
	}

	// Instrument the RPC layer
	var writeRequestStart, writeRequestDone time.Time
	if isLeaf {
		ctx = rpc.WithClientTrace(ctx, &rpc.ClientTrace{
			WriteRequestStart: func() {
				tr.LogFields(log.String("event", "rpc.write_request_start"))
				writeRequestStart = time.Now()
			},

			WriteRequestDone: func(err error) {
				fields := []log.Field{log.String("event", "rpc.write_request_done")}
				if err != nil {
					fields = append(fields, log.String("rpc.write_request.error", err.Error()))
				}
				tr.LogFields(fields...)
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
					tr.LogFields(
						log.Int64("rpc.queue_latency_ms", writeRequestStart.Sub(start).Milliseconds()),
						log.Int64("rpc.write_duration_ms", writeRequestDone.Sub(writeRequestStart).Milliseconds()),
					)
				}
				tr.LogFields(
					log.Int64("stream.latency_ms", time.Since(start).Milliseconds()),
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

	fields := []log.Field{
		log.Int("filematches", nFilesMatches),
		log.Int("events", nEvents),
		log.Int64("stream.total_send_time_ms", totalSendTimeMs),

		// Zoekt stats.
		log.Int64("stats.content_bytes_loaded", statsAgg.ContentBytesLoaded),
		log.Int64("stats.index_bytes_loaded", statsAgg.IndexBytesLoaded),
		log.Int("stats.crashes", statsAgg.Crashes),
		log.Int("stats.file_count", statsAgg.FileCount),
		log.Int("stats.files_considered", statsAgg.FilesConsidered),
		log.Int("stats.files_loaded", statsAgg.FilesLoaded),
		log.Int("stats.files_skipped", statsAgg.FilesSkipped),
		log.Int("stats.match_count", statsAgg.MatchCount),
		log.Int("stats.ngram_matches", statsAgg.NgramMatches),
		log.Int("stats.shard_files_considered", statsAgg.ShardFilesConsidered),
		log.Int("stats.shards_scanned", statsAgg.ShardsScanned),
		log.Int("stats.shards_skipped", statsAgg.ShardsSkipped),
		log.Int("stats.shards_skipped_filter", statsAgg.ShardsSkippedFilter),
		log.Int64("stats.wait_ms", statsAgg.Wait.Milliseconds()),
		log.Int("stats.regexps_considered", statsAgg.RegexpsConsidered),
		log.String("stats.flush_reason", statsAgg.FlushReason.String()),
	}
	tr.LogFields(fields...)
	event.AddField("duration_ms", time.Since(start).Milliseconds())
	if err != nil {
		event.AddField("error", err)
	}
	event.AddLogFields(fields)
	event.Send()

	// Record total duration of stream
	requestDuration.WithLabelValues(m.hostname, cat, code).Observe(time.Since(start).Seconds())

	return err
}

func (m *meteredSearcher) Search(ctx context.Context, q query.Q, opts *zoekt.SearchOptions) (*zoekt.SearchResult, error) {
	return AggregateStreamSearch(ctx, m.StreamSearch, q, opts)
}

func (m *meteredSearcher) List(ctx context.Context, q query.Q, opts *zoekt.ListOptions) (*zoekt.RepoList, error) {
	start := time.Now()

	var cat string
	var tags []trace.Tag

	if m.hostname == "" {
		cat = "ListAll"
	} else {
		if opts == nil || !opts.Minimal {
			cat = "List"
		} else {
			cat = "ListMinimal"
		}
		tags = []trace.Tag{
			{Key: "span.kind", Value: "client"},
			{Key: "peer.address", Value: m.hostname},
			{Key: "peer.service", Value: "zoekt"},
		}
	}

	qStr := queryString(q)

	tr, ctx := trace.New(ctx, "zoekt."+cat, qStr, tags...)
	tr.LogFields(trace.Stringer("opts", opts))

	event := honey.NoopEvent()
	if honey.Enabled() && cat == "ListAll" {
		event = honey.NewEvent("search-zoekt")
		event.AddField("category", cat)
		event.AddField("query", qStr)
		event.AddField("opts.minimal", opts != nil && opts.Minimal)
		for _, t := range tags {
			event.AddField(t.Key, t.Value)
		}
	}

	zsl, err := m.Streamer.List(ctx, q, opts)

	code := "200"
	if err != nil {
		code = "error"
	}

	event.AddField("duration_ms", time.Since(start).Milliseconds())
	if zsl != nil {
		event.AddField("repos", len(zsl.Repos))
		event.AddField("minimal_repos", len(zsl.Minimal))
	}
	if err != nil {
		event.AddField("error", err)
	}
	event.Send()

	requestDuration.WithLabelValues(m.hostname, cat, code).Observe(time.Since(start).Seconds())

	tr.SetError(err)
	if zsl != nil {
		tr.LogFields(log.Int("repos", len(zsl.Repos)))
	}
	tr.Finish()

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

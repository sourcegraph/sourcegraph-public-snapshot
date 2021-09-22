package backend

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/google/zoekt"
	"github.com/google/zoekt/query"
	"github.com/honeycombio/libhoney-go"
	"github.com/inconshreveable/log15"
	"github.com/keegancsmith/rpc"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/xid"

	"github.com/sourcegraph/sourcegraph/internal/honey"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
)

var searchCoreOOMDebug = os.Getenv("SRC_SEARCH_CORE_OOM_DEBUG") != ""

var requestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "src_zoekt_request_duration_seconds",
	Help:    "Time (in seconds) spent on request.",
	Buckets: prometheus.DefBuckets,
}, []string{"hostname", "category", "code"})

type meteredSearcher struct {
	zoekt.Streamer

	hostname string
}

func NewMeteredSearcher(hostname string, z zoekt.Streamer) zoekt.Streamer {
	return &meteredSearcher{
		Streamer: z,
		hostname: hostname,
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

	var evBefore, evAfter *libhoney.Event
	debugAdd := func(key string, value interface{}) {
		if evBefore == nil {
			return
		}
		evBefore.AddField(key, value)
		evAfter.AddField(key, value)
	}
	if searchCoreOOMDebug && cat == "SearchAll" {
		evBefore = honey.Event("search-core-oom-debug")
		evAfter = honey.Event("search-core-oom-debug")
		debugAdd("category", cat)
		debugAdd("query", queryString(q))
		debugAdd("xid", xid.New().String())
		for _, t := range tags {
			debugAdd(t.Key, t.Value)
		}
	}

	tr, ctx := trace.New(ctx, "zoekt."+cat, queryString(q), tags...)
	defer func() {
		tr.SetErrorIfNotContext(err)
		tr.Finish()
	}()
	if opts != nil {
		fields := []log.Field{
			log.Bool("opts.estimate_doc_count", opts.EstimateDocCount),
			log.Bool("opts.whole", opts.Whole),
			log.Int("opts.shard_max_match_count", opts.ShardMaxMatchCount),
			log.Int("opts.total_max_match_count", opts.TotalMaxMatchCount),
			log.Int("opts.shard_max_important_match", opts.ShardMaxImportantMatch),
			log.Int("opts.total_max_important_match", opts.TotalMaxImportantMatch),
			log.Int64("opts.max_wall_time_ms", opts.MaxWallTime.Milliseconds()),
			log.Int("opts.max_doc_display_count", opts.MaxDocDisplayCount),
		}
		tr.LogFields(fields...)
		for _, f := range fields {
			debugAdd(f.Key(), f.Value())
		}
	}

	if isLeaf && opts != nil && ot.ShouldTrace(ctx) {
		// Replace any existing spanContext with a new one, given we've done additional tracing
		spanContext := make(map[string]string)
		if err := ot.GetTracer(ctx).Inject(opentracing.SpanFromContext(ctx).Context(), opentracing.TextMap, opentracing.TextMapCarrier(spanContext)); err == nil {
			newOpts := *opts
			newOpts.SpanContext = spanContext
			opts = &newOpts
		} else {
			log15.Warn("meteredSearcher: Error injecting new span context into map: %s", err)
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

	if evBefore != nil {
		evBefore.Send()
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
		log.Int("stats.shards_skipped", statsAgg.ShardsSkipped),
		log.Int64("stats.wait_ms", statsAgg.Wait.Milliseconds()),
	}
	tr.LogFields(fields...)
	if evAfter != nil {
		evAfter.AddField("duration_ms", time.Since(start).Milliseconds())
		if err != nil {
			evAfter.AddField("error", err)
		}
		for _, f := range fields {
			evAfter.AddField(f.Key(), f.Value())
		}
		evAfter.Send()
	}

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
		cat = "List"
		tags = []trace.Tag{
			{Key: "span.kind", Value: "client"},
			{Key: "peer.address", Value: m.hostname},
			{Key: "peer.service", Value: "zoekt"},
		}
	}

	var evBefore, evAfter *libhoney.Event
	debugAdd := func(key string, value interface{}) {
		if evBefore == nil {
			return
		}
		evBefore.AddField(key, value)
		evAfter.AddField(key, value)
	}
	if searchCoreOOMDebug && cat == "ListAll" {
		evBefore = honey.Event("search-core-oom-debug")
		evAfter = honey.Event("search-core-oom-debug")
		debugAdd("category", cat)
		debugAdd("query", queryString(q))
		debugAdd("xid", xid.New().String())
		for _, t := range tags {
			debugAdd(t.Key, t.Value)
		}
	}

	tr, ctx := trace.New(ctx, "zoekt."+cat, queryString(q), tags...)
	tr.LogFields(trace.Stringer("opts", opts))

	debugAdd("opts.minimal", opts != nil && opts.Minimal)

	if evBefore != nil {
		evBefore.Send()
	}

	zsl, err := m.Streamer.List(ctx, q, opts)

	code := "200"
	if err != nil {
		code = "error"
	}

	if evAfter != nil {
		evAfter.AddField("duration_ms", time.Since(start).Milliseconds())
		if zsl != nil {
			evAfter.AddField("repos", len(zsl.Repos))
			evAfter.AddField("minimal_repos", len(zsl.Minimal))
		}
		if err != nil {
			evAfter.AddField("error", err)
		}
		evAfter.Send()
	}

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

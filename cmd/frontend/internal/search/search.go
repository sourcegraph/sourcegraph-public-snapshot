// Package search is search specific logic for the frontend. Also see
// github.com/sourcegraph/sourcegraph/internal/search for more generic search
// code.
package search

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sourcegraph/log"
	"go.opentelemetry.io/otel/attribute"

	searchlogs "github.com/sourcegraph/sourcegraph/cmd/frontend/internal/search/logs"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/honey"
	searchhoney "github.com/sourcegraph/sourcegraph/internal/honey/search"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/client"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	streamclient "github.com/sourcegraph/sourcegraph/internal/search/streaming/client"
	streamhttp "github.com/sourcegraph/sourcegraph/internal/search/streaming/http"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

// StreamHandler is an http handler which streams back search results.
func StreamHandler(db database.DB) http.Handler {
	logger := log.Scoped("searchStreamHandler")
	return &streamHandler{
		logger:              logger,
		db:                  db,
		searchClient:        client.New(logger, db, gitserver.NewClient("http.search.stream")),
		flushTickerInternal: 100 * time.Millisecond,
		pingTickerInterval:  5 * time.Second,
	}
}

type streamHandler struct {
	logger              log.Logger
	db                  database.DB
	searchClient        client.SearchClient
	flushTickerInternal time.Duration
	pingTickerInterval  time.Duration
}

func (h *streamHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tr, ctx := trace.New(r.Context(), "search.ServeStream")
	defer tr.End()
	r = r.WithContext(ctx)

	streamWriter, err := streamhttp.NewWriter(w)
	if err != nil {
		tr.SetError(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Log events to trace
	streamWriter.StatHook = eventStreamTraceHook(tr.AddEvent)

	eventWriter := newEventWriter(streamWriter)
	defer eventWriter.Done()

	err = h.serveHTTP(r, tr, eventWriter)
	if err != nil {
		eventWriter.Error(err)
		tr.SetError(err)
		return
	}
}

func (h *streamHandler) serveHTTP(r *http.Request, tr trace.Trace, eventWriter *eventWriter) (err error) {
	ctx := r.Context()
	start := time.Now()

	args, err := parseURLQuery(r.URL.Query())
	if err != nil {
		return err
	}
	tr.SetAttributes(
		attribute.String("query", args.Query),
		attribute.String("version", args.Version),
		attribute.String("pattern_type", args.PatternType),
		attribute.Int("search_mode", args.SearchMode),
	)

	inputs, err := h.searchClient.Plan(
		ctx,
		args.Version,
		pointers.NonZeroPtr(args.PatternType),
		args.Query,
		search.Mode(args.SearchMode),
		search.Streaming,
		args.ContextLines,
	)
	if err != nil {
		var queryErr *client.QueryError
		if errors.As(err, &queryErr) {
			eventWriter.Alert(search.AlertForQuery(queryErr.Query, queryErr.Err))
			return nil
		} else {
			return err
		}
	}

	if actor.FromContext(ctx).IsAuthenticated() {
		// Used for development to quickly test different zoekt.SearchOptions without having
		// to change the code.
		inputs.Features.ZoektSearchOptionsOverride = args.ZoektSearchOptionsOverride
	}

	// Display is the number of results we send down. If display is < 0 we
	// want to send everything we find before hitting a limit. Otherwise we
	// can only send up to limit results.
	displayLimit := args.Display
	limit := inputs.MaxResults()
	if displayLimit < 0 || displayLimit > limit {
		displayLimit = limit
	}

	progress := &streamclient.ProgressAggregator{
		Start:        start,
		Limit:        limit,
		Trace:        trace.URL(trace.ID(ctx), conf.DefaultClient()),
		DisplayLimit: displayLimit,
		RepoNamer:    streamclient.RepoNamer(ctx, h.db),
	}

	var latency *time.Duration
	logLatency := func() {
		elapsed := time.Since(start)
		metricLatency.WithLabelValues(string(GuessSource(r))).
			Observe(elapsed.Seconds())
		latency = &elapsed
	}

	// HACK: We awkwardly call an inline function here so that we can defer the
	// cleanups. Defers are guaranteed to run even when unrolling a panic, so
	// we can guarantee that the goroutines spawned by `newEventHandler` are
	// cleaned up when this function exits. This is necessary because otherwise
	// the background goroutines might try to write to the http response, which
	// is no longer valid, which will cause a panic of its own that crashes the
	// process because they are running in a goroutine that does not have a
	// panic handler. We cannot add a panic handler because the goroutines are
	// spawned by the go runtime.
	alert, err := func() (*search.Alert, error) {
		eventHandler := newEventHandler(
			ctx,
			h.logger,
			h.db,
			eventWriter,
			progress,
			h.flushTickerInternal,
			h.pingTickerInterval,
			displayLimit,
			args.EnableChunkMatches,
			logLatency,
		)
		defer eventHandler.Done()

		batchedStream := streaming.NewBatchingStream(50*time.Millisecond, eventHandler)
		defer batchedStream.Done()

		return h.searchClient.Execute(ctx, batchedStream, inputs)
	}()

	if err != nil && errors.HasType(err, &query.UnsupportedError{}) {
		eventWriter.Alert(search.AlertForQuery(inputs.OriginalQuery, err))
		err = nil
	}
	if alert != nil {
		eventWriter.Alert(alert)
	}
	logSearch(ctx, h.logger, alert, err, time.Since(start), latency, inputs.OriginalQuery, progress)
	return err
}

func logSearch(ctx context.Context, logger log.Logger, alert *search.Alert, err error, duration time.Duration, latency *time.Duration, originalQuery string, progress *streamclient.ProgressAggregator) {
	if honey.Enabled() {
		status := client.DetermineStatusForLogs(alert, progress.Stats, err)
		var alertType string
		if alert != nil {
			alertType = alert.PrometheusType
		}

		var latencyMs *int64
		if latency != nil {
			ms := latency.Milliseconds()
			latencyMs = &ms
		}

		_ = searchhoney.SearchEvent(ctx, searchhoney.SearchEventArgs{
			OriginalQuery: originalQuery,
			Typ:           "stream",
			Source:        string(trace.RequestSource(ctx)),
			Status:        status,
			AlertType:     alertType,
			DurationMs:    duration.Milliseconds(),
			LatencyMs:     latencyMs,
			ResultSize:    progress.MatchCount,
			Error:         err,
		}).Send()
	}

	isSlow := duration > searchlogs.LogSlowSearchesThreshold()
	if isSlow {
		logger.Warn("streaming: slow search request", log.String("query", originalQuery))
	}
}

type args struct {
	Query                      string
	Version                    string
	PatternType                string
	Display                    int
	EnableChunkMatches         bool
	SearchMode                 int
	ContextLines               *int32
	ZoektSearchOptionsOverride string
}

func parseURLQuery(q url.Values) (*args, error) {
	get := func(k, def string) string {
		v := q.Get(k)
		if v == "" {
			return def
		}
		return v
	}

	a := args{
		Query:                      get("q", ""),
		Version:                    get("v", "V3"),
		PatternType:                get("t", ""),
		ZoektSearchOptionsOverride: get("zoekt-search-opts", ""),
	}

	if a.Query == "" {
		return nil, errors.New("no query found")
	}

	display := get("display", "-1")
	var err error
	if a.Display, err = strconv.Atoi(display); err != nil {
		return nil, errors.Errorf("display must be an integer, got %q: %w", display, err)
	}

	chunkMatches := get("cm", "f")
	if a.EnableChunkMatches, err = strconv.ParseBool(chunkMatches); err != nil {
		return nil, errors.Errorf("chunk matches must be parseable as a boolean, got %q: %w", chunkMatches, err)
	}

	if contextLines := q.Get("cl"); contextLines != "" {
		parsedContextLines, err := strconv.ParseUint(contextLines, 10, 32)
		if err != nil {
			return nil, errors.Errorf("context lines must be parseable as a boolean, got %q: %w", contextLines, err)
		}
		a.ContextLines = pointers.Ptr(int32(parsedContextLines))
	}

	searchMode := get("sm", "0")
	if a.SearchMode, err = strconv.Atoi(searchMode); err != nil {
		return nil, errors.Errorf("search mode must be integer, got %q: %w", searchMode, err)
	}

	return &a, nil
}

// eventStreamTraceHook returns a StatHook which logs to log.
func eventStreamTraceHook(addEvent func(string, ...attribute.KeyValue)) func(streamhttp.WriterStat) {
	return func(stat streamhttp.WriterStat) {
		fields := []attribute.KeyValue{
			attribute.Int("bytes", stat.Bytes),
			attribute.Int64("duration_ms", stat.Duration.Milliseconds()),
		}
		if stat.Error != nil {
			fields = append(fields, trace.Error(stat.Error))
		}
		addEvent(stat.Event, fields...)
	}
}

var metricLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "src_search_streaming_latency_seconds",
	Help:    "Histogram with time to first result in seconds",
	Buckets: []float64{0.01, 0.02, 0.05, 0.1, 0.2, 0.5, 1, 2, 5, 10, 15, 20, 30},
}, []string{"source"})

var searchBlitzUserAgentRegexp = lazyregexp.New(`^SearchBlitz \(([^\)]+)\)$`)

// GuessSource guesses the source the request came from (browser, other HTTP client, etc.)
func GuessSource(r *http.Request) trace.SourceType {
	userAgent := r.UserAgent()
	for _, guess := range []string{
		"Mozilla",
		"WebKit",
		"Gecko",
		"Chrome",
		"Firefox",
		"Safari",
		"Edge",
	} {
		if strings.Contains(userAgent, guess) {
			return trace.SourceBrowser
		}
	}

	// We send some automated search requests in order to measure baseline search perf. Track the source of these.
	if match := searchBlitzUserAgentRegexp.FindStringSubmatch(userAgent); match != nil {
		return trace.SourceType("searchblitz_" + match[1])
	}

	return trace.SourceOther
}

func repoIDs(results []result.Match) []api.RepoID {
	ids := make(map[api.RepoID]struct{}, 5)
	for _, r := range results {
		ids[r.RepoName().ID] = struct{}{}
	}

	res := make([]api.RepoID, 0, len(ids))
	for id := range ids {
		res = append(res, id)
	}
	return res
}

// newEventHandler creates a stream that can write streaming search events to
// an HTTP stream.
func newEventHandler(
	ctx context.Context,
	logger log.Logger,
	db database.DB,
	eventWriter *eventWriter,
	progress *streamclient.ProgressAggregator,
	flushInterval time.Duration,
	progressInterval time.Duration,
	displayLimit int,
	enableChunkMatches bool,
	logLatency func(),
) *eventHandler {
	// Store marshalled matches and flush periodically or when we go over
	// 32kb. 32kb chosen to be smaller than bufio.MaxTokenSize. Note: we can
	// still write more than that.
	matchesBuf := streamhttp.NewJSONArrayBuf(32*1024, func(data []byte) error {
		return eventWriter.MatchesJSON(data)
	})

	eh := &eventHandler{
		ctx:                ctx,
		logger:             logger,
		db:                 db,
		eventWriter:        eventWriter,
		matchesBuf:         matchesBuf,
		filters:            &streaming.SearchFilters{},
		flushInterval:      flushInterval,
		progress:           progress,
		progressInterval:   progressInterval,
		displayRemaining:   displayLimit,
		enableChunkMatches: enableChunkMatches,
		first:              true,
		logLatency:         logLatency,
	}

	// Schedule the first flushes.
	// Lock because if flushInterval is small, scheduled tick could
	// race with setting eh.flushTimer.
	eh.mu.Lock()
	eh.flushTimer = time.AfterFunc(eh.flushInterval, eh.flushTick)
	eh.progressTimer = time.AfterFunc(eh.progressInterval, eh.progressTick)
	eh.mu.Unlock()

	return eh
}

type eventHandler struct {
	ctx    context.Context
	logger log.Logger
	db     database.DB

	// Config params
	enableChunkMatches bool
	flushInterval      time.Duration
	progressInterval   time.Duration

	logLatency func()

	// Everything below this line is protected by the mutex
	mu sync.Mutex

	eventWriter *eventWriter

	matchesBuf *streamhttp.JSONArrayBuf
	filters    *streaming.SearchFilters
	progress   *streamclient.ProgressAggregator

	// These timers will be non-nil unless Done() was called
	flushTimer    *time.Timer
	progressTimer *time.Timer

	displayRemaining int
	first            bool
}

func (h *eventHandler) Send(event streaming.SearchEvent) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.progress.Update(event)
	h.filters.Update(event)

	h.displayRemaining = event.Results.Limit(h.displayRemaining)

	repoMetadata, err := getEventRepoMetadata(h.ctx, h.db, event)
	if err != nil {
		if !errors.IsContextCanceled(err) {
			h.logger.Error("failed to get repo metadata", log.Error(err))
		}
		return
	}

	for _, match := range event.Results {
		repo := match.RepoName()

		// Don't send matches which we cannot map to a repo the actor has access to. This
		// check is expected to always pass. Missing metadata is a sign that we have
		// searched repos that user shouldn't have access to.
		if md, ok := repoMetadata[repo.ID]; !ok || md.Name != repo.Name {
			continue
		}

		eventMatch := search.FromMatch(match, repoMetadata, h.enableChunkMatches)
		h.matchesBuf.Append(eventMatch)
	}

	// Instantly send results if we have not sent any yet.
	if h.first && len(event.Results) > 0 {
		h.first = false
		h.eventWriter.Filters(h.filters.Compute(), false)
		h.matchesBuf.Flush()
		h.logLatency()
	}
}

// Done cleans up any background tasks and flushes any buffered data to the stream
func (h *eventHandler) Done() {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Cancel any in-flight timers
	h.flushTimer.Stop()
	h.flushTimer = nil
	h.progressTimer.Stop()
	h.progressTimer = nil

	// Flush the final state
	// TODO: make sure we actually respect timeouts
	exhaustive := !h.progress.Stats.IsLimitHit &&
		!h.progress.Stats.Status.Any(search.RepoStatusLimitHit) &&
		!h.progress.Stats.Status.Any(search.RepoStatusTimedOut)
	h.eventWriter.Filters(h.filters.Compute(), exhaustive)
	h.matchesBuf.Flush()
	h.eventWriter.Progress(h.progress.Final())
}

func (h *eventHandler) progressTick() {
	h.mu.Lock()
	defer h.mu.Unlock()

	// a nil progressTimer indicates that Done() was called
	if h.progressTimer != nil {
		h.eventWriter.Progress(h.progress.Current())

		// schedule the next progress event
		h.progressTimer = time.AfterFunc(h.progressInterval, h.progressTick)
	}
}

func (h *eventHandler) flushTick() {
	h.mu.Lock()
	defer h.mu.Unlock()

	// a nil flushTimer indicates that Done() was called
	if h.flushTimer != nil {
		if h.filters.Dirty {
			h.eventWriter.Filters(h.filters.Compute(), false)
		}
		h.matchesBuf.Flush()
		if h.progress.Dirty {
			h.eventWriter.Progress(h.progress.Current())
		}

		// schedule the next flush
		h.flushTimer = time.AfterFunc(h.flushInterval, h.flushTick)
	}
}

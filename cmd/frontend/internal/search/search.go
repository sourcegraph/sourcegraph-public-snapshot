// package search is search specific logic for the frontend. Also see
// github.com/sourcegraph/sourcegraph/internal/search for more generic search
// code.
package search

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	searchlogs "github.com/sourcegraph/sourcegraph/cmd/frontend/internal/search/logs"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/honey"
	searchhoney "github.com/sourcegraph/sourcegraph/internal/honey/search"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/client"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	streamclient "github.com/sourcegraph/sourcegraph/internal/search/streaming/client"
	streamhttp "github.com/sourcegraph/sourcegraph/internal/search/streaming/http"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// StreamHandler is an http handler which streams back search results.
func StreamHandler(db database.DB) http.Handler {
	logger := log.Scoped("searchStreamHandler", "")
	return &streamHandler{
		logger:              logger,
		db:                  db,
		searchClient:        client.NewSearchClient(logger, db, search.Indexed(), search.SearcherURLs()),
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
	tr, ctx := trace.New(r.Context(), "search.ServeStream", "")
	defer tr.Finish()
	r = r.WithContext(ctx)

	streamWriter, err := streamhttp.NewWriter(w)
	if err != nil {
		tr.SetError(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Log events to trace
	streamWriter.StatHook = eventStreamOTHook(tr.LogFields)

	eventWriter := newEventWriter(streamWriter)
	defer eventWriter.Done()

	err = h.serveHTTP(r, tr, eventWriter)
	if err != nil {
		eventWriter.Error(err)
		tr.SetError(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *streamHandler) serveHTTP(r *http.Request, tr *trace.Trace, eventWriter *eventWriter) (err error) {
	ctx := r.Context()
	start := time.Now()

	args, err := parseURLQuery(r.URL.Query())
	if err != nil {
		return err
	}
	tr.TagFields(
		otlog.String("query", args.Query),
		otlog.String("version", args.Version),
		otlog.String("pattern_type", args.PatternType),
	)

	settings, err := graphqlbackend.DecodedViewerFinalSettings(ctx, h.db)
	if err != nil {
		return err
	}

	inputs, err := h.searchClient.Plan(ctx, args.Version, strPtr(args.PatternType), args.Query, search.Streaming, settings, envvar.SourcegraphDotComMode())
	if err != nil {
		var queryErr *client.QueryError
		if errors.As(err, &queryErr) {
			eventWriter.Alert(search.AlertForQuery(queryErr.Query, queryErr.Err))
			return nil
		} else {
			return err
		}
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

	var wgLogLatency sync.WaitGroup
	defer wgLogLatency.Wait()
	logLatency := func() {
		wgLogLatency.Add(1)
		go func() {
			defer wgLogLatency.Done()
			metricLatency.WithLabelValues(string(GuessSource(r))).
				Observe(time.Since(start).Seconds())

			graphqlbackend.LogSearchLatency(ctx, h.db, inputs, int32(time.Since(start).Milliseconds()))
		}()
	}

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
	batchedStream := streaming.NewBatchingStream(50*time.Millisecond, eventHandler)
	alert, err := h.searchClient.Execute(ctx, batchedStream, inputs)
	// Clean up streams before writing to eventWriter again.
	batchedStream.Done()
	eventHandler.Done()
	if alert != nil {
		eventWriter.Alert(alert)
	}
	logSearch(ctx, h.logger, alert, err, start, inputs.OriginalQuery, progress)
	return err
}

func logSearch(ctx context.Context, logger log.Logger, alert *search.Alert, err error, start time.Time, originalQuery string, progress *streamclient.ProgressAggregator) {
	status := graphqlbackend.DetermineStatusForLogs(alert, progress.Stats, err)

	var alertType string
	if alert != nil {
		alertType = alert.PrometheusType
	}

	isSlow := time.Since(start) > searchlogs.LogSlowSearchesThreshold()
	if honey.Enabled() || isSlow {
		ev := searchhoney.SearchEvent(ctx, searchhoney.SearchEventArgs{
			OriginalQuery: originalQuery,
			Typ:           "stream",
			Source:        string(trace.RequestSource(ctx)),
			Status:        status,
			AlertType:     alertType,
			DurationMs:    time.Since(start).Milliseconds(),
			ResultSize:    progress.MatchCount,
			Error:         err,
		})

		if honey.Enabled() {
			_ = ev.Send()
		}

		if isSlow {
			logger.Warn("streaming: slow search request", log.String("query", originalQuery))
		}
	}
}

type args struct {
	Query              string
	Version            string
	PatternType        string
	Display            int
	EnableChunkMatches bool

	// Optional decoration parameters for server-side rendering a result set
	// or subset. Decorations may specify, e.g., highlighting results with
	// HTML markup up-front, and/or including context lines around file results.
	DecorationLimit        int    // The initial number of files to decorate in the result set.
	DecorationKind         string // The kind of decoration to apply (HTML highlighting, plaintext, etc.)
	DecorationContextLines int    // The number of lines of context to include around lines with matches.
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
		Query:          get("q", ""),
		Version:        get("v", "V3"),
		PatternType:    get("t", ""),
		DecorationKind: get("dk", "html"),
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

	decorationLimit := get("dl", "0")
	if a.DecorationLimit, err = strconv.Atoi(decorationLimit); err != nil {
		return nil, errors.Errorf("decorationLimit must be an integer, got %q: %w", decorationLimit, err)
	}

	decorationContextLines := get("dc", "1")
	if a.DecorationContextLines, err = strconv.Atoi(decorationContextLines); err != nil {
		return nil, errors.Errorf("decorationContextLines must be an integer, got %q: %w", decorationContextLines, err)
	}

	return &a, nil
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// withDecoration hydrates event match with decorated hunks for a corresponding file match.
func withDecoration(ctx context.Context, db database.DB, eventMatch streamhttp.EventMatch, internalResult result.Match, kind string, contextLines int) streamhttp.EventMatch {
	// FIXME: Use contextLines to constrain hunks.
	_ = contextLines
	if _, ok := internalResult.(*result.FileMatch); !ok {
		return eventMatch
	}

	event, ok := eventMatch.(*streamhttp.EventContentMatch)
	if !ok {
		return eventMatch
	}

	if kind == "html" {
		event.Hunks = DecorateFileHunksHTML(ctx, db, internalResult.(*result.FileMatch))
	}

	// TODO(team/search-product): support additional decoration for terminal clients #24617.
	return eventMatch
}

func fromMatch(match result.Match, repoCache map[api.RepoID]*types.SearchedRepo, enableChunkMatches bool) streamhttp.EventMatch {
	switch v := match.(type) {
	case *result.FileMatch:
		return fromFileMatch(v, repoCache, enableChunkMatches)
	case *result.RepoMatch:
		return fromRepository(v, repoCache)
	case *result.CommitMatch:
		return fromCommit(v, repoCache)
	default:
		panic(fmt.Sprintf("unknown match type %T", v))
	}
}

func fromFileMatch(fm *result.FileMatch, repoCache map[api.RepoID]*types.SearchedRepo, enableChunkMatches bool) streamhttp.EventMatch {
	if len(fm.Symbols) > 0 {
		return fromSymbolMatch(fm, repoCache)
	} else if fm.ChunkMatches.MatchCount() > 0 {
		return fromContentMatch(fm, repoCache, enableChunkMatches)
	}
	return fromPathMatch(fm, repoCache)
}

func fromPathMatch(fm *result.FileMatch, repoCache map[api.RepoID]*types.SearchedRepo) *streamhttp.EventPathMatch {
	pathEvent := &streamhttp.EventPathMatch{
		Type:         streamhttp.PathMatchType,
		Path:         fm.Path,
		PathMatches:  fromRanges(fm.PathMatches),
		Repository:   string(fm.Repo.Name),
		RepositoryID: int32(fm.Repo.ID),
		Commit:       string(fm.CommitID),
	}

	if r, ok := repoCache[fm.Repo.ID]; ok {
		pathEvent.RepoStars = r.Stars
		pathEvent.RepoLastFetched = r.LastFetched
	}

	if fm.InputRev != nil {
		pathEvent.Branches = []string{*fm.InputRev}
	}

	return pathEvent
}

func fromChunkMatches(cms result.ChunkMatches) []streamhttp.ChunkMatch {
	res := make([]streamhttp.ChunkMatch, 0, len(cms))
	for _, cm := range cms {
		res = append(res, fromChunkMatch(cm))
	}
	return res
}

func fromChunkMatch(cm result.ChunkMatch) streamhttp.ChunkMatch {
	return streamhttp.ChunkMatch{
		Content:      cm.Content,
		ContentStart: fromLocation(cm.ContentStart),
		Ranges:       fromRanges(cm.Ranges),
	}
}

func fromLocation(l result.Location) streamhttp.Location {
	return streamhttp.Location{
		Offset: l.Offset,
		Line:   l.Line,
		Column: l.Column,
	}
}

func fromRanges(rs result.Ranges) []streamhttp.Range {
	res := make([]streamhttp.Range, 0, len(rs))
	for _, r := range rs {
		res = append(res, streamhttp.Range{
			Start: fromLocation(r.Start),
			End:   fromLocation(r.End),
		})
	}
	return res
}

func fromContentMatch(fm *result.FileMatch, repoCache map[api.RepoID]*types.SearchedRepo, enableChunkMatches bool) *streamhttp.EventContentMatch {

	var (
		eventLineMatches  []streamhttp.EventLineMatch
		eventChunkMatches []streamhttp.ChunkMatch
	)

	if enableChunkMatches {
		eventChunkMatches = fromChunkMatches(fm.ChunkMatches)
	} else {
		lineMatches := fm.ChunkMatches.AsLineMatches()
		eventLineMatches = make([]streamhttp.EventLineMatch, 0, len(lineMatches))
		for _, lm := range lineMatches {
			eventLineMatches = append(eventLineMatches, streamhttp.EventLineMatch{
				Line:             lm.Preview,
				LineNumber:       lm.LineNumber,
				OffsetAndLengths: lm.OffsetAndLengths,
			})
		}
	}

	contentEvent := &streamhttp.EventContentMatch{
		Type:         streamhttp.ContentMatchType,
		Path:         fm.Path,
		RepositoryID: int32(fm.Repo.ID),
		Repository:   string(fm.Repo.Name),
		Commit:       string(fm.CommitID),
		LineMatches:  eventLineMatches,
		ChunkMatches: eventChunkMatches,
	}

	if fm.InputRev != nil {
		contentEvent.Branches = []string{*fm.InputRev}
	}

	if r, ok := repoCache[fm.Repo.ID]; ok {
		contentEvent.RepoStars = r.Stars
		contentEvent.RepoLastFetched = r.LastFetched
	}

	return contentEvent
}

func fromSymbolMatch(fm *result.FileMatch, repoCache map[api.RepoID]*types.SearchedRepo) *streamhttp.EventSymbolMatch {
	symbols := make([]streamhttp.Symbol, 0, len(fm.Symbols))
	for _, sym := range fm.Symbols {
		kind := sym.Symbol.LSPKind()
		kindString := "UNKNOWN"
		if kind != 0 {
			kindString = strings.ToUpper(kind.String())
		}

		symbols = append(symbols, streamhttp.Symbol{
			URL:           sym.URL().String(),
			Name:          sym.Symbol.Name,
			ContainerName: sym.Symbol.Parent,
			Kind:          kindString,
			Line:          int32(sym.Symbol.Line),
		})
	}

	symbolMatch := &streamhttp.EventSymbolMatch{
		Type:         streamhttp.SymbolMatchType,
		Path:         fm.Path,
		Repository:   string(fm.Repo.Name),
		RepositoryID: int32(fm.Repo.ID),
		Commit:       string(fm.CommitID),
		Symbols:      symbols,
	}

	if r, ok := repoCache[fm.Repo.ID]; ok {
		symbolMatch.RepoStars = r.Stars
		symbolMatch.RepoLastFetched = r.LastFetched
	}

	if fm.InputRev != nil {
		symbolMatch.Branches = []string{*fm.InputRev}
	}

	return symbolMatch
}

func fromRepository(rm *result.RepoMatch, repoCache map[api.RepoID]*types.SearchedRepo) *streamhttp.EventRepoMatch {
	var branches []string
	if rev := rm.Rev; rev != "" {
		branches = []string{rev}
	}

	repoEvent := &streamhttp.EventRepoMatch{
		Type:         streamhttp.RepoMatchType,
		RepositoryID: int32(rm.ID),
		Repository:   string(rm.Name),
		Branches:     branches,
	}

	if len(rm.DescriptionMatches) > 0 {
		dms := make([]streamhttp.Range, 0, len(rm.DescriptionMatches))
		for _, matchRange := range rm.DescriptionMatches {
			dms = append(dms, streamhttp.Range{
				Start: fromLocation(matchRange.Start),
				End:   fromLocation(matchRange.End),
			})
		}
		repoEvent.DescriptionMatches = dms
	}

	if r, ok := repoCache[rm.ID]; ok {
		repoEvent.RepoStars = r.Stars
		repoEvent.RepoLastFetched = r.LastFetched
		repoEvent.Description = r.Description
		repoEvent.Fork = r.Fork
		repoEvent.Archived = r.Archived
		repoEvent.Private = r.Private
		repoEvent.KeyValuePairs = r.KeyValuePairs
	}

	return repoEvent
}

func fromCommit(commit *result.CommitMatch, repoCache map[api.RepoID]*types.SearchedRepo) *streamhttp.EventCommitMatch {
	hls := commit.Body().ToHighlightedString()
	ranges := make([][3]int32, len(hls.Highlights))
	for i, h := range hls.Highlights {
		ranges[i] = [3]int32{h.Line, h.Character, h.Length}
	}

	commitEvent := &streamhttp.EventCommitMatch{
		Type:         streamhttp.CommitMatchType,
		Label:        commit.Label(),
		URL:          commit.URL().String(),
		Detail:       commit.Detail(),
		Repository:   string(commit.Repo.Name),
		RepositoryID: int32(commit.Repo.ID),
		OID:          string(commit.Commit.ID),
		Message:      string(commit.Commit.Message),
		AuthorName:   commit.Commit.Author.Name,
		AuthorDate:   commit.Commit.Author.Date,
		Content:      hls.Value,
		Ranges:       ranges,
	}

	if r, ok := repoCache[commit.Repo.ID]; ok {
		commitEvent.RepoStars = r.Stars
		commitEvent.RepoLastFetched = r.LastFetched
	}

	return commitEvent
}

// eventStreamOTHook returns a StatHook which logs to log.
func eventStreamOTHook(log func(...otlog.Field)) func(streamhttp.WriterStat) {
	return func(stat streamhttp.WriterStat) {
		fields := []otlog.Field{
			otlog.String("streamhttp.Event", stat.Event),
			otlog.Int("bytes", stat.Bytes),
			otlog.Int64("duration_ms", stat.Duration.Milliseconds()),
		}
		if stat.Error != nil {
			fields = append(fields, otlog.Error(stat.Error))
		}
		log(fields...)
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
	for _, result := range results {
		ids[result.RepoName().ID] = struct{}{}
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
		h.logger.Error("failed to get repo metadata", log.Error(err))
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

		eventMatch := fromMatch(match, repoMetadata, h.enableChunkMatches)
		h.matchesBuf.Append(eventMatch)
	}

	// Instantly send results if we have not sent any yet.
	if h.first && len(event.Results) > 0 {
		h.first = false
		h.eventWriter.Filters(h.filters.Compute())
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
	h.eventWriter.Filters(h.filters.Compute())
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
		h.eventWriter.Filters(h.filters.Compute())
		h.matchesBuf.Flush()
		if h.progress.Dirty {
			h.eventWriter.Progress(h.progress.Current())
		}

		// schedule the next flush
		h.flushTimer = time.AfterFunc(h.flushInterval, h.flushTick)
	}
}

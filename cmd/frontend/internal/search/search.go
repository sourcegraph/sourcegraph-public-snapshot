// package search is search specific logic for the frontend. Also see
// github.com/sourcegraph/sourcegraph/internal/search for more generic search
// code.
package search

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/inconshreveable/log15"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

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
	"github.com/sourcegraph/sourcegraph/internal/search/run"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	streamhttp "github.com/sourcegraph/sourcegraph/internal/search/streaming/http"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// StreamHandler is an http handler which streams back search results.
func StreamHandler(db database.DB) http.Handler {
	return &streamHandler{
		db:                  db,
		searchClient:        client.NewSearchClient(db, search.Indexed(), search.SearcherURLs()),
		flushTickerInternal: 100 * time.Millisecond,
		pingTickerInterval:  5 * time.Second,
	}
}

type streamHandler struct {
	db                  database.DB
	searchClient        client.SearchClient
	flushTickerInternal time.Duration
	pingTickerInterval  time.Duration
}

func (h *streamHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	args, err := parseURLQuery(r.URL.Query())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tr, ctx := trace.New(ctx, "search.ServeStream", args.Query,
		trace.Tag{Key: "version", Value: args.Version},
		trace.Tag{Key: "pattern_type", Value: args.PatternType},
	)
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	eventWriter, err := streamhttp.NewWriter(w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Always send a final done event so clients know the stream is shutting
	// down.
	defer eventWriter.Event("done", map[string]any{})

	// Log events to trace
	eventWriter.StatHook = eventStreamOTHook(tr.LogFields)

	events, inputs, results := h.startSearch(ctx, args)

	// Display is the number of results we send down. If display is < 0 we
	// want to send everything we find before hitting a limit. Otherwise we
	// can only send up to limit results.
	display := args.Display
	limit := inputs.MaxResults()
	if display < 0 || display > limit {
		display = limit
	}

	start := time.Now()

	displayLimit := display
	if display < 0 {
		displayLimit = math.MaxInt32
	}
	progress := progressAggregator{
		Start:        start,
		Limit:        inputs.MaxResults(),
		Trace:        trace.URL(trace.ID(ctx), conf.ExternalURL(), conf.Tracer()),
		DisplayLimit: displayLimit,
		RepoNamer:    repoNamer(ctx, h.db),
	}

	sendProgress := func() {
		_ = eventWriter.Event("progress", progress.Current())
	}

	// Store marshalled matches and flush periodically or when we go over
	// 32kb. 32kb chosen to be smaller than bufio.MaxTokenSize. Note: we can
	// still write more than that.
	matchesBuf := streamhttp.NewJSONArrayBuf(32*1024, func(data []byte) error {
		return eventWriter.EventBytes("matches", data)
	})
	matchesFlush := func() {
		if err := matchesBuf.Flush(); err != nil {
			// EOF
			return
		}

		if progress.Dirty {
			sendProgress()
		}
	}
	flushTicker := time.NewTicker(h.flushTickerInternal)
	defer flushTicker.Stop()

	pingTicker := time.NewTicker(h.pingTickerInterval)
	defer pingTicker.Stop()

	filters := &streaming.SearchFilters{}
	filtersFlush := func() {
		if fs := filters.Compute(); len(fs) > 0 {
			buf := make([]streamhttp.EventFilter, 0, len(fs))
			for _, f := range fs {
				buf = append(buf, streamhttp.EventFilter{
					Value:    f.Value,
					Label:    f.Label,
					Count:    f.Count,
					LimitHit: f.IsLimitHit,
					Kind:     f.Kind,
				})
			}

			if err := eventWriter.Event("filters", buf); err != nil {
				// EOF
				return
			}
		}
	}

	var wgLogLatency sync.WaitGroup
	defer wgLogLatency.Wait()

	first := true
	handleEvent := func(event streaming.SearchEvent) {
		progress.Update(event)
		filters.Update(event)

		// Truncate the event to the match limit before fetching repo metadata
		display = event.Results.Limit(display)

		repoMetadata, err := getEventRepoMetadata(ctx, h.db, event)
		if err != nil {
			log15.Error("failed to get repo metadata", "error", err)
			return
		}

		for i, match := range event.Results {
			repo := match.RepoName()

			// Don't send matches which we cannot map to a repo the actor has access to. This
			// check is expected to always pass. Missing metadata is a sign that we have
			// searched repos that user shouldn't have access to.
			if md, ok := repoMetadata[repo.ID]; !ok || md.Name != repo.Name {
				continue
			}

			eventMatch := fromMatch(match, repoMetadata)
			if args.DecorationLimit == -1 || args.DecorationLimit > i {
				eventMatch = withDecoration(ctx, h.db, eventMatch, match, args.DecorationKind, args.DecorationContextLines)
			}
			_ = matchesBuf.Append(eventMatch)
		}

		// Instantly send results if we have not sent any yet.
		if first && matchesBuf.Len() > 0 {
			first = false
			matchesFlush()
			filtersFlush()

			metricLatency.WithLabelValues(string(GuessSource(r))).
				Observe(time.Since(start).Seconds())

			graphqlbackend.LogSearchLatency(ctx, h.db, &wgLogLatency, inputs, int32(time.Since(start).Milliseconds()))
		}
	}

LOOP:
	for {
		select {
		case event, ok := <-events:
			if !ok {
				break LOOP
			}
			handleEvent(event)
		case <-flushTicker.C:
			filtersFlush()
			matchesFlush()
		case <-pingTicker.C:
			sendProgress()
		}
	}

	filtersFlush()
	matchesFlush()

	alert, err := results()
	if err != nil {
		_ = eventWriter.Event("error", streamhttp.EventError{Message: err.Error()})
		return
	}

	if alert != nil {
		var pqs []streamhttp.ProposedQuery
		for _, pq := range alert.ProposedQueries {
			pqs = append(pqs, streamhttp.ProposedQuery{
				Description: pq.Description,
				Query:       pq.QueryString(),
			})
		}
		_ = eventWriter.Event("alert", streamhttp.EventAlert{
			Title:           alert.Title,
			Description:     alert.Description,
			ProposedQueries: pqs,
		})
	}

	_ = eventWriter.Event("progress", progress.Final())

	var status, alertType string
	status = graphqlbackend.DetermineStatusForLogs(alert, progress.Stats, err)
	if alert != nil {
		alertType = alert.PrometheusType
	}

	isSlow := time.Since(start) > searchlogs.LogSlowSearchesThreshold()
	if honey.Enabled() || isSlow {
		ev := searchhoney.SearchEvent(ctx, searchhoney.SearchEventArgs{
			OriginalQuery: inputs.OriginalQuery,
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
			log15.Warn("streaming: slow search request", searchlogs.MapToLog15Ctx(ev.Fields())...)
		}
	}
}

// startSearch will start a search. It returns the events channel which
// streams out search events. Once events is closed you can call results which
// will return the results resolver and error.
func (h *streamHandler) startSearch(ctx context.Context, a *args) (<-chan streaming.SearchEvent, *run.SearchInputs, func() (*search.Alert, error)) {
	eventsC := make(chan streaming.SearchEvent)
	stream := streaming.StreamFunc(func(event streaming.SearchEvent) {
		eventsC <- event
	})
	batchedStream := streaming.NewBatchingStream(50*time.Millisecond, stream)

	settings, err := graphqlbackend.DecodedViewerFinalSettings(ctx, h.db)
	if err != nil {
		close(eventsC)
		return eventsC, &run.SearchInputs{}, func() (*search.Alert, error) {
			return nil, err
		}
	}

	inputs, err := h.searchClient.Plan(ctx, a.Version, strPtr(a.PatternType), a.Query, search.Streaming, settings, envvar.SourcegraphDotComMode())
	if err != nil {
		close(eventsC)
		var queryErr *run.QueryError
		if errors.As(err, &queryErr) {
			return eventsC, &run.SearchInputs{}, func() (*search.Alert, error) {
				return search.AlertForQuery(queryErr.Query, queryErr.Err), nil
			}
		}
		return eventsC, &run.SearchInputs{}, func() (*search.Alert, error) {
			return nil, err
		}
	}

	type finalResult struct {
		alert *search.Alert
		err   error
	}
	final := make(chan finalResult, 1)
	go func() {
		defer close(final)
		defer close(eventsC)
		defer batchedStream.Done()

		alert, err := h.searchClient.Execute(ctx, batchedStream, inputs)
		final <- finalResult{alert: alert, err: err}
	}()

	return eventsC, inputs, func() (*search.Alert, error) {
		f := <-final
		return f.alert, f.err
	}
}

type args struct {
	Query       string
	Version     string
	PatternType string
	Display     int

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
		Version:        get("v", "V2"),
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

func fromMatch(match result.Match, repoCache map[api.RepoID]*types.SearchedRepo) streamhttp.EventMatch {
	switch v := match.(type) {
	case *result.FileMatch:
		return fromFileMatch(v, repoCache)
	case *result.RepoMatch:
		return fromRepository(v, repoCache)
	case *result.CommitMatch:
		return fromCommit(v, repoCache)
	default:
		panic(fmt.Sprintf("unknown match type %T", v))
	}
}

func fromFileMatch(fm *result.FileMatch, repoCache map[api.RepoID]*types.SearchedRepo) streamhttp.EventMatch {
	if len(fm.Symbols) > 0 {
		return fromSymbolMatch(fm, repoCache)
	} else if fm.ChunkMatches.MatchCount() > 0 {
		return fromContentMatch(fm, repoCache)
	}
	return fromPathMatch(fm, repoCache)
}

func fromPathMatch(fm *result.FileMatch, repoCache map[api.RepoID]*types.SearchedRepo) *streamhttp.EventPathMatch {
	pathEvent := &streamhttp.EventPathMatch{
		Type:         streamhttp.PathMatchType,
		Path:         fm.Path,
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

func fromContentMatch(fm *result.FileMatch, repoCache map[api.RepoID]*types.SearchedRepo) *streamhttp.EventContentMatch {
	lineMatches := fm.ChunkMatches.AsLineMatches()
	eventLineMatches := make([]streamhttp.EventLineMatch, 0, len(lineMatches))
	for _, lm := range lineMatches {
		eventLineMatches = append(eventLineMatches, streamhttp.EventLineMatch{
			Line:             lm.Preview,
			LineNumber:       lm.LineNumber,
			OffsetAndLengths: lm.OffsetAndLengths,
		})
	}

	contentEvent := &streamhttp.EventContentMatch{
		Type:         streamhttp.ContentMatchType,
		Path:         fm.Path,
		RepositoryID: int32(fm.Repo.ID),
		Repository:   string(fm.Repo.Name),
		Commit:       string(fm.CommitID),
		LineMatches:  eventLineMatches,
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

	if r, ok := repoCache[rm.ID]; ok {
		repoEvent.RepoStars = r.Stars
		repoEvent.RepoLastFetched = r.LastFetched
		repoEvent.Description = r.Description
		repoEvent.Fork = r.Fork
		repoEvent.Archived = r.Archived
		repoEvent.Private = r.Private
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
		Type:       streamhttp.CommitMatchType,
		Label:      commit.Label(),
		URL:        commit.URL().String(),
		Detail:     commit.Detail(),
		Repository: string(commit.Repo.Name),
		OID:        string(commit.Commit.ID),
		Message:    string(commit.Commit.Message),
		AuthorName: commit.Commit.Author.Name,
		AuthorDate: commit.Commit.Author.Date,
		Content:    hls.Value,
		Ranges:     ranges,
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

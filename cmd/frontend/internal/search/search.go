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
	"time"

	"github.com/cockroachdb/errors"
	"github.com/inconshreveable/log15"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	searchlogs "github.com/sourcegraph/sourcegraph/cmd/frontend/internal/search/logs"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/honey"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/run"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	streamhttp "github.com/sourcegraph/sourcegraph/internal/search/streaming/http"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

// StreamHandler is an http handler which streams back search results.
func StreamHandler(db dbutil.DB) http.Handler {
	return &streamHandler{
		db:                  db,
		newSearchResolver:   defaultNewSearchResolver,
		flushTickerInternal: 100 * time.Millisecond,
		pingTickerInterval:  5 * time.Second,
	}
}

type streamHandler struct {
	db                  dbutil.DB
	newSearchResolver   func(context.Context, dbutil.DB, *graphqlbackend.SearchArgs) (searchResolver, error)
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
	defer eventWriter.Event("done", map[string]interface{}{})

	// Log events to trace
	eventWriter.StatHook = eventStreamOTHook(tr.LogFields)

	events, inputs, results := h.startSearch(ctx, args)
	events = batchEvents(events, 50*time.Millisecond)

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
		Trace:        trace.URL(trace.ID(ctx)),
		DisplayLimit: displayLimit,
	}

	sendProgress := func() {
		_ = eventWriter.Event("progress", progress.Current())
	}

	filters := &streaming.SearchFilters{
		Globbing: false, // TODO
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

	first := true
	handleEvent := func(event streaming.SearchEvent) {
		progress.Update(event)
		filters.Update(event)

		// Truncate the event to the match limit before fetching repo metadata
		for i, match := range event.Results {
			if display <= 0 {
				event.Results = event.Results[:i]
				break
			}

			display = match.Limit(display)
		}

		repoMetadata, err := getEventRepoMetadata(ctx, h.db, event)
		if err != nil {
			log15.Error("failed to get repo metadata", "error", err)
			return
		}
		for i, match := range event.Results {
			// Don't send matches which we cannot map to a repo the actor has access to. This
			// check is expected to always pass. Missing metadata is a sign that we have
			// searched repos that user shouldn't have access to.
			if md, ok := repoMetadata[match.RepoName().ID]; !ok || md.Name != match.RepoName().Name {
				continue
			}
			eventMatch := fromMatch(match, repoMetadata)
			if args.DecorationLimit == -1 || args.DecorationLimit > i {
				eventMatch = withDecoration(ctx, eventMatch, match, args.DecorationKind, args.DecorationContextLines)
			}
			_ = matchesBuf.Append(eventMatch)
		}

		// Instantly send results if we have not sent any yet.
		if first && matchesBuf.Len() > 0 {
			first = false
			matchesFlush()

			metricLatency.WithLabelValues(string(GuessSource(r))).
				Observe(time.Since(start).Seconds())

			graphqlbackend.LogSearchLatency(ctx, h.db, &inputs, int32(time.Since(start).Milliseconds()))
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
			matchesFlush()
		case <-pingTicker.C:
			sendProgress()
		}
	}

	matchesFlush()

	// Send dynamic filters once.
	if filters := filters.Compute(); len(filters) > 0 {
		buf := make([]streamhttp.EventFilter, 0, len(filters))
		for _, f := range filters {
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

	resultsResolver, err := results()
	if err != nil {
		_ = eventWriter.Event("error", streamhttp.EventError{Message: err.Error()})
		return
	}

	alert := resultsResolver.Alert()
	if alert != nil {
		var pqs []streamhttp.ProposedQuery
		if proposed := alert.ProposedQueries(); proposed != nil {
			for _, pq := range *proposed {
				pqs = append(pqs, streamhttp.ProposedQuery{
					Description: fromStrPtr(pq.Description()),
					Query:       pq.Query(),
				})
			}
		}
		_ = eventWriter.Event("alert", streamhttp.EventAlert{
			Title:           alert.Title(),
			Description:     fromStrPtr(alert.Description()),
			ProposedQueries: pqs,
		})
	}

	_ = eventWriter.Event("progress", progress.Final())

	var status, alertType string
	status = graphqlbackend.DetermineStatusForLogs(resultsResolver, err)
	if alert != nil {
		alertType = alert.PrometheusType()
	}

	isSlow := time.Since(start) > searchlogs.LogSlowSearchesThreshold()
	if honey.Enabled() || isSlow {
		ev := honey.SearchEvent(ctx, honey.SearchEventArgs{
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
func (h *streamHandler) startSearch(ctx context.Context, a *args) (events <-chan streaming.SearchEvent, inputs run.SearchInputs, results func() (*graphqlbackend.SearchResultsResolver, error)) {
	eventsC := make(chan streaming.SearchEvent)

	search, err := h.newSearchResolver(ctx, h.db, &graphqlbackend.SearchArgs{
		Query:       a.Query,
		Version:     a.Version,
		PatternType: strPtr(a.PatternType),

		Stream: streaming.StreamFunc(func(event streaming.SearchEvent) {
			eventsC <- event
		}),
	})
	if err != nil {
		close(eventsC)
		return eventsC, run.SearchInputs{}, func() (*graphqlbackend.SearchResultsResolver, error) {
			return nil, err
		}
	}

	type finalResult struct {
		resultsResolver *graphqlbackend.SearchResultsResolver
		err             error
	}
	final := make(chan finalResult, 1)
	go func() {
		defer close(final)
		defer close(eventsC)

		r, err := search.Results(ctx)
		final <- finalResult{resultsResolver: r, err: err}
	}()

	return eventsC, search.Inputs(), func() (*graphqlbackend.SearchResultsResolver, error) {
		f := <-final
		return f.resultsResolver, f.err
	}
}

type searchResolver interface {
	Results(context.Context) (*graphqlbackend.SearchResultsResolver, error)
	Inputs() run.SearchInputs
}

func defaultNewSearchResolver(ctx context.Context, db dbutil.DB, args *graphqlbackend.SearchArgs) (searchResolver, error) {
	return graphqlbackend.NewSearchImplementer(ctx, db, args)
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

func fromStrPtr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// withDecoration hydrates event match with decorated hunks for a corresponding file match.
func withDecoration(ctx context.Context, eventMatch streamhttp.EventMatch, internalResult result.Match, kind string, contextLines int) streamhttp.EventMatch {
	if _, ok := internalResult.(*result.FileMatch); !ok {
		return eventMatch
	}

	event, ok := eventMatch.(*streamhttp.EventContentMatch)
	if !ok {
		return eventMatch
	}

	if kind == "html" {
		event.Hunks = DecorateFileHunksHTML(ctx, internalResult.(*result.FileMatch))
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
	} else if len(fm.LineMatches) > 0 {
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
	lineMatches := make([]streamhttp.EventLineMatch, 0, len(fm.LineMatches))
	for _, lm := range fm.LineMatches {
		lineMatches = append(lineMatches, streamhttp.EventLineMatch{
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
		LineMatches:  lineMatches,
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
	content := commit.Body.Value

	highlights := commit.Body.Highlights
	ranges := make([][3]int32, len(highlights))
	for i, h := range highlights {
		ranges[i] = [3]int32{h.Line, h.Character, h.Length}
	}

	commitEvent := &streamhttp.EventCommitMatch{
		Type:       streamhttp.CommitMatchType,
		Label:      commit.Label(),
		URL:        commit.URL().String(),
		Detail:     commit.Detail(),
		Repository: string(commit.Repo.Name),
		Content:    content,
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
	Buckets: []float64{0.01, 0.02, 0.05, 0.1, 0.2, 0.5, 1, 2, 5, 10, 30},
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

	if userAgent == "sourcegraph/query-runner" {
		return trace.SourceQueryRunner
	}

	return trace.SourceOther
}

// batchEvents takes an event stream and merges events that come through close in time into a single event.
// This makes downstream database and network operations more efficient by enabling batch reads.
func batchEvents(source <-chan streaming.SearchEvent, delay time.Duration) <-chan streaming.SearchEvent {
	results := make(chan streaming.SearchEvent)
	go func() {
		defer close(results)

		// Send the first event without a delay
		firstEvent, ok := <-source
		if !ok {
			return
		}
		results <- firstEvent

	OUTER:
		for {
			// Wait for a first event
			event, ok := <-source
			if !ok {
				return
			}

			// Wait up to the delay for more events to come through,
			// and merge any that do into the first event
			timer := time.After(delay)
			for {
				select {
				case newEvent, ok := <-source:
					if !ok {
						// Flush the buffered event and exit
						results <- event
						return
					}
					event.Results = append(event.Results, newEvent.Results...)
					event.Stats.Update(&newEvent.Stats)
				case <-timer:
					results <- event
					continue OUTER
				}
			}
		}

	}()
	return results
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

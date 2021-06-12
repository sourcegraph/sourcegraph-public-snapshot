// package search is search specific logic for the frontend. Also see
// github.com/sourcegraph/sourcegraph/internal/search for more generic search
// code.
package search

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	searchlogs "github.com/sourcegraph/sourcegraph/cmd/frontend/internal/search/logs"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
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
		trace.Tag{Key: "version_context", Value: args.VersionContext},
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

	traceURL := ""
	if span := opentracing.SpanFromContext(ctx); span != nil {
		spanURL := trace.SpanURL(span)
		// URLs starting with # don't have a trace. eg
		// "#tracer-not-enabled"
		if !strings.HasPrefix(spanURL, "#") {
			traceURL = spanURL
		}
	}

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
		Trace:        traceURL,
		DisplayLimit: displayLimit,
	}

	sendProgress := func() {
		_ = eventWriter.Event("progress", progress.Current())
	}

	filters := &streaming.SearchFilters{
		Globbing: false, // TODO
	}

	// Store marshalled matches and flush periodically or when we go over
	// 32kb.
	matchesBuf := &jsonArrayBuf{
		// 32kb chosen to be smaller than bufio.MaxTokenSize. Note: we can
		// still write more than that.
		FlushSize: 32 * 1024,
		Write: func(data []byte) error {
			return eventWriter.EventBytes("matches", data)
		},
	}
	matchesFlush := func() {
		if err := matchesBuf.Flush(); err != nil {
			// EOF
			return
		}

		if progress.Dirty {
			sendProgress()
		}
	}
	matchesAppend := func(m streamhttp.EventMatch) {
		// Only possible error is EOF, ignore
		_ = matchesBuf.Append(m)
	}

	repoCache := make(map[api.RepoID]*types.Repo, 10)

	flushTicker := time.NewTicker(h.flushTickerInternal)
	defer flushTicker.Stop()

	pingTicker := time.NewTicker(h.pingTickerInterval)
	defer pingTicker.Stop()

	first := true

	for {
		var event streaming.SearchEvent
		var ok bool
		select {
		case event, ok = <-events:
		case <-flushTicker.C:
			ok = true
			matchesFlush()
		case <-pingTicker.C:
			ok = true
			sendProgress()
		}

		if !ok {
			break
		}

		progress.Update(event)
		filters.Update(event)

		// Ensure that, for each result in the event, we have a copy of the full repo metadata.
		if err := updateRepoCache(ctx, h.db, repoCache, event); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		for _, match := range event.Results {
			if display <= 0 {
				break
			}

			display = match.Limit(display)
			matchesAppend(fromMatch(match, repoCache))
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
		})

		if honey.Enabled() {
			_ = ev.Send()
		}

		if isSlow {
			log15.Warn("streaming: slow search request", searchlogs.MapToLog15Ctx(ev.Fields())...)
		}
	}
}

// updateRepoCache requests repo metadata for each repo referenced by results in the event. It only makes a database request if there
// are new repos that do not already exist in the cache.
func updateRepoCache(ctx context.Context, db dbutil.DB, repoCache map[api.RepoID]*types.Repo, event streaming.SearchEvent) error {
	uncachedIDs := make(map[api.RepoID]struct{}, 10)
	for _, r := range event.Results {
		var id api.RepoID
		switch v := r.(type) {
		case *result.FileMatch:
			id = v.Repo.ID
		case *result.RepoMatch:
			id = v.ID
		case *result.CommitMatch:
			id = v.RepoName.ID
		default:
			return fmt.Errorf("unknown match type %T", r)
		}

		if _, ok := repoCache[id]; !ok {
			uncachedIDs[id] = struct{}{}
		}
	}

	if len(uncachedIDs) == 0 {
		return nil
	}

	uncachedIDSlice := make([]api.RepoID, 0, len(uncachedIDs))
	for id := range uncachedIDs {
		uncachedIDSlice = append(uncachedIDSlice, id)
	}

	repos, err := database.Repos(db).GetReposSetByIDs(ctx, uncachedIDSlice...)
	if err != nil {
		return err
	}
	for id, repo := range repos {
		repoCache[id] = repo
	}
	return nil
}

// startSearch will start a search. It returns the events channel which
// streams out search events. Once events is closed you can call results which
// will return the results resolver and error.
func (h *streamHandler) startSearch(ctx context.Context, a *args) (events <-chan streaming.SearchEvent, inputs run.SearchInputs, results func() (*graphqlbackend.SearchResultsResolver, error)) {
	eventsC := make(chan streaming.SearchEvent)

	search, err := h.newSearchResolver(ctx, h.db, &graphqlbackend.SearchArgs{
		Query:          a.Query,
		Version:        a.Version,
		PatternType:    strPtr(a.PatternType),
		VersionContext: strPtr(a.VersionContext),

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
	Query          string
	Version        string
	PatternType    string
	VersionContext string
	Display        int
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
		VersionContext: get("vc", ""),
	}

	if a.Query == "" {
		return nil, errors.New("no query found")
	}

	display := get("display", "-1")
	var err error
	if a.Display, err = strconv.Atoi(display); err != nil {
		return nil, fmt.Errorf("display must be an integer, got %q: %w", display, err)
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

func fromMatch(match result.Match, repoCache map[api.RepoID]*types.Repo) streamhttp.EventMatch {
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

func fromFileMatch(fm *result.FileMatch, repoCache map[api.RepoID]*types.Repo) streamhttp.EventMatch {
	if syms := fm.Symbols; len(syms) > 0 {
		return fromSymbolMatch(fm)
	}

	lineMatches := make([]streamhttp.EventLineMatch, 0, len(fm.LineMatches))
	for _, lm := range fm.LineMatches {
		lineMatches = append(lineMatches, streamhttp.EventLineMatch{
			Line:             lm.Preview,
			LineNumber:       lm.LineNumber,
			OffsetAndLengths: lm.OffsetAndLengths,
		})
	}

	var branches []string
	if fm.InputRev != nil {
		branches = []string{*fm.InputRev}
	}

	var stars int
	if r, ok := repoCache[fm.Repo.ID]; ok {
		stars = r.Stars
	}

	return &streamhttp.EventFileMatch{
		Type:        streamhttp.FileMatchType,
		Path:        fm.Path,
		Repository:  string(fm.Repo.Name),
		RepoStars:   stars,
		Branches:    branches,
		Version:     string(fm.CommitID),
		LineMatches: lineMatches,
	}
}

func fromSymbolMatch(fm *result.FileMatch) *streamhttp.EventSymbolMatch {
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

	var branches []string
	if fm.InputRev != nil {
		branches = []string{*fm.InputRev}
	}

	return &streamhttp.EventSymbolMatch{
		Type:       streamhttp.SymbolMatchType,
		Path:       fm.Path,
		Repository: string(fm.Repo.Name),
		Branches:   branches,
		Version:    string(fm.CommitID),
		Symbols:    symbols,
	}
}

func fromRepository(rm *result.RepoMatch, repoCache map[api.RepoID]*types.Repo) *streamhttp.EventRepoMatch {
	var branches []string
	if rev := rm.Rev; rev != "" {
		branches = []string{rev}
	}

	repoEvent := &streamhttp.EventRepoMatch{
		Type:       streamhttp.RepoMatchType,
		Repository: string(rm.Name),
		Branches:   branches,
	}

	if r, ok := repoCache[rm.ID]; ok {
		repoEvent.RepoStars = r.Stars
		repoEvent.Description = r.Description
		repoEvent.Fork = r.Fork
		repoEvent.Archived = r.Archived
	}

	return repoEvent
}

func fromCommit(commit *result.CommitMatch, repoCache map[api.RepoID]*types.Repo) *streamhttp.EventCommitMatch {
	content := commit.Body.Value

	highlights := commit.Body.Highlights
	ranges := make([][3]int32, len(highlights))
	for i, h := range highlights {
		ranges[i] = [3]int32{h.Line, h.Character, h.Length}
	}

	var stars int
	if r, ok := repoCache[commit.RepoName.ID]; ok {
		stars = r.Stars
	}

	return &streamhttp.EventCommitMatch{
		Type:       streamhttp.CommitMatchType,
		Label:      commit.Label(),
		URL:        commit.URL().String(),
		Detail:     commit.Detail(),
		Repository: string(commit.RepoName.Name),
		RepoStars:  stars,
		Content:    content,
		Ranges:     ranges,
	}
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

// jsonArrayBuf builds up a JSON array by marshalling per item. Once the array
// has reached FlushSize it will be written out via Write and the buffer will
// be reset.
type jsonArrayBuf struct {
	FlushSize int
	Write     func([]byte) error

	buf bytes.Buffer
}

// Append marshals v and adds it to the json array buffer. If the size of the
// buffer exceed FlushSize the buffer is written out.
func (j *jsonArrayBuf) Append(v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}

	if j.buf.Len() == 0 {
		j.buf.WriteByte('[')
	} else {
		j.buf.WriteByte(',')
	}

	// err is always nil for a bytes.Buffer
	_, _ = j.buf.Write(b)

	if j.buf.Len() >= j.FlushSize {
		return j.Flush()
	}
	return nil
}

// Flush writes and resets the buffer if there is data to write.
func (j *jsonArrayBuf) Flush() error {
	if j.buf.Len() == 0 {
		return nil
	}

	// Terminate array
	j.buf.WriteByte(']')

	buf := j.buf.Bytes()
	j.buf.Reset()
	return j.Write(buf)
}

func (j *jsonArrayBuf) Len() int {
	return j.buf.Len()
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

	return trace.SourceOther
}

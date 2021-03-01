// package search is search specific logic for the frontend. Also see
// github.com/sourcegraph/sourcegraph/internal/search for more generic search
// code.
package search

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

// StreamHandler is an http handler which streams back search results.
func StreamHandler(db dbutil.DB) http.Handler {
	return &streamHandler{
		db:                db,
		newSearchResolver: defaultNewSearchResolver,
	}
}

type streamHandler struct {
	db                dbutil.DB
	newSearchResolver func(context.Context, dbutil.DB, *graphqlbackend.SearchArgs) (searchResolver, error)
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

	eventWriter, err := newEventStreamWriter(w)
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

	progress := progressAggregator{
		Start: time.Now(),
		Limit: inputs.MaxResults(),
	}

	// Display is the number of results we send down. If display is < 0 we
	// want to send everything we find before hitting a limit. Otherwise we
	// can only send up to limit results.
	display := args.Display
	if limit := inputs.MaxResults(); display < 0 || display > limit {
		display = limit
	}

	sendProgress := func() {
		_ = eventWriter.Event("progress", progress.Current())
	}

	filters := &graphqlbackend.SearchFilters{
		Globbing: false, // TODO
	}

	const matchesChunk = 1000
	matchesBuf := make([]interface{}, 0, matchesChunk)
	flushMatchesBuf := func() {
		if len(matchesBuf) > 0 {
			if err := eventWriter.Event("matches", matchesBuf); err != nil {
				// EOF
				return
			}
			matchesBuf = matchesBuf[:0]
		}

		if progress.Dirty {
			sendProgress()
		}
	}

	flushTicker := time.NewTicker(100 * time.Millisecond)
	defer flushTicker.Stop()

	pingTicker := time.NewTicker(5 * time.Second)
	defer pingTicker.Stop()

	first := true

	for {
		var event graphqlbackend.SearchEvent
		var ok bool
		select {
		case event, ok = <-events:
		case <-flushTicker.C:
			ok = true
			flushMatchesBuf()
		case <-pingTicker.C:
			ok = true
			sendProgress()
		}

		if !ok {
			break
		}

		progress.Update(event)
		filters.Update(event)

		for _, result := range event.Results {
			if display <= 0 {
				break
			}

			if fm, ok := result.ToFileMatch(); ok {
				display = fm.Limit(display)

				if syms := fm.Symbols(); len(syms) > 0 {
					// Inlining to avoid exporting a bunch of stuff from
					// graphqlbackend
					symbols := make([]symbol, 0, len(syms))
					for _, sym := range syms {
						u, err := sym.URL(ctx)
						if err != nil {
							continue
						}
						symbols = append(symbols, symbol{
							URL:           u,
							Name:          sym.Name(),
							ContainerName: fromStrPtr(sym.ContainerName()),
							Kind:          sym.Kind(),
						})
					}
					matchesBuf = append(matchesBuf, fromSymbolMatch(fm, symbols))
				} else {
					matchesBuf = append(matchesBuf, fromFileMatch(&fm.FileMatch))
				}
			}
			if repo, ok := result.ToRepository(); ok {
				display = repo.Limit(display)

				matchesBuf = append(matchesBuf, fromRepository(repo))
			}
			if commit, ok := result.ToCommitSearchResult(); ok {
				display = commit.Limit(display)

				matchesBuf = append(matchesBuf, fromCommit(commit))
			}
			if len(matchesBuf) == cap(matchesBuf) {
				flushMatchesBuf()
			}
		}

		// Instantly send results if we have not sent any yet.
		if first && len(matchesBuf) > 0 {
			first = false
			flushMatchesBuf()
		}
	}

	flushMatchesBuf()

	// Send dynamic filters once.
	if filters := filters.Compute(); len(filters) > 0 {
		buf := make([]eventFilter, 0, len(filters))
		for _, f := range filters {
			buf = append(buf, eventFilter{
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
		_ = eventWriter.Event("error", eventError{Message: err.Error()})
		return
	}

	if alert := resultsResolver.Alert(); alert != nil {
		var pqs []proposedQuery
		if proposed := alert.ProposedQueries(); proposed != nil {
			for _, pq := range *proposed {
				pqs = append(pqs, proposedQuery{
					Description: fromStrPtr(pq.Description()),
					Query:       pq.Query(),
				})
			}
		}
		_ = eventWriter.Event("alert", eventAlert{
			Title:           alert.Title(),
			Description:     fromStrPtr(alert.Description()),
			ProposedQueries: pqs,
		})
	}

	_ = eventWriter.Event("progress", progress.Final())
}

// startSearch will start a search. It returns the events channel which
// streams out search events. Once events is closed you can call results which
// will return the results resolver and error.
func (h *streamHandler) startSearch(ctx context.Context, a *args) (events <-chan graphqlbackend.SearchEvent, inputs graphqlbackend.SearchInputs, results func() (*graphqlbackend.SearchResultsResolver, error)) {
	eventsC := make(chan graphqlbackend.SearchEvent)

	search, err := h.newSearchResolver(ctx, h.db, &graphqlbackend.SearchArgs{
		Query:          a.Query,
		Version:        a.Version,
		PatternType:    strPtr(a.PatternType),
		VersionContext: strPtr(a.VersionContext),

		Stream: graphqlbackend.StreamFunc(func(event graphqlbackend.SearchEvent) {
			eventsC <- event
		}),
	})
	if err != nil {
		close(eventsC)
		return eventsC, graphqlbackend.SearchInputs{}, func() (*graphqlbackend.SearchResultsResolver, error) {
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
	Inputs() graphqlbackend.SearchInputs
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
		PatternType:    get("t", "literal"),
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

func fromFileMatch(fm *graphqlbackend.FileMatch) eventFileMatch {
	lineMatches := make([]eventLineMatch, 0, len(fm.LineMatches))
	for _, lm := range fm.LineMatches {
		lineMatches = append(lineMatches, eventLineMatch{
			Line:             lm.Preview,
			LineNumber:       lm.LineNumber,
			OffsetAndLengths: lm.OffsetAndLengths,
		})
	}

	var branches []string
	if fm.InputRev != nil {
		branches = []string{*fm.InputRev}
	}

	return eventFileMatch{
		Type:        fileMatch,
		Path:        fm.Path,
		Repository:  string(fm.Repo.Name),
		Branches:    branches,
		Version:     string(fm.CommitID),
		LineMatches: lineMatches,
	}
}

func fromSymbolMatch(fm *graphqlbackend.FileMatchResolver, symbols []symbol) eventSymbolMatch {
	var branches []string
	if fm.InputRev != nil {
		branches = []string{*fm.InputRev}
	}

	return eventSymbolMatch{
		Type:       symbolMatch,
		Path:       fm.Path,
		Repository: string(fm.Repo.Name),
		Branches:   branches,
		Version:    string(fm.CommitID),
		Symbols:    symbols,
	}
}

func fromRepository(repo *graphqlbackend.RepositoryResolver) eventRepoMatch {
	var branches []string
	if rev := repo.Rev(); rev != "" {
		branches = []string{rev}
	}

	return eventRepoMatch{
		Type:       repoMatch,
		Repository: repo.Name(),
		Branches:   branches,
	}
}

func fromCommit(commit *graphqlbackend.CommitSearchResultResolver) eventCommitMatch {
	var content string
	var ranges [][3]int32
	if matches := commit.Matches(); len(matches) == 1 {
		match := matches[0]
		content = match.Body().Text()
		highlights := match.Highlights()
		ranges = make([][3]int32, len(highlights))
		for i, h := range highlights {
			ranges[i] = [3]int32{h.Line(), h.Character(), h.Length()}
		}
	}
	return eventCommitMatch{
		Type:    commitMatch,
		Icon:    commit.Icon(),
		Label:   commit.Label().Text(),
		URL:     commit.URL(),
		Detail:  commit.Detail().Text(),
		Content: content,
		Ranges:  ranges,
	}
}

// eventFileMatch is a subset of zoekt.FileMatch for our event API.
type eventFileMatch struct {
	// Type is always fileMatch. Included here for marshalling.
	Type matchType `json:"type"`

	Path       string   `json:"name"`
	Repository string   `json:"repository"`
	Branches   []string `json:"branches,omitempty"`
	Version    string   `json:"version,omitempty"`

	LineMatches []eventLineMatch `json:"lineMatches"`
}

// eventLineMatch is a subset of zoekt.LineMatch for our event API.
type eventLineMatch struct {
	Line             string     `json:"line"`
	LineNumber       int32      `json:"lineNumber"`
	OffsetAndLengths [][2]int32 `json:"offsetAndLengths"`
}

// eventRepoMatch is a subset of zoekt.FileMatch for our event API.
type eventRepoMatch struct {
	// Type is always repoMatch. Included here for marshalling.
	Type matchType `json:"type"`

	Repository string   `json:"repository"`
	Branches   []string `json:"branches,omitempty"`
}

// eventSymbolMatch is eventFileMatch but with Symbols instead of LineMatches
type eventSymbolMatch struct {
	// Type is always symbolMatch. Included here for marshalling.
	Type matchType `json:"type"`

	Path       string   `json:"name"`
	Repository string   `json:"repository"`
	Branches   []string `json:"branches,omitempty"`
	Version    string   `json:"version,omitempty"`

	Symbols []symbol `json:"symbols"`
}

type symbol struct {
	URL           string `json:"url"`
	Name          string `json:"name"`
	ContainerName string `json:"containerName"`
	Kind          string `json:"kind"`
}

// eventCommitMatch is the generic results interface from GQL. There is a lot
// of potential data that may be useful here, and some thought needs to be put
// into what is actually useful in a commit result / or if we should have a
// "type" for that.
type eventCommitMatch struct {
	// Type is always commitMatch. Included here for marshalling.
	Type matchType `json:"type"`

	Icon    string `json:"icon"`
	Label   string `json:"label"`
	URL     string `json:"url"`
	Detail  string `json:"detail"`
	Content string `json:"content"`
	// [line, character, length]
	Ranges [][3]int32 `json:"ranges"`
}

// eventFilter is a suggestion for a search filter. Currently has a 1-1
// correspondance with the SearchFilter graphql type.
type eventFilter struct {
	Value    string `json:"value"`
	Label    string `json:"label"`
	Count    int    `json:"count"`
	LimitHit bool   `json:"limitHit"`
	Kind     string `json:"kind"`
}

// eventAlert is GQL.SearchAlert. It replaces when sent to match existing
// behaviour.
type eventAlert struct {
	Title           string          `json:"title"`
	Description     string          `json:"description,omitempty"`
	ProposedQueries []proposedQuery `json:"proposedQueries"`
}

// proposedQuery is a suggested query to run when we emit an alert.
type proposedQuery struct {
	Description string `json:"description,omitempty"`
	Query       string `json:"query"`
}

// eventError emulates a JavaScript error with a message property
// as is returned when the search encounters an error.
type eventError struct {
	Message string `json:"message"`
}

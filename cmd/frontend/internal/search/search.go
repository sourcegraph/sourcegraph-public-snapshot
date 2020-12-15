// package search is search specific logic for the frontend. Also see
// github.com/sourcegraph/sourcegraph/internal/search for more generic search
// code.
package search

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

// StreamHandler is an http handler which streams back search results.
var StreamHandler http.Handler = &streamHandler{
	newSearchResolver: defaultNewSearchResolver,
}

type streamHandler struct {
	newSearchResolver func(context.Context, *graphqlbackend.SearchArgs) (searchResolver, error)
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

	// Log events to trace
	eventWriter.StatHook = eventStreamOTHook(tr.LogFields)

	search, err := h.newSearchResolver(ctx, &graphqlbackend.SearchArgs{
		Query:          args.Query,
		Version:        args.Version,
		PatternType:    strPtr(args.PatternType),
		VersionContext: strPtr(args.VersionContext),
	})
	if err != nil {
		_ = eventWriter.Event("error", err.Error())
		return
	}

	resultsStream, resultsStreamDone := newResultsStream(ctx, search)

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
	}

	flushTicker := time.NewTicker(100 * time.Millisecond)
	defer flushTicker.Stop()

	for {
		var results []graphqlbackend.SearchResultResolver
		var ok bool
		select {
		case results, ok = <-resultsStream:
		case <-flushTicker.C:
			ok = true
			flushMatchesBuf()
		}

		if !ok {
			break
		}

		for _, result := range results {
			if fm, ok := result.ToFileMatch(); ok {
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
					matchesBuf = append(matchesBuf, fromFileMatch(fm))
				}
			}
			if repo, ok := result.ToRepository(); ok {
				matchesBuf = append(matchesBuf, fromRepository(repo))
			}
			if commit, ok := result.ToCommitSearchResult(); ok {
				matchesBuf = append(matchesBuf, fromCommit(commit))
			}
			if len(matchesBuf) == cap(matchesBuf) {
				flushMatchesBuf()
			}
		}
	}

	flushMatchesBuf()

	final := <-resultsStreamDone
	resultsResolver, err := final.resultsResolver, final.err
	if err != nil {
		_ = eventWriter.Event("error", err.Error())
		return
	}

	// Send dynamic filters once. When this is true streaming we may want to
	// send updated filters as we find more results.
	if filters := resultsResolver.DynamicFilters(ctx); len(filters) > 0 {
		buf := make([]eventFilter, 0, len(filters))
		for _, f := range filters {
			buf = append(buf, eventFilter{
				Value:    f.Value(),
				Label:    f.Label(),
				Count:    int(f.Count()),
				LimitHit: f.LimitHit(),
				Kind:     f.Kind(),
			})
		}

		if err := eventWriter.Event("filters", buf); err != nil {
			// EOF
			return
		}
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

	pr := resultsResolver.Progress()
	pr.Done = true
	_ = eventWriter.Event("progress", pr)

	// TODO done event includes progress
	_ = eventWriter.Event("done", map[string]interface{}{})
}

type searchResolver interface {
	Results(context.Context) (*graphqlbackend.SearchResultsResolver, error)
	SetResultChannel(c chan<- []graphqlbackend.SearchResultResolver)
}

func defaultNewSearchResolver(ctx context.Context, args *graphqlbackend.SearchArgs) (searchResolver, error) {
	searchImpl, err := graphqlbackend.NewSearchImplementer(ctx, args)
	if err != nil {
		return nil, err
	}
	search, ok := searchImpl.(searchResolver)
	if !ok {
		return nil, errors.New("SearchImplementer does not support streaming")
	}
	return search, nil
}

type args struct {
	Query          string
	Version        string
	PatternType    string
	VersionContext string
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

type finalResult struct {
	resultsResolver *graphqlbackend.SearchResultsResolver
	err             error
}

// newResultsStream will return a channel which streams out the results found
// by search.Results. Once done it will send the return value of
// search.Results. Both channels need to be read until closed, otherwise
// goroutines will be leaked.
//
//   - results is written to 0 or more times before closing.
//   - final is written to once.
func newResultsStream(ctx context.Context, search searchResolver) (results <-chan []graphqlbackend.SearchResultResolver, final <-chan finalResult) {
	resultsC := make(chan []graphqlbackend.SearchResultResolver)
	finalC := make(chan finalResult, 1)
	go func() {
		defer close(finalC)
		defer close(resultsC)

		search.SetResultChannel(resultsC)

		r, err := search.Results(ctx)
		finalC <- finalResult{resultsResolver: r, err: err}
	}()
	return resultsC, finalC
}

func fromFileMatch(fm *graphqlbackend.FileMatchResolver) eventFileMatch {
	lineMatches := make([]eventLineMatch, 0, len(fm.JLineMatches))
	for _, lm := range fm.JLineMatches {
		lineMatches = append(lineMatches, eventLineMatch{
			Line:             lm.JPreview,
			LineNumber:       lm.JLineNumber,
			OffsetAndLengths: lm.JOffsetAndLengths,
		})
	}

	var branches []string
	if fm.InputRev != nil {
		branches = []string{*fm.InputRev}
	}

	return eventFileMatch{
		Type:        fileMatch,
		Path:        fm.JPath,
		Repository:  fm.Repo.Name(),
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
		Path:       fm.JPath,
		Repository: fm.Repo.Name(),
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

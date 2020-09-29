// package search is search specific logic for the frontend. Also see
// github.com/sourcegraph/sourcegraph/internal/search for more generic search
// code.
package search

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

// ServeStream is an http handler which streams back search results.
func ServeStream(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	qvals := r.URL.Query()
	queryStr := qvals.Get("q")
	if queryStr == "" {
		http.Error(w, "no query found", http.StatusBadRequest)
		return
	}

	eventWriter, err := newEventStreamWriter(w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	search, err := graphqlbackend.NewSearchImplementer(ctx, &graphqlbackend.SearchArgs{
		Query:   queryStr,
		Version: "V2",
	})
	if err != nil {
		eventWriter.Event("error", err.Error())
		return
	}

	resultsResolver, err := search.Results(ctx)
	if err != nil {
		eventWriter.Event("error", err.Error())
		return
	}

	const filematchesChunk = 1000
	filematchesBuf := make([]eventFileMatch, 0, filematchesChunk)

	for _, result := range resultsResolver.Results() {
		fm, ok := result.ToFileMatch()
		if !ok {
			continue
		}

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

		filematchesBuf = append(filematchesBuf, eventFileMatch{
			Path:        fm.JPath,
			Repository:  fm.Repo.Name(),
			Branches:    branches,
			Version:     string(fm.CommitID),
			LineMatches: lineMatches,
		})

		if len(filematchesBuf) == cap(filematchesBuf) {
			if err := eventWriter.Event("filematches", filematchesBuf); err != nil {
				// EOF
				return
			}
			filematchesBuf = filematchesBuf[:0]
		}
	}

	if len(filematchesBuf) > 0 {
		if err := eventWriter.Event("filematches", filematchesBuf); err != nil {
			// EOF
			return
		}
		filematchesBuf = filematchesBuf[:0]
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

	// TODO stats
	_ = eventWriter.Event("done", map[string]interface{}{})
}

type eventStreamWriter struct {
	w     io.Writer
	enc   *json.Encoder
	flush func()
}

func newEventStreamWriter(w http.ResponseWriter) (*eventStreamWriter, error) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, errors.New("http flushing not supported")
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	return &eventStreamWriter{
		w:     w,
		enc:   json.NewEncoder(w),
		flush: flusher.Flush,
	}, nil
}

func (e *eventStreamWriter) Event(event string, data interface{}) error {
	if event != "" {
		// event: $event\n
		if _, err := e.w.Write([]byte("event: ")); err != nil {
			return err
		}
		if _, err := e.w.Write([]byte(event)); err != nil {
			return err
		}
		if _, err := e.w.Write([]byte("\n")); err != nil {
			return err
		}
	}

	// data: json(data)\n\n
	if _, err := e.w.Write([]byte("data: ")); err != nil {
		return err
	}
	if err := e.enc.Encode(data); err != nil {
		return err
	}
	// Encode writes a newline, so only need to write one newline.
	if _, err := e.w.Write([]byte("\n")); err != nil {
		return err
	}

	e.flush()

	return nil
}

// eventFileMatch is a subset of zoekt.FileMatch for our event API.
type eventFileMatch struct {
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

// eventFilter is a suggestion for a search filter. Currently has a 1-1
// correspondance with the SearchFilter graphql type.
type eventFilter struct {
	Value    string `json:"value"`
	Label    string `json:"label"`
	Count    int    `json:"count"`
	LimitHit bool   `json:"limitHit"`
	Kind     string `json:"kind"`
}

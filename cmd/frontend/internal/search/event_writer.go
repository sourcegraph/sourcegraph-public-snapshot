package search

import (
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming/api"
	streamhttp "github.com/sourcegraph/sourcegraph/internal/search/streaming/http"
)

func newEventWriter(inner *streamhttp.Writer) *eventWriter {
	return &eventWriter{inner: inner}
}

// eventWriter is a type that wraps a streamhttp.Writer with typed
// methods for each of the supported evens in a frontend stream.
type eventWriter struct {
	inner *streamhttp.Writer
}

func (e *eventWriter) Done() error {
	return e.inner.Event("done", map[string]any{})
}

func (e *eventWriter) Progress(current api.Progress) error {
	return e.inner.Event("progress", current)
}

func (e *eventWriter) MatchesJSON(data []byte) error {
	return e.inner.EventBytes("matches", data)
}

func (e *eventWriter) Filters(fs []*streaming.Filter) error {
	if len(fs) > 0 {
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

		return e.inner.Event("filters", buf)
	}
	return nil
}

func (e *eventWriter) Error(err error) error {
	return e.inner.Event("error", streamhttp.EventError{Message: err.Error()})
}

func (e *eventWriter) Alert(alert *search.Alert) error {
	var pqs []streamhttp.QueryDescription
	for _, pq := range alert.ProposedQueries {
		annotations := make([]streamhttp.Annotation, 0, len(pq.Annotations))
		for name, value := range pq.Annotations {
			annotations = append(annotations, streamhttp.Annotation{Name: string(name), Value: value})
		}

		pqs = append(pqs, streamhttp.QueryDescription{
			Description: pq.Description,
			Query:       pq.QueryString(),
			Annotations: annotations,
		})
	}
	return e.inner.Event("alert", streamhttp.EventAlert{
		Title:           alert.Title,
		Description:     alert.Description,
		Kind:            alert.Kind,
		ProposedQueries: pqs,
	})
}

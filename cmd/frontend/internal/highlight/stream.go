package highlight

import (
	"net/http"

	"github.com/sourcegraph/scip/bindings/go/scip"

	streamhttp "github.com/sourcegraph/sourcegraph/internal/search/streaming/http"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

func StreamHandler() http.Handler {
	return &streamHandler{}
}

type streamHandler struct {
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
	eventWriter := newEventWriter(streamWriter)
	defer eventWriter.Done()
}

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
func (e *eventWriter) Plaintext(text string) error {
	return e.inner.Event("plaintext", text)
}

func (e *eventWriter) ScipDocument(document scip.Document) error {
	return e.inner.Event("scip-document", document)
}

package httpapi

import (
	"fmt"
	"net/http"
	"time"

	"sourcegraph.com/sourcegraph/appdash"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/httputil/httpctx"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/traceutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/traceutil/appdashctx"
)

type PageLoadEvent struct {
	S, E time.Time

	// Name of the event.
	Name string
}

// Schema implements the appdash.Event interface.
func (e PageLoadEvent) Schema() string { return "PageLoad" }

// Start implements the appdash.TimespanEvent interface.
func (e PageLoadEvent) Start() time.Time { return e.S }

// End implements the appdash.TimespanEvent interface.
func (e PageLoadEvent) End() time.Time { return e.E }

func init() {
	appdash.RegisterEvent(PageLoadEvent{})
}

// serveInternalAppdashRecordSpan is an endpoint that records a very simple
// span with a name and duration as a child of the trace root.
//
// This mostly works around the fact that Appdash does not support JavaScript
// tracing yet.
func serveInternalAppdashRecordSpan(w http.ResponseWriter, r *http.Request) error {
	ctx := httpctx.FromRequest(r)

	// Decode query parameters into an event.
	ev := &PageLoadEvent{}
	if err := schemaDecoder.Decode(ev, r.URL.Query()); err != nil {
		return err
	}

	// The `internal.appdash.record-span` span for this POST request is tiny
	// and thus not easily accessible within the UI. We use a workaround by
	// attaching the span we will generate to the parent of this POST request
	// span. I.e. they are sublings.

	// Grab the collector from the context.
	collector := appdashctx.Collector(ctx)
	if collector == nil {
		return fmt.Errorf("no Appdash collector set in context")
	}

	// Grab the SpanID from the context.
	spanID := traceutil.SpanIDFromContext(ctx)
	if spanID.Trace == 0 {
		return fmt.Errorf("no Appdash trace ID set in context")
	}

	newSpan := appdash.NewSpanID(appdash.SpanID{
		Trace: spanID.Trace,

		// newSpan.Parent will be this span, so set it to this POST request
		// span's parent span ID so we become a sibling.
		Span: spanID.Parent,
	})
	rec := traceutil.NewRecorder(newSpan, collector)
	rec.Name(fmt.Sprintf("Browser %s", ev.Name))
	rec.Event(ev)
	return nil
}

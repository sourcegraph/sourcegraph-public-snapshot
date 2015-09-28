package ui

import (
	"net/http"

	"fmt"
	"time"

	"sourcegraph.com/sourcegraph/appdash"
	"sourcegraph.com/sourcegraph/sourcegraph/util/httputil/httpctx"
	"sourcegraph.com/sourcegraph/sourcegraph/util/traceutil"
	"sourcegraph.com/sourcegraph/sourcegraph/util/traceutil/appdashctx"
)

type PageLoadEvent struct {
	S, E time.Time
}

// Schema implements the appdash.Event interface.
func (e PageLoadEvent) Schema() string { return "PageLoad" }

// Start implements the appdash.TimespanEvent interface.
func (e PageLoadEvent) Start() time.Time { return e.S }

// End implements the appdash.TimespanEvent interface.
func (e PageLoadEvent) End() time.Time { return e.E }

func init() { appdash.RegisterEvent(PageLoadEvent{}) }

// serveAppdashUploadPageLoad is an endpoint that simply generates a 'fake'
// PageLoadEvent Appdash timespan event to represent how long exactly
// the frontend took to load everything. The client is responsible for
// determining the start and end times (we just generate the event because
// JavaScript can't record Appdash events yet).
func serveAppdashUploadPageLoad(w http.ResponseWriter, r *http.Request) error {
	ctx := httpctx.FromRequest(r)

	// Decode query parameters into an event.
	ev := &PageLoadEvent{}
	if err := schemaDecoder.Decode(ev, r.URL.Query()); err != nil {
		return err
	}

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

	// Note: If we were to record directly to spanID we would end up with
	// "PageLoad" being shown as a subspan to this request
	// ("appdash.upload-page-load") which is always extremely quick, making it's
	// display in the Appdash UI very small and unnoticeable. To mitigate this
	// and give it a prominent display position in the UI, we simply record to a
	// subspan of the root (the trace).
	newSpan := appdash.NewSpanID(appdash.SpanID{
		Trace: spanID.Trace,
		Span:  spanID.Trace,
	})
	rec := appdash.NewRecorder(newSpan, collector)
	rec.Name(ev.Schema())
	rec.Event(ev)
	if errs := rec.Errors(); len(errs) > 0 {
		return errs[0]
	}
	return nil
}

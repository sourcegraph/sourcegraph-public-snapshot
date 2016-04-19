package httpapi

import (
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"sourcegraph.com/sourcegraph/appdash"
	"sourcegraph.com/sourcegraph/sourcegraph/util/httputil/httpctx"
	"sourcegraph.com/sourcegraph/sourcegraph/util/traceutil"
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

var pageLoadLabels = []string{"name"}
var pageLoadDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Namespace: "src",
	Subsystem: "trace",
	Name:      "browser_span_duration_seconds",
	Help:      "Total time taken to perform a given browser operation.",
	Buckets:   []float64{1, 5, 10, 60, 300},
}, pageLoadLabels)

func init() {
	appdash.RegisterEvent(PageLoadEvent{})
	prometheus.MustRegister(pageLoadDuration)
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

	// Record page load duration in Prometheus histogram.
	labels := prometheus.Labels{
		"name": ev.Name,
	}
	elapsed := ev.E.Sub(ev.S)
	pageLoadDuration.With(labels).Observe(elapsed.Seconds())

	rec := traceutil.Recorder(ctx).Child()
	rec.Name(fmt.Sprintf("Browser %s", ev.Name))
	rec.Event(ev)
	return nil
}

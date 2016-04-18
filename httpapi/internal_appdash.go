package httpapi

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"sourcegraph.com/sourcegraph/appdash"
	"sourcegraph.com/sourcegraph/sourcegraph/util/httputil/httpctx"
	"sourcegraph.com/sourcegraph/sourcegraph/util/traceutil"
)

type PageLoadEvent struct {
	S, E time.Time

	// route and template name of the rendered page
	Route, Template string
}

// Schema implements the appdash.Event interface.
func (e PageLoadEvent) Schema() string { return "PageLoad" }

// Start implements the appdash.TimespanEvent interface.
func (e PageLoadEvent) Start() time.Time { return e.S }

// End implements the appdash.TimespanEvent interface.
func (e PageLoadEvent) End() time.Time { return e.E }

var pageLoadLabels = []string{"route", "template"}
var pageLoadDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Namespace: "src",
	Subsystem: "trace",
	Name:      "page_load_duration_seconds",
	Help:      "Total time taken to load the entire page.",
	Buckets:   []float64{1, 5, 10, 60, 300},
}, pageLoadLabels)

func init() {
	appdash.RegisterEvent(PageLoadEvent{})
	prometheus.MustRegister(pageLoadDuration)
}

// serveInternalAppdashUploadPageLoad is an endpoint that simply
// generates a 'fake' PageLoadEvent Appdash timespan event to
// represent how long exactly the frontend took to load
// everything. The client is responsible for determining the start and
// end times (we just generate the event because JavaScript can't
// record Appdash events yet).
func serveInternalAppdashUploadPageLoad(w http.ResponseWriter, r *http.Request) error {
	ctx := httpctx.FromRequest(r)

	// Decode query parameters into an event.
	ev := &PageLoadEvent{}
	if err := schemaDecoder.Decode(ev, r.URL.Query()); err != nil {
		return err
	}

	// Record page load duration in Prometheus histogram.
	labels := prometheus.Labels{
		"route":    ev.Route,
		"template": ev.Template,
	}
	elapsed := ev.E.Sub(ev.S)
	pageLoadDuration.With(labels).Observe(elapsed.Seconds())

	rec := traceutil.Recorder(ctx).Child()
	rec.Name(ev.Schema())
	rec.Event(ev)
	return nil
}

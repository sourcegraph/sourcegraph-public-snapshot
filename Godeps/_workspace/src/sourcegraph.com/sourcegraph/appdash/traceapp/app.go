// Package traceapp implements the Appdash web UI.
//
// The web UI can be effectively launched using the appdash command (see
// cmd/appdash) or via embedding this package within your app.
//
// Templates and other resources needed by this package to render the UI are
// built into the program using vfsgen, so you still get to have single
// binary deployment.
//
// For an example of embedding the Appdash web UI within your own application
// via the traceapp package, see the examples/cmd/webapp example.
package traceapp

import (
	"encoding/json"
	"errors"
	htmpl "html/template"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"sort"
	"strings"
	"sync"

	"github.com/gorilla/mux"

	"sourcegraph.com/sourcegraph/appdash"
	static "sourcegraph.com/sourcegraph/appdash-data"
)

// App is an HTTP application handler that also exposes methods for
// constructing URL routes.
type App struct {
	*Router

	Store   appdash.Store
	Queryer appdash.Queryer

	tmplLock sync.Mutex
	tmpls    map[string]*htmpl.Template
}

// New creates a new application handler. If r is nil, a new router is
// created.
func New(r *Router) *App {
	if r == nil {
		r = NewRouter(nil)
	}

	app := &App{Router: r}

	r.r.Get(RootRoute).Handler(handlerFunc(app.serveRoot))
	r.r.Get(TraceRoute).Handler(handlerFunc(app.serveTrace))
	r.r.Get(TraceSpanRoute).Handler(handlerFunc(app.serveTrace))
	r.r.Get(TraceProfileRoute).Handler(handlerFunc(app.serveTrace))
	r.r.Get(TraceSpanProfileRoute).Handler(handlerFunc(app.serveTrace))
	r.r.Get(TraceUploadRoute).Handler(handlerFunc(app.serveTraceUpload))
	r.r.Get(TracesRoute).Handler(handlerFunc(app.serveTraces))
	r.r.Get(DashboardRoute).Handler(handlerFunc(app.serveDashboard))
	r.r.Get(DashboardDataRoute).Handler(handlerFunc(app.serveDashboardData))
	r.r.Get(AggregateRoute).Handler(handlerFunc(app.serveAggregate))

	// Static file serving.
	r.r.Get(StaticRoute).Handler(http.StripPrefix("/static/", http.FileServer(static.Data)))

	return app
}

// ServeHTTP implements http.Handler.
func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.Router.r.ServeHTTP(w, r)
}

func (a *App) serveRoot(w http.ResponseWriter, r *http.Request) error {
	return a.renderTemplate(w, r, "root.html", http.StatusOK, &struct {
		TemplateCommon
	}{})
}

func (a *App) serveTrace(w http.ResponseWriter, r *http.Request) error {
	v := mux.Vars(r)

	traceID, err := appdash.ParseID(v["Trace"])
	if err != nil {
		return err
	}

	trace, err := a.Store.Trace(traceID)
	if err != nil {
		return err
	}

	// Get sub-span if the Span route var is present.
	if spanIDStr := v["Span"]; spanIDStr != "" {
		spanID, err := appdash.ParseID(spanIDStr)
		if err != nil {
			return err
		}
		trace = trace.FindSpan(spanID)
		if trace == nil {
			return errors.New("could not find the specified trace span")
		}
	}

	// We could use a separate handler for this, but as we need the above to
	// determine the correct trace (or therein sub-trace), we just handle any
	// JSON profile requests here.
	if path.Base(r.URL.Path) == "profile" {
		return a.profile(trace, w)
	}

	// Do not show d3 timeline chart when timeline item fields are invalid.
	// So we avoid JS code breaking due missing values.
	var showTimelineChart bool = true
	visData, err := a.d3timeline(trace)
	switch err {
	case errTimelineItemValidation:
		showTimelineChart = false
	case nil:
		break
	default:
		return err
	}

	// Determine the profile URL.
	var profile *url.URL
	if trace.ID.Parent == 0 {
		profile, err = a.Router.URLToTraceProfile(trace.Span.ID.Trace)
	} else {
		profile, err = a.Router.URLToTraceSpanProfile(trace.Span.ID.Trace, trace.Span.ID.Span)
	}
	if err != nil {
		return err
	}

	return a.renderTemplate(w, r, "trace.html", http.StatusOK, &struct {
		TemplateCommon
		Trace             *appdash.Trace
		ShowTimelineChart bool
		VisData           []timelineItem
		ProfileURL        string
	}{
		Trace:             trace,
		ShowTimelineChart: showTimelineChart,
		VisData:           visData,
		ProfileURL:        profile.String(),
	})
}

func (a *App) serveTraces(w http.ResponseWriter, r *http.Request) error {
	traces, err := a.Queryer.Traces()
	if err != nil {
		return err
	}

	// Parse the query for a comma-separated list of traces that we should only
	// show (all others are hidden).
	var showJust []appdash.ID
	if show := r.URL.Query().Get("show"); len(show) > 0 {
		for _, idStr := range strings.Split(show, ",") {
			id, err := appdash.ParseID(idStr)
			if err == nil {
				showJust = append(showJust, id)
			}
		}
	}

	// Sort the traces by ID to ensure that the display order doesn't change upon
	// multiple page reloads if Queryer.Traces is e.g. backed by a map (which has
	// a random iteration order).
	sort.Sort(tracesByID(traces))

	return a.renderTemplate(w, r, "traces.html", http.StatusOK, &struct {
		TemplateCommon
		Traces  []*appdash.Trace
		Visible func(*appdash.Trace) bool
	}{
		Traces: traces,
		Visible: func(t *appdash.Trace) bool {
			// Hide the traces that contain aggregate events (that's all they have, so
			// they are not very useful to users).
			if t.IsAggregate() {
				return false
			}

			if len(showJust) > 0 {
				// Showing just a few traces.
				for _, id := range showJust {
					if id == t.Span.ID.Trace {
						return true
					}
				}
				return false
			}
			return true
		},
	})
}

func (a *App) serveAggregate(w http.ResponseWriter, r *http.Request) error {
	// By default we display all traces.
	traces, err := a.Queryer.Traces()
	if err != nil {
		return err
	}

	q := r.URL.Query()

	// If they specified a comma-separated list of specific trace IDs that they
	// are interested in, then we only show those.
	selection := q.Get("selection")
	if len(selection) > 0 {
		var selected []*appdash.Trace
		for _, idStr := range strings.Split(selection, ",") {
			id, err := appdash.ParseID(idStr)
			if err != nil {
				return err
			}
			for _, t := range traces {
				if t.Span.ID.Trace == id {
					selected = append(selected, t)
				}
			}
		}
		traces = selected
	}

	// Perform the aggregation and render the data.
	aggregated, err := a.aggregate(traces, parseAggMode(q.Get("view-mode")))
	if err != nil {
		return err
	}
	return a.renderTemplate(w, r, "aggregate.html", http.StatusOK, &struct {
		TemplateCommon
		Aggregated []*aggItem
	}{
		Aggregated: aggregated,
	})
}

func (a *App) serveTraceUpload(w http.ResponseWriter, r *http.Request) error {
	// Read the uploaded JSON trace data.
	defer r.Body.Close()
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	// Unmarshal the trace.
	var traces []*appdash.Trace
	err = json.Unmarshal(data, &traces)
	if err != nil {
		return err
	}

	// Collect the unmarshaled traces, ignoring any previously existing ones (i.e.
	// ones that would collide / be merged together).
	for _, trace := range traces {
		_, err = a.Store.Trace(trace.Span.ID.Trace)
		if err != appdash.ErrTraceNotFound {
			// The trace collides with an existing trace, ignore it.
			continue
		}

		// Collect the trace (store it for later viewing).
		if err = collectTrace(a.Store, trace); err != nil {
			return err
		}
	}
	return nil
}

package traceapp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// dashboardRow represents a single row in the dashboard. It is encoded to JSON.
type dashboardRow struct {
	Name                      string
	Average, Min, Max, StdDev time.Duration
	Timespans                 int
	URL                       string
}

// serverDashboard serves the dashboard page.
func (a *App) serveDashboard(w http.ResponseWriter, r *http.Request) error {
	if a.Aggregator == nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Dashboard is disabled.")
		return nil
	}

	uData, err := a.Router.URLTo(DashboardDataRoute)
	if err != nil {
		return err
	}

	return a.renderTemplate(w, r, "dashboard.html", http.StatusOK, &struct {
		TemplateCommon
		DataURL       string
		HaveDashboard bool
	}{
		DataURL:       uData.String(),
		HaveDashboard: a.Aggregator != nil,
	})
}

// serveDashboardData serves the JSON data requested by the dashboards table.
func (a *App) serveDashboardData(w http.ResponseWriter, r *http.Request) error {
	if a.Aggregator == nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Dashboard is disabled.")
		return nil
	}

	// Parse the query for the start & end timeline durations.
	var (
		query      = r.URL.Query()
		start, end time.Duration
	)
	if s := query.Get("start"); len(s) > 0 {
		v, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return err
		}
		start = time.Duration(v) * time.Hour
		start -= 72 * time.Hour
	}
	if s := query.Get("end"); len(s) > 0 {
		v, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return err
		}
		end = time.Duration(v) * time.Hour
		end -= 72 * time.Hour
	}

	results, err := a.Aggregator.Aggregate(start, end)
	if err != nil {
		return err
	}

	// Grab the URL to the traces page.
	tracesURL, err := a.Router.URLTo(TracesRoute)
	if err != nil {
		return err
	}

	rows := make([]*dashboardRow, len(results))
	for i, r := range results {
		var stringIDs []string
		for _, slowest := range r.Slowest {
			stringIDs = append(stringIDs, slowest.String())
		}
		tracesURL.RawQuery = "show=" + strings.Join(stringIDs, ",")

		rows[i] = &dashboardRow{
			Name:      r.RootSpanName,
			Average:   r.Average / time.Millisecond,
			Min:       r.Min / time.Millisecond,
			Max:       r.Max / time.Millisecond,
			StdDev:    r.StdDev / time.Millisecond,
			Timespans: int(r.Samples),
			URL:       tracesURL.String(),
		}
	}

	// Encode to JSON.
	j, err := json.Marshal(rows)
	if err != nil {
		return err
	}

	// Write out.
	_, err = io.Copy(w, bytes.NewReader(j))
	return err
}

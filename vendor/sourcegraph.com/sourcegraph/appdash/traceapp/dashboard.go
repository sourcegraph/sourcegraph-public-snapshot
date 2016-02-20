package traceapp

import (
	"bytes"
	"encoding/json"
	"io"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cznic/mathutil"
	"sourcegraph.com/sourcegraph/appdash"
)

// dashboardRow represents a single row in the dashboard. It is encoded to JSON.
type dashboardRow struct {
	Name                      string
	Average, Min, Max, StdDev time.Duration
	Timespans                 int
	URL                       string
}

// newDashboardRow returns a new dashboard row with it's items calculated from
// the given aggregation event and timespan events (the returned row represents
// the whole aggregation event).
//
// The returned row does not have the URL field set.
func newDashboardRow(a appdash.AggregateEvent, timespans []appdash.TimespanEvent) dashboardRow {
	row := dashboardRow{
		Name:      a.Name,
		Timespans: len(timespans),
	}

	// Calculate sum and mean (row.Average), while determining min/max.
	sum := big.NewInt(0)
	for _, ts := range timespans {
		d := ts.End().Sub(ts.Start())
		sum.Add(sum, big.NewInt(int64(d)))
		if row.Min == 0 || d < row.Min {
			row.Min = d
		}
		if row.Max == 0 || d > row.Max {
			row.Max = d
		}
	}
	avg := big.NewInt(0).Div(sum, big.NewInt(int64(len(timespans))))
	row.Average = time.Duration(avg.Int64())

	// Calculate std. deviation.
	sqDiffSum := big.NewInt(0)
	for _, ts := range timespans {
		d := ts.End().Sub(ts.Start())
		diff := big.NewInt(0).Sub(big.NewInt(int64(d)), avg)
		sqDiffSum.Add(sqDiffSum, diff.Mul(diff, diff))
	}
	stdDev := big.NewInt(0).Div(sqDiffSum, big.NewInt(int64(len(timespans))))
	stdDev = mathutil.SqrtBig(stdDev)
	row.StdDev = time.Duration(stdDev.Int64())

	// TODO(slimsag): if we can make the table display the data as formatted by
	// Go (row.Average.String), we'll get much nicer display. But it means we'll
	// need to perform custom sorting on the table (it will think "9ms" > "1m",
	// for example).

	// Divide into milliseconds.
	row.Average = row.Average / time.Millisecond
	row.Min = row.Min / time.Millisecond
	row.Max = row.Max / time.Millisecond
	row.StdDev = row.StdDev / time.Millisecond
	return row
}

// aggTimeFilter removes timespans and slowest-trace IDs from the given
// aggregate event if they were not defined inside the given start and end time.
func aggTimeFilter(a appdash.AggregateEvent, timespans []appdash.TimespanEvent, start, end time.Time) (appdash.AggregateEvent, []appdash.TimespanEvent, bool) {
	cpy := a
	cpy.Slowest = nil
	var cpyTimes []appdash.TimespanEvent
	for n, ts := range timespans {
		if ts.Start().UnixNano() < start.UnixNano() || ts.End().UnixNano() > end.UnixNano() {
			// It started before or after the time period we want.
			continue
		}
		cpyTimes = append(cpyTimes, ts)
		if n < len(a.Slowest) {
			cpy.Slowest = append(cpy.Slowest, a.Slowest[n])
		}
	}
	return cpy, cpyTimes, len(cpyTimes) > 0
}

// serverDashboard serves the dashboard page.
func (a *App) serveDashboard(w http.ResponseWriter, r *http.Request) error {
	uData, err := a.Router.URLTo(DashboardDataRoute)
	if err != nil {
		return err
	}

	return a.renderTemplate(w, r, "dashboard.html", http.StatusOK, &struct {
		TemplateCommon
		DataURL string
	}{
		DataURL: uData.String(),
	})
}

// serveDashboardData serves the JSON data requested by the dashboards table.
func (a *App) serveDashboardData(w http.ResponseWriter, r *http.Request) error {
	traces, err := a.Queryer.Traces()
	if err != nil {
		return err
	}

	// Parse the query for the start & end timeline durations.
	var (
		query      = r.URL.Query()
		start, end time.Time
	)
	basis := time.Now().Add(-72 * time.Hour)
	if s := query.Get("start"); len(s) > 0 {
		v, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return err
		}
		// e.g. if (v)start==0, it'll be -72hrs ago
		start = basis.Add(time.Duration(v) * time.Hour)
	}
	if s := query.Get("end"); len(s) > 0 {
		v, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return err
		}
		// .eg. if (v)end==72, it'll be time.Now()
		end = basis.Add(time.Duration(v) * time.Hour)
	}

	// Grab the URL to the traces page.
	tracesURL, err := a.Router.URLTo(TracesRoute)
	if err != nil {
		return err
	}

	// Important: If it is a nil slice it will be encoded to JSON as null, and the
	// bootstrap-table library will not update the table with "no entries".
	rows := make([]dashboardRow, 0)

	// Produce the rows of data.
	for _, trace := range traces {
		// Grab the aggregation event from the trace, if any.
		agg, timespans, err := trace.Aggregated()
		if err != nil {
			return err
		}
		if agg == nil {
			continue // No aggregation event.
		}

		// Filter the event by our timeline.
		a, timespans, any := aggTimeFilter(*agg, timespans, start, end)
		if !any {
			continue
		}

		// Create a list of slowest trace IDs (but as strings), then produce a
		// URL which will query for it.
		var stringIDs []string
		for _, slowest := range a.Slowest {
			stringIDs = append(stringIDs, slowest.String())
		}
		tracesURL.RawQuery = "show=" + strings.Join(stringIDs, ",")

		// Create the row of data.
		row := newDashboardRow(a, timespans)
		row.URL = tracesURL.String()
		rows = append(rows, row)
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

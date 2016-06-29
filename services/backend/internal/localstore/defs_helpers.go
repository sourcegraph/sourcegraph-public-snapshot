package localstore

import (
	"time"

	"gopkg.in/inconshreveable/log15.v2"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/repotrackutil"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	defsUpdateDuration *prometheus.SummaryVec
	defsSearchDuration *prometheus.SummaryVec
)

func init() {
	defsUpdateDuration = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: "src",
		Subsystem: "defs",
		Name:      "update_duration_seconds",
		Help:      "Duration for updating a def",
		MaxAge:    time.Hour,
	}, []string{"table", "repo", "part"})
	prometheus.MustRegister(defsUpdateDuration)

	defsSearchDuration = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: "src",
		Subsystem: "defs",
		Name:      "search_duration_seconds",
		Help:      "Duration for def search",
		MaxAge:    time.Hour,
	}, []string{"table", "repo", "part"})
	prometheus.MustRegister(defsSearchDuration)
}

type observer struct {
	table       string
	trackedRepo string
	summaryVec  *prometheus.SummaryVec
}

func newDefsUpdateObserver(table, repo string) *observer {
	return &observer{
		table:       table,
		trackedRepo: repotrackutil.GetTrackedRepo(repo),
		summaryVec:  defsUpdateDuration,
	}
}

func newDefsSearchObserver(table, repo string) *observer {
	return &observer{
		table:       table,
		trackedRepo: repotrackutil.GetTrackedRepo(repo),
		summaryVec:  defsSearchDuration,
	}
}

func (d *observer) start(part string) func() {
	start := time.Now()
	var observed bool
	return func() {
		if observed {
			log15.Error("Called observe more than once", "table", d.table, "part", part)
			return
		}
		observed = true
		d.summaryVec.WithLabelValues(d.table, d.trackedRepo, part).Observe(time.Since(start).Seconds())
	}
}

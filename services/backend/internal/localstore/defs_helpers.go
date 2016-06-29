package localstore

import (
	"time"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/repotrackutil"

	"github.com/prometheus/client_golang/prometheus"
)

var defsUpdateDuration = prometheus.NewSummaryVec(prometheus.SummaryOpts{
	Namespace: "src",
	Subsystem: "defs",
	Name:      "update_duration_seconds",
	Help:      "Duration for updating a def",
	MaxAge:    time.Hour,
}, []string{"table", "repo", "part"})

func init() {
	prometheus.MustRegister(defsUpdateDuration)
}

type defsUpdateObserver struct {
	table       string
	trackedRepo string
	starts      map[string]time.Time
}

func newDefsUpdateObserver(table, repo string) *defsUpdateObserver {
	return &defsUpdateObserver{
		table:       table,
		trackedRepo: repotrackutil.GetTrackedRepo(repo),
		starts:      make(map[string]time.Time),
	}
}

func (d *defsUpdateObserver) start(part string) {
	d.starts[part] = time.Now()
}

func (d *defsUpdateObserver) end(part string) {
	since := time.Since(d.starts[part])
	defsUpdateDuration.WithLabelValues(d.table, d.trackedRepo, part).Observe(since.Seconds())
}

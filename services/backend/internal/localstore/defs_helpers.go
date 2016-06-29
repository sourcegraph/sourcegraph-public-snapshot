package localstore

import (
	"sync"
	"time"

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
	mu          sync.Mutex
	table       string
	trackedRepo string
	starts      map[string]time.Time
	summaryVec  *prometheus.SummaryVec
}

func newDefsUpdateObserver(table, repo string) *observer {
	return &observer{
		table:       table,
		trackedRepo: repotrackutil.GetTrackedRepo(repo),
		starts:      make(map[string]time.Time),
		summaryVec:  defsUpdateDuration,
	}
}

func newDefsSearchObserver(table, repo string) *observer {
	return &observer{
		table:       table,
		trackedRepo: repotrackutil.GetTrackedRepo(repo),
		starts:      make(map[string]time.Time),
		summaryVec:  defsSearchDuration,
	}
}

func (d *observer) start(part string) {
	d.mu.Lock()
	d.starts[part] = time.Now()
	d.mu.Unlock()
}

func (d *observer) end(part string) {
	d.mu.Lock()
	since := time.Since(d.starts[part])
	d.summaryVec.WithLabelValues(d.table, d.trackedRepo, part).Observe(since.Seconds())
	d.mu.Unlock()
}

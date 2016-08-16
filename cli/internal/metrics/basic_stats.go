package metrics

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"context"

	"gopkg.in/inconshreveable/log15.v2"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
)

var buildLabels = []string{"build_type"}
var numBuildsGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Namespace: "src",
	Subsystem: "usage_stats",
	Name:      "builds_total",
	Help:      "Total builds on the local Sourcegraph instance.",
}, buildLabels)

var oldestEnqueuedBuildGauge = prometheus.NewGauge(prometheus.GaugeOpts{
	Namespace: "src",
	Subsystem: "builds",
	Name:      "oldest_enqueued_build_seconds",
	Help:      "Age of oldest build in the queue",
})

func init() {
	prometheus.MustRegister(numBuildsGauge)
	prometheus.MustRegister(oldestEnqueuedBuildGauge)
}

// ComputeUsageStats takes a daily snapshot of the basic statistics of all
// local repos.
func ComputeUsageStats(ctx context.Context, interval time.Duration) {
	cl, err := sourcegraph.NewClientFromContext(ctx)
	if err != nil {
		log15.Warn(fmt.Sprintf("ComputeUsageStats: could not construct client, usage stats will not be computed: %s", err))
		return
	}
	for {
		updateNumBuilds(cl, ctx)
		updateOldestEnqueuedRepo(cl, ctx)

		time.Sleep(interval)
	}
}

func updateNumBuilds(cl *sourcegraph.Client, ctx context.Context) {
	numBuilds, err := ComputeBuildStats(cl, ctx)
	if err != nil {
		log15.Warn("ComputeUsageStats: could not compute number of builds", "error", err)
		return
	}
	for buildType, buildCount := range numBuilds {
		numBuildsGauge.WithLabelValues(buildType).Set(float64(buildCount))
	}
}

func updateOldestEnqueuedRepo(cl *sourcegraph.Client, ctx context.Context) {
	buildQueue, err := cl.Builds.List(ctx, &sourcegraph.BuildListOptions{
		Queued:    true,
		Sort:      "created_at",
		Direction: "asc",
		ListOptions: sourcegraph.ListOptions{
			PerPage: 1,
		},
	})
	if err != nil {
		log15.Warn("ComputeUsageStats: could not compute the oldest build in the queue")
		return
	}

	if len(buildQueue.Builds) == 0 {
		// we need to set the age of the oldest build to 0 since
		// it would appear that there was still something on the queue
		oldestEnqueuedBuildGauge.Set(float64(0))
		return
	}

	build := buildQueue.Builds[0]

	age := time.Now().Sub(build.CreatedAt.Time())

	oldestEnqueuedBuildGauge.Set(age.Seconds())
}

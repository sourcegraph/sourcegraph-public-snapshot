package definitions

import (
	"time"

	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

func GitHub() *monitoring.Dashboard {
	return &monitoring.Dashboard{
		Name:        "github",
		Title:       "GitHub",
		Description: "Dashboard to track requests and global concurrency locks for talking to github.com.",
		Groups: []monitoring.Group{
			{
				Title: "GitHub API monitoring",
				Rows: []monitoring.Row{
					{
						{
							Name:        "src_githubcom_concurrency_lock_waiting_requests",
							Description: "number of requests waiting on the global mutex",
							Query:       `max(src_githubcom_concurrency_lock_waiting_requests)`,
							Warning:     monitoring.Alert().GreaterOrEqual(100).For(5 * time.Minute),
							Panel:       monitoring.Panel().LegendFormat("requests waiting"),
							Owner:       monitoring.ObservableOwnerSource,
							NextSteps: `
								- **Check container logs for network connection issues and log entries from the githubcom-concurrency-limiter logger.
								- **Check redis-store health.
								- **Check GitHub status.`,
						},
					},
					{
						{
							Name:        "src_githubcom_concurrency_lock_failed_lock_requests",
							Description: "number of lock failures",
							Query:       `sum(rate(src_githubcom_concurrency_lock_failed_lock_requests[5m]))`,
							Warning:     monitoring.Alert().GreaterOrEqual(100).For(5 * time.Minute),
							Panel:       monitoring.Panel().LegendFormat("failed lock requests"),
							Owner:       monitoring.ObservableOwnerSource,
							NextSteps: `
							- **Check container logs for network connection issues and log entries from the githubcom-concurrency-limiter logger.
							- **Check redis-store health.`,
						},
						{
							Name:        "src_githubcom_concurrency_lock_failed_unlock_requests",
							Description: "number of unlock failures",
							Query:       `sum(rate(src_githubcom_concurrency_lock_failed_unlock_requests[5m]))`,
							Warning:     monitoring.Alert().GreaterOrEqual(100).For(5 * time.Minute),
							Panel:       monitoring.Panel().LegendFormat("failed unlock requests"),
							Owner:       monitoring.ObservableOwnerSource,
							NextSteps: `
							- **Check container logs for network connection issues and log entries from the githubcom-concurrency-limiter logger.
							- **Check redis-store health.`,
						},
					},
					{
						{
							Name:           "src_githubcom_concurrency_lock_requests",
							Description:    "number of locks taken global mutex",
							Query:          `sum(rate(src_githubcom_concurrency_lock_requests[5m]))`,
							NoAlert:        true,
							Panel:          monitoring.Panel().LegendFormat("number of requests"),
							Owner:          monitoring.ObservableOwnerSource,
							Interpretation: "A high number of locks indicates heavy usage of the GitHub API. This might not be a problem, but you should check if request counts are expected.",
						},
						{
							Name:           "src_githubcom_concurrency_lock_acquire_duration_seconds_latency_p75",
							Description:    "75 percentile latency of src_githubcom_concurrency_lock_acquire_duration_seconds",
							Query:          `histogram_quantile(0.75, sum(rate(src_githubcom_concurrency_lock_acquire_duration_seconds_bucket[5m])) by (le))`,
							NoAlert:        true,
							Panel:          monitoring.Panel().LegendFormat("lock acquire latency").Unit(monitoring.Milliseconds),
							Owner:          monitoring.ObservableOwnerSource,
							Interpretation: `99 percentile latency of acquiring the global GitHub concurrency lock.`,
						},
					},
				},
			},
		},
	}
}

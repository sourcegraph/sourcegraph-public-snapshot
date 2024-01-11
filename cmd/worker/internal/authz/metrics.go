package authz

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// The metrics that are exposed to Prometheus.
var (
	metricsSyncDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "src_repo_perms_syncer_sync_duration_seconds",
		Help:    "Time spent on syncing permissions",
		Buckets: []float64{1, 2, 5, 10, 30, 60, 120},
	}, []string{"type", "success"})
	metricsSyncErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "src_repo_perms_syncer_sync_errors_total",
		Help: "Total number of permissions sync errors",
	}, []string{"type"})
	metricsSuccessPermsSyncs = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "src_repo_perms_syncer_success_syncs",
		Help: "Total number of successful permissions syncs",
	}, []string{"type"})
	metricsFirstPermsSyncs = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "src_repo_perms_syncer_initial_syncs",
		Help: "Total number of new user/repo permissions syncs",
	}, []string{"type"})
	metricsPermsConsecutiveSyncDelay = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "src_repo_perms_syncer_perms_consecutive_sync_delay",
		Help: "The duration in seconds between last and current complete premissions sync.",
	}, []string{"type"})
	metricsPermsFirstSyncDelay = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "src_repo_perms_syncer_perms_first_sync_delay",
		Help: "The duration in seconds it took for first user/repo complete perms sync after creation",
	}, []string{"type"})
)

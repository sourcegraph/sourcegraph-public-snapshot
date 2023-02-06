package authz

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// The metrics that are exposed to Prometheus.
var (
	metricsNoPerms = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "src_repoupdater_perms_syncer_no_perms",
		Help: "The number of records that do not have any permissions",
	}, []string{"type"})
	metricsStalePerms = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "src_repoupdater_perms_syncer_stale_perms",
		Help: "The number of records that have stale permissions",
	}, []string{"type"})
	metricsStrictStalePerms = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "src_repoupdater_perms_syncer_strict_stale_perms",
		Help: "The number of records that have permissions older than 1h",
	}, []string{"type"})
	metricsPermsGap = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "src_repoupdater_perms_syncer_perms_gap_seconds",
		Help: "The time gap between least and most up to date permissions",
	}, []string{"type"})
	metricsSyncDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "src_repoupdater_perms_syncer_sync_duration_seconds",
		Help:    "Time spent on syncing permissions",
		Buckets: []float64{1, 2, 5, 10, 30, 60, 120},
	}, []string{"type", "success"})
	metricsSyncErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "src_repoupdater_perms_syncer_sync_errors_total",
		Help: "Total number of permissions sync errors",
	}, []string{"type"})
	metricsQueueSize = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "src_repoupdater_perms_syncer_queue_size",
		Help: "The size of the sync request queue",
	})
	metricsConcurrentSyncs = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "src_repoupdater_perms_syncer_concurrent_syncs",
		Help: "The number of concurrent permissions syncs",
	}, []string{"type"})
	metricsSuccessPermsSyncs = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "src_repoupdater_perms_syncer_success_syncs",
		Help: "Total number of successful permissions syncs",
	}, []string{"type"})
	metricsFailedPermsSyncs = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "src_repoupdater_perms_syncer_failed_syncs",
		Help: "Total number of failed permissions syncs",
	}, []string{"type"})
	metricsFirstPermsSyncs = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "src_repoupdater_perms_syncer_initial_syncs",
		Help: "Total number of new user/repo permissions syncs",
	}, []string{"type"})
	metricsPermsConsecutiveSyncDelay = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "src_repoupdater_perms_syncer_perms_consecutive_sync_delay",
		Help: "The duration in seconds between last and current complete premissions sync.",
	}, []string{"type"})
	metricsPermsFirstSyncDelay = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "src_repoupdater_perms_syncer_perms_first_sync_delay",
		Help: "The duration in seconds it took for first user/repo complete perms sync after creation",
	}, []string{"type"})
	metricsItemsSyncScheduled = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "src_repoupdater_perms_syncer_items_sync_scheduled",
		Help: "The number of users/repos scheduled for sync",
	}, []string{"type", "priority"})
)

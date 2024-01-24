package repos

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	tagFamily  = "family"
	tagOwner   = "owner"
	tagSuccess = "success"
	tagState   = "state"
	tagReason  = "reason"
)

var (
	lastSync = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "src_repoupdater_syncer_sync_last_time",
		Help: "The last time a sync finished",
	}, []string{tagFamily})

	syncStarted = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "src_repoupdater_syncer_start_sync",
		Help: "A sync was started",
	}, []string{tagFamily, tagOwner})

	syncErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "src_repoupdater_syncer_sync_errors_total",
		Help: "Total number of sync errors",
	}, []string{tagFamily, tagOwner, tagReason})

	syncDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "src_repoupdater_syncer_sync_duration_seconds",
		Help: "Time spent syncing",
	}, []string{tagSuccess, tagFamily})

	syncedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "src_repoupdater_syncer_synced_repos_total",
		Help: "Total number of synced repositories",
	}, []string{tagState})
)

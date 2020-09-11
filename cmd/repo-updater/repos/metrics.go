package repos

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	phabricatorUpdateTime = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "src_repoupdater_time_last_phabricator_sync",
		Help: "The last time a comprehensive Phabricator sync finished",
	}, []string{"id"})

	lastSync = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "src_repoupdater_syncer_sync_last_time",
		Help: "The last time a sync finished",
	}, []string{})

	syncedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "src_repoupdater_syncer_synced_repos_total",
		Help: "Total number of synced repositories",
	}, []string{"state"})

	syncErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "src_repoupdater_syncer_sync_errors_total",
		Help: "Total number of sync errors",
	}, []string{})

	syncDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "src_repoupdater_syncer_sync_duration_seconds",
		Help: "Time spent syncing",
	}, []string{"success"})

	purgeSuccess = promauto.NewCounter(prometheus.CounterOpts{
		Name: "src_repoupdater_purge_success",
		Help: "Incremented each time we remove a repository clone.",
	})
	purgeFailed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "src_repoupdater_purge_failed",
		Help: "Incremented each time we try and fail to remove a repository clone.",
	})

	schedError = promauto.NewCounter(prometheus.CounterOpts{
		Name: "src_repoupdater_sched_error",
		Help: "Incremented each time we encounter an error updating a repository.",
	})
	schedLoops = promauto.NewCounter(prometheus.CounterOpts{
		Name: "src_repoupdater_sched_loops",
		Help: "Incremented each time the scheduler loops.",
	})
	schedAutoFetch = promauto.NewCounter(prometheus.CounterOpts{
		Name: "src_repoupdater_sched_auto_fetch",
		Help: "Incremented each time the scheduler updates a managed repository due to hitting a deadline.",
	})
	schedManualFetch = promauto.NewCounter(prometheus.CounterOpts{
		Name: "src_repoupdater_sched_manual_fetch",
		Help: "Incremented each time the scheduler updates a repository due to user traffic.",
	})
	schedKnownRepos = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "src_repoupdater_sched_known_repos",
		Help: "The number of repositories that are managed by the scheduler.",
	})
)

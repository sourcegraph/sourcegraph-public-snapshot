package repos

import (
	"context"

	"github.com/inconshreveable/log15"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/sourcegraph/sourcegraph/internal/db/dbutil"
)

const (
	tagExternalServiceID = "external_service_id"
	tagFamily            = "family"
	tagID                = "id"
	tagState             = "state"
	tagSuccess           = "success"
)

var (
	phabricatorUpdateTime = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "src_repoupdater_time_last_phabricator_sync",
		Help: "The last time a comprehensive Phabricator sync finished",
	}, []string{tagID})

	lastSync = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "src_repoupdater_syncer_sync_last_time",
		Help: "The last time a sync finished",
	}, []string{tagExternalServiceID, tagFamily})

	syncedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "src_repoupdater_syncer_synced_repos_total",
		Help: "Total number of synced repositories",
	}, []string{tagState, tagExternalServiceID, tagFamily})

	syncErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "src_repoupdater_syncer_sync_errors_total",
		Help: "Total number of sync errors",
	}, []string{tagExternalServiceID, tagFamily})

	syncDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "src_repoupdater_syncer_sync_duration_seconds",
		Help: "Time spent syncing",
	}, []string{tagSuccess, tagExternalServiceID, tagFamily})

	syncBackoffDuration = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "src_repoupdater_syncer_sync_backoff_duration_seconds",
		Help: "Backoff duration after a sync",
	}, []string{tagExternalServiceID})

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

func MustRegisterMetrics(db dbutil.DB) {
	scanCount := func(sql string) (float64, error) {
		row := db.QueryRowContext(context.Background(), sql)
		var count int64
		err := row.Scan(&count)
		if err != nil {
			return 0, err
		}

		return float64(count), nil
	}

	promauto.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "src_repoupdater_user_external_services_total",
		Help: "The total number of external services added by users",
	}, func() float64 {
		count, err := scanCount(`
-- source: cmd/repo-updater/repos/metrics.go:src_repoupdater_user_external_services_total
SELECT COUNT(*) FROM external_services
WHERE namespace_user_id IS NOT NULL
`)
		if err != nil {
			log15.Error("Failed to get total user external services", "err", err)
			return 0
		}
		return count
	})

	promauto.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "src_repoupdater_user_repos_total",
		Help: "The total number of repositories added by users",
	}, func() float64 {
		count, err := scanCount(`
-- source: cmd/repo-updater/repos/metrics.go:src_repoupdater_user_repos_total
SELECT COUNT(*) FROM external_service_repos
WHERE external_service_id IN (
		SELECT DISTINCT(id) FROM external_services
		WHERE namespace_user_id IS NOT NULL
	)
`)
		if err != nil {
			log15.Error("Failed to get total user repositories", "err", err)
			return 0
		}
		return count
	})

	promauto.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "src_repoupdater_user_with_external_services_total",
		Help: "The total number of users who have added external services",
	}, func() float64 {
		count, err := scanCount(`
-- source: cmd/repo-updater/repos/metrics.go:src_repoupdater_user_with_external_services_total
SELECT COUNT(DISTINCT(namespace_user_id)) AS total
FROM external_services
WHERE namespace_user_id IS NOT NULL
`)
		if err != nil {
			log15.Error("Failed to get total users with external services", "err", err)
			return 0
		}
		return count
	})
}

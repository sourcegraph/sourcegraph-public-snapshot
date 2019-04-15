package repos

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	githubUpdateTime = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "src",
		Subsystem: "repoupdater",
		Name:      "time_last_github_sync",
		Help:      "The last time a comprehensive GitHub sync finished",
	}, []string{"id"})
	gitlabUpdateTime = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "src",
		Subsystem: "repoupdater",
		Name:      "time_last_gitlab_sync",
		Help:      "The last time a comprehensive GitLab sync finished",
	}, []string{"id"})
	awsCodeCommitUpdateTime = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "src",
		Subsystem: "repoupdater",
		Name:      "time_last_awscodecommit_sync",
		Help:      "The last time a comprehensive AWS Code Commit sync finished",
	}, []string{"id"})
	phabricatorUpdateTime = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "src",
		Subsystem: "repoupdater",
		Name:      "time_last_phabricator_sync",
		Help:      "The last time a comprehensive Phabricator sync finished",
	}, []string{"id"})
	bitbucketServerUpdateTime = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "src",
		Subsystem: "repoupdater",
		Name:      "time_last_bitbucketserver_sync",
		Help:      "The last time a comprehensive Bitbucket Server sync finished",
	}, []string{"id"})

	// gitoliteUpdateTime is an ugly duckling as repo-updater just
	// hits a HTTP endpoint on the frontend that returns after
	// updating all Gitolite connections. As as result, we cannot
	// apply an "id" label to differentiate update times for different
	// Gitolite connections, so this is a Gauge instead of a GaugeVec.
	gitoliteUpdateTime = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "src",
		Subsystem: "repoupdater",
		Name:      "time_last_gitolite_sync",
		Help:      "The last time a comprehensive Gitolite sync finished",
	})

	lastSync = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "src",
		Subsystem: "repoupdater",
		Name:      "syncer_sync_last_time",
		Help:      "The last time a sync finished",
	}, []string{})

	syncedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "src",
		Subsystem: "repoupdater",
		Name:      "syncer_synced_repos_total",
		Help:      "Total number of synced repositories",
	}, []string{"state"})

	syncErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "src",
		Subsystem: "repoupdater",
		Name:      "syncer_sync_errors_total",
		Help:      "Total number of sync errors",
	}, []string{})

	syncDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "src",
		Subsystem: "repoupdater",
		Name:      "other_external_services_sync_duration",
		Help:      "Time spent syncing",
	}, []string{"success"})

	purgeSuccess = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "src",
		Subsystem: "repoupdater",
		Name:      "purge_success",
		Help:      "Incremented each time we remove a repository clone.",
	})
	purgeFailed = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "src",
		Subsystem: "repoupdater",
		Name:      "purge_failed",
		Help:      "Incremented each time we try and fail to remove a repository clone.",
	})
	purgeSkipped = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "src",
		Subsystem: "repoupdater",
		Name:      "purge_skipped",
		Help:      "Incremented each time we skip a repository clone to remove.",
	})

	schedError = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "src",
		Subsystem: "repoupdater",
		Name:      "sched_error",
		Help:      "Incremented each time we encounter an error updating a repository.",
	})
	schedLoops = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "src",
		Subsystem: "repoupdater",
		Name:      "sched_loops",
		Help:      "Incremented each time the scheduler loops.",
	})
	schedAutoFetch = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "src",
		Subsystem: "repoupdater",
		Name:      "sched_auto_fetch",
		Help:      "Incremented each time the scheduler updates a managed repository due to hitting a deadline.",
	})
	schedManualFetch = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "src",
		Subsystem: "repoupdater",
		Name:      "sched_manual_fetch",
		Help:      "Incremented each time the scheduler updates a repository due to user traffic.",
	})
	schedKnownRepos = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "src",
		Subsystem: "repoupdater",
		Name:      "sched_known_repos",
		Help:      "The number of repositories that are managed by the scheduler.",
	})
)

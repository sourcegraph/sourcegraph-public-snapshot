package repos

import "github.com/prometheus/client_golang/prometheus"

var (
	githubUpdateTime = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "src",
		Subsystem: "repoupdater",
		Name:      "time_last_github_sync",
		Help:      "The last time a comprehensive GitHub sync finished",
	}, []string{"id"})
	gitlabUpdateTime = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "src",
		Subsystem: "repoupdater",
		Name:      "time_last_gitlab_sync",
		Help:      "The last time a comprehensive GitLab sync finished",
	}, []string{"id"})
	awsCodeCommitUpdateTime = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "src",
		Subsystem: "repoupdater",
		Name:      "time_last_awscodecommit_sync",
		Help:      "The last time a comprehensive AWS Code Commit sync finished",
	}, []string{"id"})
	phabricatorUpdateTime = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "src",
		Subsystem: "repoupdater",
		Name:      "time_last_phabricator_sync",
		Help:      "The last time a comprehensive Phabricator sync finished",
	}, []string{"id"})
	bitbucketServerUpdateTime = prometheus.NewGaugeVec(prometheus.GaugeOpts{
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
	gitoliteUpdateTime = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "src",
		Subsystem: "repoupdater",
		Name:      "time_last_gitolite_sync",
		Help:      "The last time a comprehensive Gitolite sync finished",
	})

	repoListUpdateTime = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "src",
		Subsystem: "repoupdater",
		Name:      "time_last_repolist_sync",
		Help:      "The time the last repository sync loop completed",
	})

	purgeSuccess = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "src",
		Subsystem: "repoupdater",
		Name:      "purge_success",
		Help:      "Incremented each time we remove a repository clone.",
	})
	purgeFailed = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "src",
		Subsystem: "repoupdater",
		Name:      "purge_failed",
		Help:      "Incremented each time we try and fail to remove a repository clone.",
	})
	purgeSkipped = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "src",
		Subsystem: "repoupdater",
		Name:      "purge_skipped",
		Help:      "Incremented each time we skip a repository clone to remove.",
	})

	schedError = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "src",
		Subsystem: "repoupdater",
		Name:      "sched_error",
		Help:      "Incremented each time we encounter an error updating a repository.",
	})
	schedLoops = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "src",
		Subsystem: "repoupdater",
		Name:      "sched_loops",
		Help:      "Incremented each time the scheduler loops.",
	})
	schedAutoFetch = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "src",
		Subsystem: "repoupdater",
		Name:      "sched_auto_fetch",
		Help:      "Incremented each time the scheduler updates a managed repository due to hitting a deadline.",
	})
	schedManualFetch = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "src",
		Subsystem: "repoupdater",
		Name:      "sched_manual_fetch",
		Help:      "Incremented each time the scheduler updates a repository due to user traffic.",
	})
	schedKnownRepos = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "src",
		Subsystem: "repoupdater",
		Name:      "sched_known_repos",
		Help:      "The number of unique repositories that have been managed by the scheduler.",
	})
	schedScale = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "src",
		Subsystem: "repoupdater",
		Name:      "sched_scale",
		Help:      "The scheduler interval scale.",
	})
)

func init() {
	prometheus.MustRegister(githubUpdateTime)
	prometheus.MustRegister(gitlabUpdateTime)
	prometheus.MustRegister(awsCodeCommitUpdateTime)
	prometheus.MustRegister(phabricatorUpdateTime)
	prometheus.MustRegister(bitbucketServerUpdateTime)
	prometheus.MustRegister(gitoliteUpdateTime)
	prometheus.MustRegister(repoListUpdateTime)
	prometheus.MustRegister(purgeSuccess)
	prometheus.MustRegister(purgeFailed)
	prometheus.MustRegister(purgeSkipped)
	prometheus.MustRegister(schedError)
	prometheus.MustRegister(schedLoops)
	prometheus.MustRegister(schedAutoFetch)
	prometheus.MustRegister(schedManualFetch)
	prometheus.MustRegister(schedKnownRepos)
	prometheus.MustRegister(schedScale)
}

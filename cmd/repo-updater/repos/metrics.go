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
)

func init() {
	prometheus.MustRegister(githubUpdateTime)
	prometheus.MustRegister(gitlabUpdateTime)
	prometheus.MustRegister(awsCodeCommitUpdateTime)
	prometheus.MustRegister(phabricatorUpdateTime)
	prometheus.MustRegister(bitbucketServerUpdateTime)
	prometheus.MustRegister(gitoliteUpdateTime)
	prometheus.MustRegister(repoListUpdateTime)
}

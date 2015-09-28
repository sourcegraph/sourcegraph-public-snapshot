package statsutil

import (
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/prometheus"
	stdprometheus "github.com/prometheus/client_golang/prometheus"

	"golang.org/x/net/context"
	"gopkg.in/inconshreveable/log15.v2"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/fed"
)

var numReposGauge = prometheus.NewGauge(stdprometheus.GaugeOpts{
	Namespace: "src",
	Subsystem: "usage_stats",
	Name:      "repos_total",
	Help:      "Total repos on the local Sourcegraph instance.",
}, nil)

var committerLabels = []string{"domain"}
var numCommittersGauge = prometheus.NewGauge(stdprometheus.GaugeOpts{
	Namespace: "src",
	Subsystem: "usage_stats",
	Name:      "committers_total",
	Help:      "Total committers on the local Sourcegraph instance.",
}, committerLabels)

var buildLabels = []string{"build_type"}
var numBuildsGauge = prometheus.NewGauge(stdprometheus.GaugeOpts{
	Namespace: "src",
	Subsystem: "usage_stats",
	Name:      "builds_total",
	Help:      "Total builds on the local Sourcegraph instance.",
}, buildLabels)

var numUsersGauge = prometheus.NewGauge(stdprometheus.GaugeOpts{
	Namespace: "src",
	Subsystem: "usage_stats",
	Name:      "users_total",
	Help:      "Total users on the local Sourcegraph instance.",
}, nil)

// ComputeUsageStats takes a daily snapshot of the basic statistics of all
// local repos.
func ComputeUsageStats(ctx context.Context, interval time.Duration) {
	cl := sourcegraph.NewClientFromContext(ctx)
	if cl == nil {
		log15.Warn("ComputeUsageStats: could not construct client, usage stats will not be computed")
		return
	}
	for {
		updateNumReposAndCommitters(cl, ctx)
		updateNumBuilds(cl, ctx)
		updateNumUsers(cl, ctx)

		time.Sleep(interval)
	}
}

func updateNumReposAndCommitters(cl *sourcegraph.Client, ctx context.Context) {
	reposList, err := cl.Repos.List(ctx, &sourcegraph.RepoListOptions{})
	if err != nil {
		log15.Warn("ComputeUsageStats: could not compute number of repos", "error", err)
		return
	}
	numReposGauge.Set(float64(len(reposList.Repos)))

	if fed.Config.IsRoot {
		// don't compute committer stats on the mothership.
		return
	}
	numCommitters, err := CountCommittersPerDomain(cl, ctx, reposList)
	if err != nil {
		log15.Warn("ComputeUsageStats: could not compute number of committers", "error", err)
		return
	}
	for domain, count := range numCommitters {
		domainLabel := metrics.Field{Key: "domain", Value: domain}
		numCommittersGauge.With(domainLabel).Set(float64(count))
	}
}

func updateNumBuilds(cl *sourcegraph.Client, ctx context.Context) {
	if fed.Config.IsRoot {
		// don't compute build stats on the mothership.
		// TODO(pararth): measure ComputeBuildStats performance on
		// sourcegraph.com and decide whether to turn this on for prod.
		return
	}
	numBuilds, err := ComputeBuildStats(cl, ctx)
	if err != nil {
		log15.Warn("ComputeUsageStats: could not compute number of builds", "error", err)
		return
	}
	for buildType, buildCount := range numBuilds {
		buildTypeLabel := metrics.Field{Key: "build_type", Value: buildType}
		numBuildsGauge.With(buildTypeLabel).Set(float64(buildCount))
	}
}

func updateNumUsers(cl *sourcegraph.Client, ctx context.Context) {
	usersList, err := cl.Users.List(ctx, &sourcegraph.UsersListOptions{})
	if err != nil {
		log15.Warn("ComputeUsageStats: could not compute number of users", "error", err)
		return
	}
	numUsersGauge.Set(float64(len(usersList.Users)))
}

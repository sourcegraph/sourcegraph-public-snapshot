//docker:user sourcegraph

// Command repo-updater periodically updates repositories configured in site configuration and serves repository
// metadata from multiple external code hosts.
package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/opentracing-contrib/go-stdlib/nethttp"
	opentracing "github.com/opentracing/opentracing-go"
	"gopkg.in/inconshreveable/log15.v2"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/repo-updater/repoupdater"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/debugserver"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/tracer"
)

var (
	profBindAddr = env.Get("SRC_PROF_HTTP", "", "net/http/pprof http bind address.")
)

func main() {
	ctx := context.Background()
	env.Lock()
	env.HandleHelpFlag()
	tracer.Init("repo-updater")

	// Filter log output by level.
	lvl, err := log15.LvlFromString(env.LogLevel)
	if err == nil {
		log15.Root().SetHandler(log15.LvlFilterHandler(lvl, log15.StderrHandler))
	}

	if profBindAddr != "" {
		go debugserver.Start(profBindAddr)
		log.Printf("Profiler available on %s/pprof", profBindAddr)
	}

	// Start up handler that frontend relies on
	var repoupdater repoupdater.Server
	handler := nethttp.Middleware(opentracing.GlobalTracer(), repoupdater.Handler())
	log15.Info("repo-updater: listening", "addr", ":3182")
	srv := &http.Server{Addr: ":3182", Handler: handler}
	go func() { log.Fatal(srv.ListenAndServe()) }()

	// Sync relies on access to frontend, so wait until it has started up.
	waitForFrontend(ctx)

	// Repos List syncing thread
	go func() {
		repos.RunRepositorySyncWorker(ctx)
	}()

	// GitHub Repository syncing thread
	go repos.RunGitHubRepositorySyncWorker(ctx)

	// GitLab Repository syncing thread
	go repos.RunGitLabRepositorySyncWorker(ctx)

	// AWS CodeCommit repository syncing thread
	go repos.RunAWSCodeCommitRepositorySyncWorker(ctx)

	// Phabricator Repository syncing thread
	go repos.RunPhabricatorRepositorySyncWorker(ctx)

	// Gitolite syncing thread
	go repos.RunGitoliteRepositorySyncWorker(ctx)

	// Bitbucket Server syncing thread
	go repos.RunBitbucketServerRepositorySyncWorker(ctx)

	select {}
}

func waitForFrontend(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := api.InternalClient.RetryPingUntilAvailable(ctx); err != nil {
		log15.Warn("frontend not available at startup (will periodically try to reconnect)", "err", err)
	}
}

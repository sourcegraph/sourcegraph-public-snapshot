// Command repo-updater periodically updates repositories configured in site configuration and serves repository
// metadata from multiple external code hosts.
package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/opentracing-contrib/go-stdlib/nethttp"
	opentracing "github.com/opentracing/opentracing-go"
	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repoupdater"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/debugserver"
	"github.com/sourcegraph/sourcegraph/pkg/env"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver"
	"github.com/sourcegraph/sourcegraph/pkg/repoupdater/protocol"
	"github.com/sourcegraph/sourcegraph/pkg/tracer"
)

const port = "3182"

func main() {
	ctx := context.Background()
	env.Lock()
	env.HandleHelpFlag()
	tracer.Init()

	go debugserver.Start(debugserver.Endpoint{
		Name: "Repo Updater State",
		Path: "/repo-updater-state",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var data interface{}
			if conf.UpdateScheduler2Enabled() {
				data = repos.Scheduler.DebugDump()
			} else {
				data = repos.QueueSnapshot()
			}

			d, err := json.MarshalIndent(data, "", "  ")
			if err != nil {
				http.Error(w, "failed to marshal snapshot: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write(d)
		}),
	})

	// Synced repos will be sent here.
	synced := make(chan *protocol.RepoInfo)

	// Other external services syncing thread. Repo updates will be sent on the given channel.
	otherSyncer := repos.NewOtherReposSyncer(
		repos.NewInternalAPI(10*time.Second),
		synced,
	)

	// Start up handler that frontend relies on
	repoupdater := repoupdater.Server{OtherReposSyncer: otherSyncer}
	handler := nethttp.Middleware(opentracing.GlobalTracer(), repoupdater.Handler())
	host := ""
	if env.InsecureDev {
		host = "127.0.0.1"
	}
	addr := net.JoinHostPort(host, port)
	log15.Info("repo-updater: listening", "addr", addr)
	srv := &http.Server{Addr: addr, Handler: handler}
	go func() { log.Fatal(srv.ListenAndServe()) }()

	// Sync relies on access to frontend and git-server, so wait until they started up.
	api.WaitForFrontend(ctx)
	gitserver.DefaultClient.WaitForGitServers(ctx)

	// Repos List syncing thread
	go repos.RunRepositorySyncWorker(ctx)

	// Repos purging thread
	go repos.RunRepositoryPurgeWorker(ctx)

	// GitHub connections and repos syncing threads
	go repos.SyncGitHubConnections(ctx)
	go repos.RunGitHubRepositorySyncWorker(ctx)

	// GitLab connections and repos syncing threads
	go repos.SyncGitLabConnections(ctx)
	go repos.RunGitLabRepositorySyncWorker(ctx)

	// AWS CodeCommit connections and repos syncing threads
	go repos.SyncAWSCodeCommitConnections(ctx)
	go repos.RunAWSCodeCommitRepositorySyncWorker(ctx)

	// Phabricator Repository syncing thread
	go repos.RunPhabricatorRepositorySyncWorker(ctx)

	// Gitolite syncing thread
	go repos.RunGitoliteRepositorySyncWorker(ctx)

	// Bitbucket connections and repos syncing threads
	go repos.SyncBitbucketServerConnections(ctx)
	go repos.RunBitbucketServerRepositorySyncWorker(ctx)

	// Start other repos syncer syncing thread
	go func() { log.Fatal(otherSyncer.Run(ctx, repos.GetUpdateInterval())) }()

	// Start other repos updates scheduler relay thread.
	go func() {
		for repo := range synced {
			if conf.Get().DisableAutoGitUpdates {
				continue
			} else if conf.UpdateScheduler2Enabled() {
				repos.Scheduler.UpdateOnce(repo.Name, repo.VCS.URL)
			} else {
				repos.UpdateOnce(ctx, repo.Name, repo.VCS.URL)
			}
		}
	}()

	select {}
}

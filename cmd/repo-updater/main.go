// Command repo-updater periodically updates repositories configured in site configuration and serves repository
// metadata from multiple external code hosts.
package main

import (
	"context"
	"database/sql"
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

	db, err := repos.NewDB(repos.NewDSNFromEnv())
	store, err := repos.NewDBStore(ctx, db, sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		log.Fatalf("failed to initalise db store: %v", err)
	}

	newSyncerEnabled := map[string]bool{}
	for _, kind := range conf.NewRepoSyncerEnabledExternalServices() {
		newSyncerEnabled[kind] = true
	}

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

	var sources []repos.Source
	for _, kind := range []string{
		"AWSCODECOMMIT",
		"BITBUCKETSERVER",
		"GITHUB",
		"GITLAB",
		"GITOLITE",
		"PHABRICATOR",
		"OTHER",
	} {
		if ok := newSyncerEnabled[kind]; ok {
			switch kind {
			case "GITHUB":
				sources = append(sources, repos.NewGithubSource(conf.GitHubConfigs))
			default:
				log.Fatalf("repo.Source not implemented yet for external service of kind %q", kind)
			}
			continue
		}

		switch kind {
		case "AWSCODECOMMIT":
			go repos.SyncAWSCodeCommitConnections(ctx)
			go repos.RunAWSCodeCommitRepositorySyncWorker(ctx)
		case "BITBUCKETSERVER":
			go repos.SyncBitbucketServerConnections(ctx)
			go repos.RunBitbucketServerRepositorySyncWorker(ctx)
		case "GITHUB":
			go repos.SyncGitHubConnections(ctx)
			go repos.RunGitHubRepositorySyncWorker(ctx)
		case "GITLAB":
			go repos.SyncGitLabConnections(ctx)
			go repos.RunGitLabRepositorySyncWorker(ctx)
		case "GITOLITE":
			go repos.RunGitoliteRepositorySyncWorker(ctx)
		case "PHABRICATOR":
			go repos.RunPhabricatorRepositorySyncWorker(ctx)
		case "OTHER":
			go func() { log.Fatal(otherSyncer.Run(ctx, repos.GetUpdateInterval())) }()

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

		default:
			log.Fatalf("unknown external service kind %q", kind)
		}
	}

	syncer := repos.NewSyncer(10*time.Second, store, sources, func() time.Time {
		// XXX(tsenart): It seems like the current db layer in the frontend API
		// doesn't set the timezone to UTC. Figure out how to migrate TZs to UTC
		// and ensure it's the used timezone across the board.
		return time.Now().UTC()
	})

	go func() { log.Fatal(syncer.Run(ctx)) }()

	select {}
}

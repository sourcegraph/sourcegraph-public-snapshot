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

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repoupdater"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbutil"
	"github.com/sourcegraph/sourcegraph/pkg/debugserver"
	"github.com/sourcegraph/sourcegraph/pkg/env"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver"
	"github.com/sourcegraph/sourcegraph/pkg/httpcli"
	"github.com/sourcegraph/sourcegraph/pkg/trace"
	"github.com/sourcegraph/sourcegraph/pkg/tracer"
	"github.com/sourcegraph/sourcegraph/schema"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

const port = "3182"

func main() {
	ctx := context.Background()
	env.Lock()
	env.HandleHelpFlag()
	tracer.Init()

	clock := func() time.Time { return time.Now().UTC() }

	// Syncing relies on access to frontend and git-server, so wait until they started up.
	if err := api.InternalClient.WaitForFrontend(ctx); err != nil {
		log.Fatalf("sourcegraph-frontend not reachable: %v", err)
	}
	log15.Debug("detected frontend ready")

	gitserver.DefaultClient.WaitForGitServers(ctx)
	log15.Debug("detected gitservers ready")

	dsn := conf.Get().ServiceConnections.PostgresDSN
	conf.Watch(func() {
		newDSN := conf.Get().ServiceConnections.PostgresDSN
		if dsn != newDSN {
			// The DSN was changed (e.g. by someone modifying the env vars on
			// the frontend). We need to respect the new DSN. Easiest way to do
			// that is to restart our service (kubernetes/docker/goreman will
			// handle starting us back up).
			log.Fatalf("Detected repository DSN change, restarting to take effect: %q", newDSN)
		}
	})
	db, err := dbutil.NewDB(dsn, "repo-updater")
	if err != nil {
		log.Fatalf("failed to initialize db store: %v", err)
	}

	var store repos.Store
	{
		m := repos.NewStoreMetrics()
		for _, om := range []*repos.OperationMetrics{
			m.Transact,
			m.Done,
			m.ListRepos,
			m.UpsertRepos,
			m.ListExternalServices,
			m.UpsertExternalServices,
			m.ListAllRepoNames,
		} {
			om.MustRegister(prometheus.DefaultRegisterer)
		}

		store = repos.NewObservedStore(
			repos.NewDBStore(db, sql.TxOptions{Isolation: sql.LevelSerializable}),
			log15.Root(),
			m,
			trace.Tracer{Tracer: opentracing.GlobalTracer()},
		)
	}

	cf := repos.NewHTTPClientFactory()
	var src repos.Sourcer
	{
		m := repos.NewSourceMetrics()
		m.ListRepos.MustRegister(prometheus.DefaultRegisterer)

		src = repos.NewSourcer(cf, repos.ObservedSource(log15.Root(), m))
	}

	scheduler := repos.NewUpdateScheduler()
	server := repoupdater.Server{
		Store:           store,
		Scheduler:       scheduler,
		GitserverClient: gitserver.DefaultClient,
	}

	var handler http.Handler
	{
		m := repoupdater.NewHandlerMetrics()
		m.ServeHTTP.MustRegister(prometheus.DefaultRegisterer)
		handler = repoupdater.ObservedHandler(
			log15.Root(),
			m,
			opentracing.GlobalTracer(),
		)(server.Handler())
	}

	if envvar.SourcegraphDotComMode() {
		src, err := makeGitHubDotComSrc(ctx, store, cf)
		if err != nil {
			log.Fatalf("failed to create Github.com source: %v", err)
		}
		server.GithubDotComSource = src
	}

	diffs := make(chan repos.Diff)
	syncer := repos.NewSyncer(store, src, diffs, clock)
	syncer.FailFullSync = envvar.SourcegraphDotComMode()
	server.Syncer = syncer

	if !envvar.SourcegraphDotComMode() {
		go func() { log.Fatal(syncer.Run(ctx, repos.GetUpdateInterval())) }()
	}

	gps := repos.NewGitolitePhabricatorMetadataSyncer(store)

	// Start new repo syncer updates scheduler relay thread.
	go func() {
		for diff := range diffs {
			if len(diff.Added) > 0 {
				log15.Debug("syncer.sync", "diff.added", diff.Added.Names())
			}

			if len(diff.Modified) > 0 {
				log15.Debug("syncer.sync", "diff.modified", diff.Modified.Names())
			}

			if len(diff.Deleted) > 0 {
				log15.Debug("syncer.sync", "diff.deleted", diff.Deleted.Names())
			}

			if !envvar.SourcegraphDotComMode() {
				rs := diff.Repos()
				if !conf.Get().DisableAutoGitUpdates {
					scheduler.Update(rs...)
				}

				go func() {
					if err := gps.Sync(ctx, rs); err != nil {
						log15.Error("GitolitePhabricatorMetadataSyncer", "error", err)
					}
				}()
			}
		}
	}()
	log15.Debug("started new repo syncer updates scheduler relay thread")

	go repos.RunPhabricatorRepositorySyncWorker(ctx, store)

	if !envvar.SourcegraphDotComMode() {
		// git-server repos purging thread
		go repos.RunRepositoryPurgeWorker(ctx)
	}

	// Git fetches scheduler
	go repos.RunScheduler(ctx, scheduler)
	log15.Debug("started scheduler")

	host := ""
	if env.InsecureDev {
		host = "127.0.0.1"
	}

	addr := net.JoinHostPort(host, port)
	log15.Info("server listening", "addr", addr)
	srv := &http.Server{Addr: addr, Handler: handler}
	go func() { log.Fatal(srv.ListenAndServe()) }()

	go debugserver.Start(debugserver.Endpoint{
		Name: "Repo Updater State",
		Path: "/repo-updater-state",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			d, err := json.MarshalIndent(scheduler.DebugDump(), "", "  ")
			if err != nil {
				http.Error(w, "failed to marshal snapshot: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(d)
		}),
	})

	select {}
}

func makeGitHubDotComSrc(ctx context.Context, store repos.Store, cf *httpcli.Factory) (*repos.GithubSource, error) {
	es, err := store.ListExternalServices(ctx, repos.StoreListExternalServicesArgs{
		Kinds: []string{"github"},
	})
	if err != nil {
		return nil, errors.Errorf("failed to list external services: %v", err)
	}

	var githubDotComSvc *repos.ExternalService
	for _, e := range es {
		cfg, err := e.Configuration()
		if err != nil {
			return nil, errors.Errorf("unable to get external service configuration: %v", err)
		}
		if cfg.(*schema.GitHubConnection).Token != "" {
			githubDotComSvc = e
			break
		}
	}

	if githubDotComSvc == nil {
		return nil, errors.Errorf("no external service for Github.com found")
	}

	return repos.NewGithubSource(githubDotComSvc, cf)
}

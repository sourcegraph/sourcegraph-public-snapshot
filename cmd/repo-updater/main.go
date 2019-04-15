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
	"strconv"
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
	"github.com/sourcegraph/sourcegraph/pkg/tracer"
)

const port = "3182"

func main() {
	syncerEnabled, _ := strconv.ParseBool(env.Get("SRC_SYNCER_ENABLED", "true", "Use the new repo metadata syncer."))

	ctx := context.Background()
	env.Lock()
	env.HandleHelpFlag()
	tracer.Init()

	clock := func() time.Time { return time.Now().UTC() }

	// Syncing relies on access to frontend and git-server, so wait until they started up.
	if err := api.InternalClient.WaitForFrontend(ctx); err != nil {
		log.Fatalf("sourcegraph-frontend not reachable: %v", err)
	}

	gitserver.DefaultClient.WaitForGitServers(ctx)

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
	db, err := repos.NewDB(dsn)
	if err != nil {
		log.Fatalf("failed to initialize db store: %v", err)
	}

	store := repos.NewDBStore(ctx, db, sql.TxOptions{Isolation: sql.LevelSerializable})

	cliFactory := repos.NewHTTPClientFactory()
	src := repos.NewSourcer(cliFactory)

	migrations := []repos.Migration{
		repos.GithubSetDefaultRepositoryQueryMigration(clock),
		repos.GitLabSetDefaultProjectQueryMigration(clock),
	}

	var kinds []string
	if syncerEnabled {
		kinds = append(kinds,
			"GITHUB",
			"GITLAB",
			"BITBUCKETSERVER",
			"OTHER",
		)
		migrations = append(migrations,
			repos.EnabledStateDeprecationMigration(src, clock, kinds...),
		)
	}

	for _, m := range migrations {
		if err := m.Run(ctx, store); err != nil {
			log.Fatalf("failed to run migration: %s", err)
		}
	}

	newSyncerEnabled := make(map[string]bool, len(kinds))
	for _, kind := range kinds {
		newSyncerEnabled[kind] = true
	}

	frontendAPI := repos.NewInternalAPI(10 * time.Second)

	for _, kind := range []string{
		"AWSCODECOMMIT",
		"BITBUCKETSERVER",
		"GITHUB",
		"GITLAB",
		"GITOLITE",
		"PHABRICATOR",
		"OTHER",
	} {
		if newSyncerEnabled[kind] {
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
			log15.Warn("Other external service kind only supported with SRC_SYNCER_ENABLED=true")
		default:
			log.Fatalf("unknown external service kind %q", kind)
		}
	}

	var syncer *repos.Syncer

	if syncerEnabled {
		diffs := make(chan repos.Diff)
		syncer = repos.NewSyncer(store, src, diffs, clock)

		log15.Info("starting new syncer", "external service kinds", kinds)
		go func() { log.Fatal(syncer.Run(ctx, repos.GetUpdateInterval(), kinds...)) }()

		// Start new repo syncer updates scheduler relay thread.
		go func() {
			for diff := range diffs {
				if !conf.Get().DisableAutoGitUpdates {
					repos.Scheduler.Update(diff.Repos()...)
				}
			}
		}()
	}

	// Git fetches scheduler
	go repos.RunScheduler(ctx)

	// git-server repos purging thread
	go repos.RunRepositoryPurgeWorker(ctx)

	// Start up handler that frontend relies on
	repoupdater := repoupdater.Server{
		Kinds:       kinds,
		Store:       store,
		Syncer:      syncer,
		InternalAPI: frontendAPI,
	}

	handler := nethttp.Middleware(opentracing.GlobalTracer(), repoupdater.Handler())
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
			d, err := json.MarshalIndent(repos.Scheduler.DebugDump(), "", "  ")
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

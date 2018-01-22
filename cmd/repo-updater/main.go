package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"gopkg.in/inconshreveable/log15.v2"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
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

	waitForFrontend(ctx)

	// Repos List syncing thread
	go func() {
		if err := repos.RunRepositorySyncWorker(ctx); err != nil {
			log.Fatalf("Fatal error RunRepositorySyncWorker: %s", err)
		}
	}()

	// GitHub Repository syncing thread
	go func() {
		if err := repos.RunGitHubRepositorySyncWorker(ctx); err != nil {
			log.Fatalf("Fatal error RunGitHubRepositorySyncWorker: %s", err)
		}
	}()

	// Phabricator Repository syncing thread
	go func() {
		if err := repos.RunPhabricatorRepositorySyncWorker(ctx); err != nil {
			log.Fatalf("Fatal error RunPhabricatorRepositorySyncworker: %s", err)
		}
	}()

	// Gitolite syncing thread
	go func() {
		if err := repos.RunGitoliteRepositorySyncWorker(ctx); err != nil {
			log.Fatalf("Fatal error RunGitoliteRepositorySyncWorker: %s", err)
		}
	}()

	// Listen on :3182 so that we pass k8s health checks (for forward-compat with
	// https://github.com/sourcegraph/infrastructure/pull/416, but don't actually do anything until
	// https://github.com/sourcegraph/sourcegraph/pull/9100 lands.
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.Error(w, "not implemented", http.StatusNotFound)
		}
	})
	log15.Info("repo-updater: listening", "addr", ":3182")
	srv := &http.Server{Addr: ":3182", Handler: handler}
	log.Fatal(srv.ListenAndServe())
}

func waitForFrontend(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := api.InternalClient.RetryPingUntilAvailable(ctx); err != nil {
		log15.Warn("frontend not available at startup (will periodically try to reconnect)", "err", err)
	}
}

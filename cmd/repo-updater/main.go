package main

import (
	"context"
	"log"
	"strconv"
	"sync"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/gitserver"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/tracer"
)

var (
	interval, _ = strconv.Atoi(env.Get("REPO_LIST_UPDATE_INTERVAL", "", "interval (in minutes) for checking code hosts (e.g. gitolite) for new repositories"))
)

func main() {
	ctx := context.Background()
	env.Lock()
	env.HandleHelpFlag()
	tracer.Init("repo-updater")
	gitserver.DefaultClient.NoCreds = true

	var wg sync.WaitGroup

	// Repos List syncing thread
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := repos.RunRepositorySyncWorker(ctx); err != nil {
			log.Fatalf("Fatal error RunRepositorySyncWorker: %s", err)
		}
	}()

	// GitHub Repository syncing thread
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := repos.RunGitHubRepositorySyncWorker(ctx); err != nil {
			log.Fatalf("Fatal error RunGitHubRepositorySyncWorker: %s", err)
		}
	}()

	// Phabricator Repository syncing thread
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := repos.RunPhabricatorRepositorySyncWorker(ctx); err != nil {
			log.Fatalf("Fatal error RunPhabricatorRepositorySyncworker: %s", err)
		}
	}()

	// Gitolite syncing thread
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := repos.RunGitoliteRepositorySyncWorker(ctx); err != nil {
			log.Fatalf("Fatal error RunGitoliteRepositorySyncWorker: %s", err)
		}
	}()

	wg.Wait()
}

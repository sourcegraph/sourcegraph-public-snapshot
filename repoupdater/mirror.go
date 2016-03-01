package repoupdater

import (
	"time"

	"golang.org/x/net/context"
	"gopkg.in/inconshreveable/log15.v2"
	"src.sourcegraph.com/sourcegraph/app/appconf"
	"src.sourcegraph.com/sourcegraph/events"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	sgxcli "src.sourcegraph.com/sourcegraph/sgx/cli"
)

func init() {
	sgxcli.ServeInit = append(sgxcli.ServeInit, func() {
		// If we're updating repos in the background, kick off the updates initially.
		if !appconf.Flags.DisableMirrorRepoBackgroundUpdate {
			events.RegisterListener(&mirrorRepoUpdater{})
		}
	})
}

type mirrorRepoUpdater struct{}

func (r *mirrorRepoUpdater) Scopes() []string {
	return []string{"app:repo-auto-cloner"}
}

func (r *mirrorRepoUpdater) Start(ctx context.Context) {
	go func() {
		for {
			err := r.mirrorRepos(ctx)
			if err != nil {
				log15.Error("Mirrored repos updater failed", "error", err)
				break
			}
		}
	}()
}

func (r *mirrorRepoUpdater) mirrorRepos(ctx context.Context) error {
	cl, err := sourcegraph.NewClientFromContext(ctx)
	if err != nil {
		return err
	}
	repos, err := cl.Repos.List(ctx, &sourcegraph.RepoListOptions{
		ListOptions: sourcegraph.ListOptions{
			PerPage: 100000,
		},
	})
	if err != nil {
		return err
	}
	for _, repo := range repos.Repos {
		if repo.Mirror {
			// Sleep a tiny bit longer than MirrorUpdateRate to avoid our
			// enqueue being no-op / hitting "was recently updated".
			time.Sleep(appconf.Flags.MirrorRepoUpdateRate + (200 * time.Millisecond))
			Enqueue(repo)
		}
	}
	return nil
}

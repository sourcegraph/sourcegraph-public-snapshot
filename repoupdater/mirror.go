package repoupdater

import (
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
		apiclient, err := sourcegraph.NewClientFromContext(ctx)
		if err != nil {
			log15.Error("mirrorRepoUpdater: could not create client", "error", err)
			return
		}
		repos, err := apiclient.Repos.List(ctx, &sourcegraph.RepoListOptions{
			ListOptions: sourcegraph.ListOptions{
				PerPage: 100000,
			},
		})
		if err != nil {
			log15.Error("mirrorRepoUpdater: could not list repos", "error", err)
			return
		}
		for _, repo := range repos.Repos {
			if repo.Mirror {
				RepoUpdater.enqueue(repo)
			}
		}
	}()
}

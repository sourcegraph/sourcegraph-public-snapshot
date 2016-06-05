package repoupdater

import (
	"time"

	"gopkg.in/inconshreveable/log15.v2"

	"github.com/jpillora/backoff"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/app/appconf"
	sgxcli "sourcegraph.com/sourcegraph/sourcegraph/cli/cli"
	"sourcegraph.com/sourcegraph/sourcegraph/services/events"
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
	return []string{"internal:repo-auto-cloner"}
}

func (r *mirrorRepoUpdater) Start(ctx context.Context) {
	go func() {
		b := &backoff.Backoff{
			Max:    time.Minute,
			Jitter: true,
		}
		for {
			err := r.mirrorRepos(ctx)
			if err != nil {
				d := b.Duration()
				log15.Error("Mirrored repos updater failed, sleeping before next try", "error", err, "sleep", d)
				time.Sleep(d)
				continue
			}
			b.Reset()
		}
	}()
}

func (r *mirrorRepoUpdater) mirrorRepos(ctx context.Context) error {
	cl, err := sourcegraph.NewClientFromContext(ctx)
	if err != nil {
		return err
	}
	repos, err := cl.Repos.List(ctx, &sourcegraph.RepoListOptions{
		// Only update public mirror repos in the background, as we cannot reliably identify
		// the auth token to use for updating private mirrors.
		// TODO: make it possible to background update private mirror repos.
		Type: "public",
		ListOptions: sourcegraph.ListOptions{
			PerPage: 100000,
		},
	})
	if err != nil {
		return err
	}
	hasMirror := false
	for _, repo := range repos.Repos {
		if repo.Mirror && !repo.Private {
			// Sleep a tiny bit longer than MirrorUpdateRate to avoid our
			// enqueue being no-op / hitting "was recently updated".
			time.Sleep(appconf.Flags.MirrorRepoUpdateRate + (200 * time.Millisecond))
			Enqueue(repo.ID, nil)
			hasMirror = true
		}
	}
	if !hasMirror {
		// If we don't have a mirror, lets sleep to prevent us spamming Repos.List
		time.Sleep(time.Minute)
	}
	return nil
}

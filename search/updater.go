package search

import (
	"math/rand"
	"time"

	"github.com/jpillora/backoff"

	"golang.org/x/net/context"
	"gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/conf/feature"
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
)

const (
	updateIntervalMinutes = 2 * 60
)

func BackgroundUpdateIndex(ctx context.Context) {
	if !feature.Features.GlobalSearch {
		return
	}
	// Sleep for a random interval before starting to avoid
	// "thundering herds" in case this app is deployed with replication.
	time.Sleep(time.Duration(rand.Intn(updateIntervalMinutes/2)) * time.Minute)
	go func() {
		b := &backoff.Backoff{
			Max:    time.Minute,
			Jitter: true,
		}
		for {
			err := updateSearchIndex(ctx)
			if err != nil {
				d := b.Duration()
				log15.Error("Search index updater failed, sleeping before next try", "error", err, "sleep", d)
				time.Sleep(d)
				continue
			}
			b.Reset()
		}
	}()
}

func updateSearchIndex(ctx context.Context) error {
	cl, err := sourcegraph.NewClientFromContext(ctx)
	if err != nil {
		return err
	}
	repos, err := cl.Repos.List(ctx, &sourcegraph.RepoListOptions{
		// Only update index for public mirror repos in the background, as we cannot
		// reliably identify the auth token to use for updating private mirrors.
		// TODO: make it possible to background update private mirror repos.
		Type: "public",
		ListOptions: sourcegraph.ListOptions{
			PerPage: 100000,
		},
	})
	if err != nil {
		return err
	}
	for _, repo := range repos.Repos {
		// Refresh one repo at a time to avoid long running db operations.
		repoSpec := repo.RepoSpec()
		_, err := cl.Search.RefreshIndex(ctx, &sourcegraph.SearchRefreshIndexOp{
			Repos: []*sourcegraph.RepoSpec{&repoSpec},
		})
		if err != nil {
			log15.Error("Search index updater failed to update repo", "repo", repo.URI, "error", err)
		}
	}

	time.Sleep(time.Duration(updateIntervalMinutes) * time.Minute)
	return nil
}

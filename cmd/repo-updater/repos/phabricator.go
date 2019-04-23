package repos

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/phabricator"
	"github.com/sourcegraph/sourcegraph/pkg/httpcli"
	"github.com/sourcegraph/sourcegraph/schema"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

// RunPhabricatorRepositorySyncWorker runs the worker that syncs repositories from Phabricator to Sourcegraph
func RunPhabricatorRepositorySyncWorker(ctx context.Context, s Store, cf httpcli.Factory) {
	for {
		phabs, err := s.ListExternalServices(ctx, StoreListExternalServicesArgs{
			Kinds: []string{"PHABRICATOR"},
		})

		if err != nil {
			log15.Error("unable to fetch Phabricator connections", "err", err)
		}

		for _, phab := range phabs {
			src, err := NewPhabricatorSource(phab, cf)
			if err != nil {
				log15.Error("failed to instantiate PhabricatorSource", "err", err)
				continue
			}

			repos, err := src.ListRepos(ctx)
			if err != nil {
				log15.Error("Error fetching Phabricator repos", "err", err)
				continue
			}

			err = updatePhabRepos(ctx, repos)
			if err != nil {
				log15.Error("Error updating Phabricator repos", "err", err)
				continue
			}

			cfg, err := phab.Configuration()
			if err != nil {
				log15.Error("failed to parse Phabricator config", "err", err)
				continue
			}

			phabricatorUpdateTime.WithLabelValues(
				cfg.(*schema.PhabricatorConnection).Url,
			).Set(float64(time.Now().Unix()))
		}

		time.Sleep(GetUpdateInterval())
	}
}

// updatePhabRepos ensures that all provided repositories exist in the phabricator_repos table.
func updatePhabRepos(ctx context.Context, repos []*Repo) error {
	for _, r := range repos {
		repo := r.Metadata.(*phabricator.Repo)
		err := api.InternalClient.PhabricatorRepoCreate(
			ctx,
			api.RepoName(r.Name),
			repo.Callsign,
			r.ExternalRepo.ServiceID,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

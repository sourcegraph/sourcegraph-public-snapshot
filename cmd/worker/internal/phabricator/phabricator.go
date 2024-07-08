package phabricator

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/phabricator"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/repos"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

const (
	tagID = "id"
)

var (
	phabricatorUpdateTime = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "src_phabricator_last_time_sync",
		Help: "The last time a comprehensive Phabricator sync finished",
	}, []string{tagID})
)

// NewRepositorySyncWorker runs the worker that syncs repositories from Phabricator to Sourcegraph.
func NewRepositorySyncWorker(ctx context.Context, db database.DB, logger log.Logger, s repos.Store) goroutine.BackgroundRoutine {
	cf := httpcli.NewExternalClientFactory(
		httpcli.NewLoggingMiddleware(logger),
	)

	return goroutine.NewPeriodicGoroutine(
		actor.WithInternalActor(ctx),
		goroutine.HandlerFunc(func(ctx context.Context) error {
			phabs, err := s.ExternalServiceStore().List(ctx, database.ExternalServicesListOptions{
				Kinds: []string{extsvc.KindPhabricator},
			})
			if err != nil {
				return errors.Wrap(err, "unable to fetch Phabricator connections")
			}

			var errs error

			for _, phab := range phabs {
				src, err := repos.NewPhabricatorSource(ctx, logger, phab, cf)
				if err != nil {
					errs = errors.Append(errs, errors.Wrap(err, "failed to instantiate PhabricatorSource"))
					continue
				}

				repos, err := repos.ListAll(ctx, src)
				if err != nil {
					errs = errors.Append(errs, errors.Wrap(err, "error fetching Phabricator repos"))
					continue
				}

				err = updatePhabRepos(ctx, db, repos)
				if err != nil {
					errs = errors.Append(errs, errors.Wrap(err, "error updating Phabricator repos"))
					continue
				}

				cfg, err := phab.Configuration(ctx)
				if err != nil {
					errs = errors.Append(errs, errors.Wrap(err, "failed to parse Phabricator config"))
					continue
				}

				phabricatorUpdateTime.WithLabelValues(
					cfg.(*schema.PhabricatorConnection).Url,
				).Set(float64(time.Now().Unix()))
			}

			return errs
		}),
		goroutine.WithName("repo-updater.phabricator-repository-syncer"),
		goroutine.WithDescription("periodically syncs repositories from Phabricator to Sourcegraph"),
		goroutine.WithIntervalFunc(func() time.Duration {
			return conf.RepoListUpdateInterval()
		}),
	)
}

// updatePhabRepos ensures that all provided repositories exist in the phabricator_repos table.
func updatePhabRepos(ctx context.Context, db database.DB, repos []*types.Repo) error {
	for _, r := range repos {
		repo := r.Metadata.(*phabricator.Repo)
		_, err := db.Phabricator().CreateOrUpdate(ctx, repo.Callsign, r.Name, r.ExternalRepo.ServiceID)
		if err != nil {
			return err
		}
	}
	return nil
}

package dependencies

import (
	"context"
	"strings"
	"sync"

	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go/log"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/store"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/types"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// Service encapsulates the resolution and persistence of dependencies at the repository and package levels.
type Service struct {
	dependenciesStore *store.Store
	lockfilesSvc      LockfilesService
	syncer            Syncer
	operations        *operations
}

func newService(depsStore *store.Store, lockfilesSvc LockfilesService, syncer Syncer, observationContext *observation.Context) *Service {
	return &Service{
		dependenciesStore: depsStore,
		lockfilesSvc:      lockfilesSvc,
		syncer:            syncer,
		operations:        newOperations(observationContext),
	}
}

// Dependencies resolves the (transitive) dependencies for a set of repository and revisions.
// Both the input repoRevs and the output dependencyRevs are a map from repository names to revspecs.
func (s *Service) Dependencies(ctx context.Context, repoRevs map[api.RepoName]types.RevSpecSet) (dependencyRevs map[api.RepoName]types.RevSpecSet, err error) {
	logFields := make([]log.Field, 0, 2)
	if len(repoRevs) == 1 {
		for repoName, revs := range repoRevs {
			revStrs := make([]string, 0, len(revs))
			for rev := range revs {
				revStrs = append(revStrs, string(rev))
			}

			logFields = append(logFields,
				log.String("repo", string(repoName)),
				log.String("revs", strings.Join(revStrs, ",")),
			)
		}
	} else {
		logFields = append(logFields, log.Int("repoRevs", len(repoRevs)))
	}

	ctx, endObservation := s.operations.dependencies.With(ctx, &err, observation.Args{LogFields: logFields})
	defer func() {
		endObservation(1, observation.Args{LogFields: []log.Field{
			log.Int("numDependencyRevs", len(dependencyRevs)),
		}})
	}()

	var mu sync.Mutex
	dependencyRevs = make(map[api.RepoName]types.RevSpecSet)

	sem := semaphore.NewWeighted(32)
	g, ctx := errgroup.WithContext(ctx)

	for repoName, revs := range repoRevs {
		for rev := range revs {
			repoName, rev := repoName, rev

			g.Go(func() error {
				if err := sem.Acquire(ctx, 1); err != nil {
					return err
				}
				defer sem.Release(1)

				deps, err := s.lockfilesSvc.ListDependencies(ctx, repoName, string(rev))
				if err != nil {
					return err
				}

				for _, dep := range deps {
					if err := sem.Acquire(ctx, 1); err != nil {
						return err
					}

					dep := dep

					g.Go(func() error {
						defer sem.Release(1)

						isNew, err := s.dependenciesStore.UpsertDependencyRepo(ctx, dep)
						if err != nil {
							return err
						}

						depName := dep.RepoName()
						depRev := api.RevSpec(dep.GitTagFromVersion())

						if isNew {
							if err := s.syncer.Sync(ctx, depName); err != nil {
								log15.Warn("failed to sync dependency repo", "repo", depName, "rev", depRev, "error", err)
							}
						}

						mu.Lock()
						defer mu.Unlock()

						if _, ok := dependencyRevs[depName]; !ok {
							dependencyRevs[depName] = types.RevSpecSet{}
						}
						dependencyRevs[depName][depRev] = struct{}{}

						return nil
					})
				}

				return nil
			})
		}
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return dependencyRevs, nil
}

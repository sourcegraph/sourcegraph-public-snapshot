package repos

import (
	"context"
	"sync"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"

	"github.com/sourcegraph/sourcegraph/internal/api"
	codeinteldbstore "github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/lockfiles"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// RevSpecSet is a utility type for a set of RevSpecs.
type RevSpecSet map[api.RevSpec]struct{}

type DependenciesResolver interface {
	Dependencies(ctx context.Context, repoRevs map[api.RepoName]RevSpecSet) (_ map[api.RepoName]RevSpecSet, err error)
}

type dependenciesResolver struct {
	db       database.DB
	listFunc func(context.Context, database.ExternalServicesListOptions) (dependencyRevs []*types.ExternalService, err error)
	syncFunc func(context.Context, int64) error
}

func NewDependenciesResolver(
	db database.DB,
	listFunc func(context.Context, database.ExternalServicesListOptions) (dependencyRevs []*types.ExternalService, err error),
	syncFunc func(context.Context, int64) error,
) *dependenciesResolver {
	return &dependenciesResolver{
		db:       db,
		listFunc: listFunc,
		syncFunc: syncFunc,
	}
}

// Dependencies resolves the (transitive) dependencies for a set of repository and revisions.
// Both the input repoRevs and the output dependencyRevs are a map from repository names to
// `40-char commit hashes belonging to that repository.
func (r *dependenciesResolver) Dependencies(ctx context.Context, repoRevs map[api.RepoName]RevSpecSet) (_ map[api.RepoName]RevSpecSet, err error) {
	depsStore := codeinteldbstore.NewDependencyInserter(r.db, r.listFunc, r.syncFunc)
	defer func() {
		if flushErr := depsStore.Flush(context.Background()); flushErr != nil {
			err = errors.Append(err, flushErr)
		}
	}()

	var (
		mu             sync.Mutex
		dependencyRevs = make(map[api.RepoName]RevSpecSet)
	)

	rg, ctx := errgroup.WithContext(ctx)
	svc := &lockfiles.Service{GitArchive: gitserver.DefaultClient.Archive}
	sem := semaphore.NewWeighted(16)

	for repoName, revs := range repoRevs {
		for rev := range revs {
			repoName, rev := repoName, rev

			rg.Go(func() error {
				if err := sem.Acquire(ctx, 1); err != nil {
					return err
				}
				defer sem.Release(1)

				deps, err := svc.ListDependencies(ctx, repoName, string(rev))
				if err != nil {
					return err
				}

				for _, dep := range deps {
					if err := depsStore.Insert(ctx, dep); err != nil {
						return err
					}

					depName := dep.RepoName()
					depRev := api.RevSpec(dep.GitTagFromVersion())

					mu.Lock()

					if _, ok := dependencyRevs[depName]; !ok {
						dependencyRevs[depName] = RevSpecSet{}
					}

					dependencyRevs[depName][depRev] = struct{}{}

					mu.Unlock()
				}

				return nil
			})
		}
	}

	if err := rg.Wait(); err != nil {
		return nil, err
	}

	return dependencyRevs, nil
}

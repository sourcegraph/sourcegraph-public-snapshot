package codeintel

import (
	"context"
	"database/sql"
	"sync"

	"github.com/sourcegraph/sourcegraph/internal/types"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"

	"github.com/sourcegraph/sourcegraph/internal/api"
	codeinteldbstore "github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/lockfiles"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

// RevSpecSet is a utility type for a set of RevSpecs.
type RevSpecSet map[api.RevSpec]struct{}

// DependenciesServices encapsulates the resolution and persistence of dependencies at the repository
// and package levels.
type DependenciesService struct {
	db              database.DB
	sync            func(context.Context, api.RepoName) (*types.Repo, error)
	lockfileService *lockfiles.Service
}

func NewDependenciesService(
	db database.DB,
	sync func(context.Context, api.RepoName) (*types.Repo, error),
) *DependenciesService {
	return &DependenciesService{
		db:              db,
		sync:            sync,
		lockfileService: &lockfiles.Service{GitArchive: gitserver.DefaultClient.Archive},
	}
}

// Dependencies resolves the (transitive) dependencies for a set of repository and revisions.
// Both the input repoRevs and the output dependencyRevs are a map from repository names to revspecs.
func (r *DependenciesService) Dependencies(ctx context.Context, repoRevs map[api.RepoName]RevSpecSet) (dependencyRevs map[api.RepoName]RevSpecSet, err error) {
	tr, ctx := trace.New(ctx, "DependenciesService", "Dependencies")
	defer func() {
		if len(repoRevs) > 1 {
			tr.LazyPrintf("repoRevsCount: %d", len(repoRevs))
		} else {
			tr.LazyPrintf("repoRevs: %v", repoRevs)
		}

		tr.LazyPrintf("depRevsCount: %d", len(dependencyRevs))
		tr.SetError(err)
		tr.Finish()
	}()

	var mu sync.Mutex
	dependencyRevs = make(map[api.RepoName]RevSpecSet)

	depsStore := codeinteldbstore.Store{Store: basestore.NewWithDB(r.db, sql.TxOptions{})}

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

				deps, err := r.lockfileService.ListDependencies(ctx, repoName, string(rev))
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

						if err := depsStore.UpsertDependencyRepo(ctx, dep); err != nil {
							return err
						}

						depName := dep.RepoName()
						if _, err := r.sync(ctx, depName); err != nil {
							return err
						}

						depRev := api.RevSpec(dep.GitTagFromVersion())

						mu.Lock()
						defer mu.Unlock()

						if _, ok := dependencyRevs[depName]; !ok {
							dependencyRevs[depName] = RevSpecSet{}
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

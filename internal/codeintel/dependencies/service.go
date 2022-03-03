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
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// Service encapsulates the resolution and persistence of dependencies at the repository and package levels.
type Service struct {
	dependenciesStore  *store.Store
	lockfilesSvc       LockfilesService
	lockfilesSemaphore *semaphore.Weighted
	syncer             Syncer
	syncerSemaphore    *semaphore.Weighted
	operations         *operations
}

var (
	lockfilesSemaphoreWeight = env.MustGetInt("CODEINTEL_DEPENDENCIES_LOCKFILES_SEMAPHORE_WEIGHT", 64, "The maximum number of concurrent routines parsing lockfile contents.")
	syncerSemaphoreWeight    = env.MustGetInt("CODEINTEL_DEPENDENCIES_LOCKFILES_SYNCER_WEIGHT", 64, "The maximum number of concurrent routines actively syncing repositories.")
)

func newService(depsStore *store.Store, lockfilesSvc LockfilesService, syncer Syncer, observationContext *observation.Context) *Service {
	return &Service{
		dependenciesStore:  depsStore,
		lockfilesSvc:       lockfilesSvc,
		lockfilesSemaphore: semaphore.NewWeighted(int64(lockfilesSemaphoreWeight)),
		syncer:             syncer,
		syncerSemaphore:    semaphore.NewWeighted(int64(syncerSemaphoreWeight)),
		operations:         newOperations(observationContext),
	}
}

// Dependencies resolves the (transitive) dependencies for a set of repository and revisions.
// Both the input repoRevs and the output dependencyRevs are a map from repository names to revspecs.
func (s *Service) Dependencies(ctx context.Context, repoRevs map[api.RepoName]types.RevSpecSet) (dependencyRevs map[api.RepoName]types.RevSpecSet, err error) {
	ctx, endObservation := s.operations.dependencies.With(ctx, &err, observation.Args{LogFields: constructLogFields(repoRevs)})
	defer func() {
		endObservation(1, observation.Args{LogFields: []log.Field{
			log.Int("numDependencyRevs", len(dependencyRevs)),
		}})
	}()

	// Parse lockfile contents for the given repository and revision pairs
	deps, err := s.lockfileDependencies(ctx, repoRevs)
	if err != nil {
		return nil, err
	}

	// Populate return value map from the given information. In the same pass, populate
	// auxiliary data structures that can be used to feed the upsert and sync operations
	// below.
	dependencyRevs = make(map[api.RepoName]types.RevSpecSet, len(repoRevs))
	dependencies := []store.DependencyRepo{}
	repoNamesByDependency := map[store.DependencyRepo]api.RepoName{}

	for _, dep := range deps {
		repo := dep.RepoName()
		rev := api.RevSpec(dep.GitTagFromVersion())
		scheme := dep.Scheme()
		name := dep.PackageSyntax()
		version := dep.PackageVersion()

		if _, ok := dependencyRevs[repo]; !ok {
			dependencyRevs[repo] = types.RevSpecSet{}
		}
		dependencyRevs[repo][rev] = struct{}{}

		dependencyRepo := store.DependencyRepo{Scheme: scheme, Name: name, Version: version}
		dependencies = append(dependencies, dependencyRepo)
		repoNamesByDependency[dependencyRepo] = repo
	}

	// Write depenencies to database and sync all of the ones that were newly inserted
	newDependencies, err := s.dependenciesStore.UpsertDependencyRepos(ctx, dependencies)
	if err != nil {
		return nil, err
	}
	if err := s.sync(ctx, newDependencies, repoNamesByDependency); err != nil {
		return nil, err
	}

	return dependencyRevs, nil
}

// lockfileDependencies returns a flattened list of package dependencies for every repo and
// revision pair specified in the given map.
func (s *Service) lockfileDependencies(ctx context.Context, repoRevs map[api.RepoName]types.RevSpecSet) ([]reposource.PackageDependency, error) {
	n := 0
	for _, revs := range repoRevs {
		n += len(revs)
	}
	var (
		closeOnce         sync.Once
		depsChan          = make(chan []reposource.PackageDependency, n)
		closeDepsChanOnce = func() { closeOnce.Do(func() { close(depsChan) }) }
	)
	defer closeDepsChanOnce()

	g, ctx := errgroup.WithContext(ctx)
	for repoName, revs := range repoRevs {
		for rev := range revs {
			// Capture outside of goroutine below
			repoName, rev := repoName, rev

			// Acquire semaphore before spawning goroutine to ensure that we limit the total number
			// of concurrent _routines_, whether they are actively processing lockfiles or not. Any
			// non-nil returned from here is a context timeout error, so we are guaranteed to clean
			// up the errgroup on exit.
			if err := s.lockfilesSemaphore.Acquire(ctx, 1); err != nil {
				return nil, err
			}

			g.Go(func() error {
				defer s.lockfilesSemaphore.Release(1)

				repoDeps, err := s.lockfilesSvc.ListDependencies(ctx, repoName, string(rev))
				if err != nil {
					return err
				}

				depsChan <- repoDeps
				return nil
			})
		}
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	// Writers exited - close channel so we can consume it to completion below
	closeDepsChanOnce()

	var deps []reposource.PackageDependency
	for batch := range depsChan {
		deps = append(deps, batch...)
	}

	return deps, nil
}

// sync calls sync on every repo in the supplied slice. It is assumed that for every value in the
// slice there is an associated value in the given map correlating a DependencyRepo struct to a repo
// name usable by the syncer.
func (s *Service) sync(ctx context.Context, newDependencies []store.DependencyRepo, repoNamesByDependency map[store.DependencyRepo]api.RepoName) error {
	g, ctx := errgroup.WithContext(ctx)
	for _, dep := range newDependencies {
		// Capture outside of goroutine below
		repo := repoNamesByDependency[dep]

		// Acquire semaphore before spawning goroutine to ensure that we limit the total number
		// of concurrent _routines_, whether they are actively syncing repo sources or not. Any
		// non-nil returned from here is a context timeout error, so we are guaranteed to clean
		// up the errgroup on exit.
		if err := s.syncerSemaphore.Acquire(ctx, 1); err != nil {
			return err
		}

		g.Go(func() error {
			defer s.syncerSemaphore.Release(1)

			if err := s.syncer.Sync(ctx, repo); err != nil {
				log15.Warn("Failed to sync dependency repo", "repo", repo, "error", err)
			}

			return nil
		})
	}

	return g.Wait()
}

func constructLogFields(repoRevs map[api.RepoName]types.RevSpecSet) []log.Field {
	if len(repoRevs) == 1 {
		for repoName, revs := range repoRevs {
			revStrs := make([]string, 0, len(revs))
			for rev := range revs {
				revStrs = append(revStrs, string(rev))
			}

			return []log.Field{
				log.String("repo", string(repoName)),
				log.String("revs", strings.Join(revStrs, ",")),
			}
		}
	}

	return []log.Field{
		log.Int("repoRevs", len(repoRevs)),
	}
}

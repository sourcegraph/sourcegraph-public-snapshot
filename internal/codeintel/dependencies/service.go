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
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/types"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Service encapsulates the resolution and persistence of dependencies at the repository and package levels.
type Service struct {
	dependenciesStore  Store
	gitSvc             localGitService
	lockfilesSvc       LockfilesService
	lockfilesSemaphore *semaphore.Weighted
	syncer             Syncer
	syncerSemaphore    *semaphore.Weighted
	operations         *operations
}

func newService(
	dependenciesStore Store,
	gitSvc localGitService,
	lockfilesSvc LockfilesService,
	lockfilesSemaphore *semaphore.Weighted,
	syncer Syncer,
	syncerSemaphore *semaphore.Weighted,
	observationContext *observation.Context,
) *Service {
	return &Service{
		dependenciesStore:  dependenciesStore,
		gitSvc:             gitSvc,
		lockfilesSvc:       lockfilesSvc,
		lockfilesSemaphore: lockfilesSemaphore,
		syncer:             syncer,
		syncerSemaphore:    syncerSemaphore,
		operations:         newOperations(observationContext),
	}
}

// Dependencies resolves the (transitive) dependencies for a set of repository and revisions.
// Both the input repoRevs and the output dependencyRevs are a map from repository names to revspecs.
func (s *Service) Dependencies(ctx context.Context, repoRevs map[api.RepoName]types.RevSpecSet) (dependencyRevs map[api.RepoName]types.RevSpecSet, err error) {
	ctx, _, endObservation := s.operations.dependencies.With(ctx, &err, observation.Args{LogFields: constructLogFields(repoRevs)})
	defer func() {
		endObservation(1, observation.Args{LogFields: []log.Field{
			log.Int("numDependencyRevs", len(dependencyRevs)),
		}})
	}()

	// Resolve the revhashes for the source repo-commit pairs
	repoCommits, err := s.resolveRepoCommits(ctx, repoRevs)
	if err != nil {
		return nil, err
	}

	// Parse lockfile contents for the given repository and revision pairs
	deps, err := s.lockfileDependencies(ctx, repoCommits)
	if err != nil {
		return nil, err
	}

	hash := func(dep Repo) string {
		return strings.Join([]string{dep.Scheme, dep.Name, dep.Version}, ":")
	}

	// Populate return value map from the given information. In the same pass, populate
	// auxiliary data structures that can be used to feed the upsert and sync operations
	// below.
	dependencyRevs = make(map[api.RepoName]types.RevSpecSet, len(repoRevs))
	dependencies := []Repo{}
	repoNamesByDependency := map[string]api.RepoName{}

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

		dep := Repo{Scheme: scheme, Name: name, Version: version}
		dependencies = append(dependencies, dep)
		repoNamesByDependency[hash(dep)] = repo
	}

	// Write dependencies to database
	newDependencies, err := s.dependenciesStore.UpsertDependencyRepos(ctx, dependencies)
	if err != nil {
		return nil, errors.Wrap(err, "store.UpsertDependencyRepos")
	}

	// Determine the set of repo names that were recently inserted. Package and repository
	// names are generally distinct, so we need to re-translate the dependency scheme, name,
	// and version back to the repository name.
	newRepos := make([]api.RepoName, 0, len(newDependencies))
	newReposSet := make(map[api.RepoName]struct{}, len(newDependencies))
	for _, dep := range newDependencies {
		repoName := repoNamesByDependency[hash(dep)]
		if _, ok := newReposSet[repoName]; ok {
			continue
		}

		newRepos = append(newRepos, repoName)
		newReposSet[repoName] = struct{}{}
	}

	// Lazily sync all the repos that were newly added
	if err := s.sync(ctx, newRepos); err != nil {
		return nil, err
	}

	return dependencyRevs, nil
}

type repoCommitResolvedCommit struct {
	api.RepoCommit
	ResolvedCommit string
}

// resolveRepoCommits flattens the given map into a slice of api.RepoCommits with an extra
// field indicating the canonical 40-character commit hash of the given revlike, which is
// often symbolic.
func (s *Service) resolveRepoCommits(ctx context.Context, repoRevs map[api.RepoName]types.RevSpecSet) ([]repoCommitResolvedCommit, error) {
	n := 0
	for _, revs := range repoRevs {
		n += len(revs)
	}

	repoCommits := make([]api.RepoCommit, 0, n)
	for repoName, revs := range repoRevs {
		for rev := range revs {
			repoCommits = append(repoCommits, api.RepoCommit{
				Repo:     repoName,
				CommitID: api.CommitID(rev),
			})
		}
	}

	commits, err := s.gitSvc.GetCommits(ctx, repoCommits, false)
	if err != nil {
		return nil, errors.Wrap(err, "git.GetCommits")
	}
	if len(commits) != len(repoCommits) {
		// Add assertion here so that the blast radius of new or newly discovered errors
		// southbound from the internal/vcs/git package does not leak into code intelligence.
		return nil, errors.Newf("expected slice returned from git.GetCommits to have len %d, but has len %d", len(repoCommits), len(commits))
	}

	resolvedCommits := make([]repoCommitResolvedCommit, 0, len(repoCommits))
	for i, repoCommit := range repoCommits {
		resolvedCommits = append(resolvedCommits, repoCommitResolvedCommit{
			RepoCommit:     repoCommit,
			ResolvedCommit: string(commits[i].ID),
		})
	}

	return resolvedCommits, nil
}

// lockfileDependencies returns a flattened list of package dependencies for every repo and
// revision pair specified in the given map.
func (s *Service) lockfileDependencies(ctx context.Context, repoCommits []repoCommitResolvedCommit) ([]reposource.PackageDependency, error) {
	var (
		mu   sync.Mutex
		deps []reposource.PackageDependency
	)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)
	for _, repoCommit := range repoCommits {
		// Capture outside of goroutine below
		repoName, rev := repoCommit.Repo, repoCommit.CommitID

		// Acquire semaphore before spawning goroutine to ensure that we limit the total number
		// of concurrent _routines_, whether they are actively processing lockfiles or not.
		if err := s.lockfilesSemaphore.Acquire(ctx, 1); err != nil {
			return nil, errors.Wrap(err, "lockfiles semaphore")
		}

		g.Go(func() error {
			defer s.lockfilesSemaphore.Release(1)

			repoDeps, err := s.lockfilesSvc.ListDependencies(ctx, repoName, string(rev))
			if err != nil {
				return errors.Wrap(err, "lockfiles.ListDependencies")
			}

			mu.Lock()
			deps = append(deps, repoDeps...)
			mu.Unlock()

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return deps, nil
}

// sync invokes the Syncer for every repo in the supplied slice.
func (s *Service) sync(ctx context.Context, repos []api.RepoName) error {
	g, ctx := errgroup.WithContext(ctx)
	for _, repo := range repos {
		// Capture outside of goroutine below
		repo := repo

		// Acquire semaphore before spawning goroutine to ensure that we limit the total number
		// of concurrent _routines_, whether they are actively syncing repo sources or not. Any
		// non-nil returned from here is a context timeout error, so we are guaranteed to clean
		// up the errgroup on exit.
		if err := s.syncerSemaphore.Acquire(ctx, 1); err != nil {
			return errors.Wrap(err, "syncer semaphore")
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

// Dependents resolves the (transitive) inverse dependencies for a set of repository and revisions.
// Both the input repoRevs and the output dependencyRevs are a map from repository names to revspecs.
func (s *Service) Dependents(ctx context.Context, repoRevs map[api.RepoName]types.RevSpecSet) (dependencyRevs map[api.RepoName]types.RevSpecSet, err error) {
	// To be implemented after #31643
	return nil, errors.New("unimplemented: dependencies.Dependents")
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

type Repo = shared.Repo

type ListDependencyReposOpts struct {
	Scheme      string
	Name        string
	After       int
	Limit       int
	NewestFirst bool
}

func (s *Service) ListDependencyRepos(ctx context.Context, opts ListDependencyReposOpts) ([]Repo, error) {
	return s.dependenciesStore.ListDependencyRepos(ctx, store.ListDependencyReposOpts(opts))
}

func (s *Service) UpsertDependencyRepos(ctx context.Context, deps []Repo) ([]Repo, error) {
	return s.dependenciesStore.UpsertDependencyRepos(ctx, deps)
}

func (s *Service) DeleteDependencyReposByID(ctx context.Context, ids ...int) error {
	return s.dependenciesStore.DeleteDependencyReposByID(ctx, ids...)
}

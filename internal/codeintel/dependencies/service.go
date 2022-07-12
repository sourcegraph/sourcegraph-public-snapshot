package dependencies

import (
	"context"
	"strings"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go/log"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/internal/lockfiles"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/types"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Service encapsulates the resolution and persistence of dependencies at the repository and package levels.
type Service struct {
	dependenciesStore  store.Store
	gitSvc             localGitService
	lockfilesSvc       LockfilesService
	lockfilesSemaphore *semaphore.Weighted
	syncer             Syncer
	syncerSemaphore    *semaphore.Weighted
	operations         *operations
}

func newService(
	dependenciesStore store.Store,
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
func (s *Service) Dependencies(ctx context.Context, repoRevs map[api.RepoName]types.RevSpecSet, includeTransitive bool) (dependencyRevs map[api.RepoName]types.RevSpecSet, notFound map[api.RepoName]types.RevSpecSet, err error) {
	ctx, _, endObservation := s.operations.dependencies.With(ctx, &err, observation.Args{LogFields: constructLogFields(repoRevs)})
	defer func() {
		endObservation(1, observation.Args{LogFields: []log.Field{
			log.Int("numDependencyRevs", len(dependencyRevs)),
		}})
	}()

	// Resolve the revhashes for the source repo-commit pairs.
	// TODO - Process unresolved commits.
	repoCommits, _, err := s.resolveRepoCommits(ctx, repoRevs)
	if err != nil {
		return nil, nil, err
	}

	// Load lockfile dependencies for the given repository and revision pairs
	deps, notFoundRepoCommits, err := s.resolveLockfileDependenciesFromStore(ctx, repoCommits, includeTransitive)
	if err != nil {
		return nil, nil, err
	}

	// Populate return value map from the given information.
	dependencyRevs = make(map[api.RepoName]types.RevSpecSet, len(repoRevs))
	for _, dep := range deps {
		repo := dep.RepoName()
		rev := api.RevSpec(dep.GitTagFromVersion())

		if _, ok := dependencyRevs[repo]; !ok {
			dependencyRevs[repo] = types.RevSpecSet{}
		}
		dependencyRevs[repo][rev] = struct{}{}
	}

	notFound = make(map[api.RepoName]types.RevSpecSet, len(notFoundRepoCommits))
	for _, repoCommit := range notFoundRepoCommits {
		repo := repoCommit.Repo
		// TODO: This is wrong. what we want is to find out which revspec we
		// couldn't find results for. So if the user is querying for
		// dependencies of foo@v1 we want to tell user that we couldn't find
		// anything for foo@v1 and not foo@d34db33f, if that is the commit that
		// v1 resolved to.
		rev := api.RevSpec(repoCommit.CommitID)

		if _, ok := notFound[repo]; !ok {
			notFound[repo] = types.RevSpecSet{}
		}
		notFound[repo][rev] = struct{}{}
	}

	if !enablePreciseQueries {
		return dependencyRevs, notFound, nil
	}

	for _, repoCommit := range repoCommits {
		// TODO - batch these requests in the store layer
		preciseDeps, err := s.dependenciesStore.PreciseDependencies(ctx, string(repoCommit.Repo), repoCommit.ResolvedCommit)
		if err != nil {
			return nil, nil, errors.Wrap(err, "store.PreciseDependencies")
		}

		for repoName, commits := range preciseDeps {
			if _, ok := dependencyRevs[repoName]; !ok {
				dependencyRevs[repoName] = types.RevSpecSet{}
			}
			for commit := range commits {
				dependencyRevs[repoName][commit] = struct{}{}
			}

			// If we found a precise result, we remove repoRev from notFound.
			if notFoundRevs, ok := notFound[repoCommit.Repo]; ok {
				for commit := range commits {
					delete(notFoundRevs, commit)
				}

				if len(notFoundRevs) == 0 {
					delete(notFound, repoCommit.Repo)
				}
			}
		}
	}

	return dependencyRevs, notFound, nil
}

type repoCommitResolvedCommit struct {
	api.RepoCommit
	ResolvedCommit string
}

// resolveRepoCommits flattens the given map into a slice of api.RepoCommits with an extra
// field indicating the canonical 40-character commit hash of the given revlike, which is
// often symbolic. The commits that failed to resolve are returned in a separate slice.
func (s *Service) resolveRepoCommits(ctx context.Context, repoRevs map[api.RepoName]types.RevSpecSet) ([]repoCommitResolvedCommit, []api.RepoCommit, error) {
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

	commits, err := s.gitSvc.GetCommits(ctx, repoCommits, true)
	if err != nil {
		return nil, nil, errors.Wrap(err, "git.GetCommits")
	}
	if len(commits) != len(repoCommits) {
		// Add assertion here so that the blast radius of new or newly discovered errors
		// southbound from the internal/vcs/git package does not leak into code intelligence.
		return nil, nil, errors.Newf("expected slice returned from git.GetCommits to have len %d, but has len %d", len(repoCommits), len(commits))
	}

	resolvedCommits := make([]repoCommitResolvedCommit, 0, len(repoCommits))
	var unresolvedCommits []api.RepoCommit
	for i, repoCommit := range repoCommits {
		if commits[i] == nil {
			unresolvedCommits = append(unresolvedCommits, repoCommit)
			continue
		}
		resolvedCommits = append(resolvedCommits, repoCommitResolvedCommit{
			RepoCommit:     repoCommit,
			ResolvedCommit: string(commits[i].ID),
		})
	}

	return resolvedCommits, unresolvedCommits, nil
}

// resolveLockfileDependenciesFromStore returns a flattened list of package dependencies for each
// of the given repo-commit pairs from the database. The given `repoCommits` slice is altered in-place.
// The returned `numUnqueried` value is the number of elements at the prefix of the slice that had no data.
// It is expected that the remaining elements be passed to the fallback dependencies resolver, if one is
// registered.
func (s *Service) resolveLockfileDependenciesFromStore(ctx context.Context, repoCommits []repoCommitResolvedCommit, includeTransitive bool) (deps []shared.PackageDependency, notFound []repoCommitResolvedCommit, err error) {
	ctx, _, endObservation := s.operations.resolveLockfileDependenciesFromStore.With(ctx, &err, observation.Args{})
	defer func() {
		endObservation(1, observation.Args{LogFields: []log.Field{
			log.Int("notFound", len(notFound)),
		}})
	}()

	for _, repoCommit := range repoCommits {
		// TODO - batch these requests in the store layer
		if repoDeps, ok, err := s.dependenciesStore.LockfileDependencies(ctx, store.LockfileDependenciesOpts{
			RepoName:          string(repoCommit.Repo),
			Commit:            repoCommit.ResolvedCommit,
			IncludeTransitive: includeTransitive,
		}); err != nil {
			return nil, notFound, errors.Wrap(err, "store.LockfileDependencies")
		} else if !ok {
			notFound = append(notFound, repoCommit)
		} else {
			deps = append(deps, repoDeps...)
		}
	}

	return deps, notFound, nil
}

// listAndPersistLockfileDependencies gathers dependencies from the lockfiles service for the
// given repo-commit pair and persists the result to the database. This aids in both caching
// and building an inverted index to power dependents search.
func (s *Service) listAndPersistLockfileDependencies(ctx context.Context, repoCommit repoCommitResolvedCommit) ([]shared.PackageDependency, error) {
	results, err := s.lockfilesSvc.ListDependencies(ctx, repoCommit.Repo, string(repoCommit.CommitID))
	if err != nil {
		return nil, errors.Wrap(err, "lockfiles.ListDependencies")
	}

	if len(results) == 0 {
		// If we haven't found anything in that repository, we still persist
		// the result to make sure we can tell the user that we have or haven't
		// indexed the repository.
		// TODO: There should be a better solution for that.
		results = append(results, lockfiles.Result{Lockfile: "NOT-FOUND", Deps: []reposource.VersionedPackage{}})
	}

	var (
		allDeps []shared.PackageDependency
		set     = make(map[string]struct{})
	)

	for _, result := range results {
		serializableRepoDeps := shared.SerializePackageDependencies(result.Deps)
		serializableGraph := shared.SerializeDependencyGraph(result.Graph)

		err = s.dependenciesStore.UpsertLockfileGraph(
			ctx,
			string(repoCommit.Repo),
			repoCommit.ResolvedCommit,
			result.Lockfile,
			serializableRepoDeps,
			serializableGraph,
		)
		if err != nil {
			return nil, errors.Wrap(err, "store.UpsertLockfileDependencies")
		}

		for _, d := range serializableRepoDeps {
			k := d.PackageSyntax() + d.PackageVersion()
			if _, ok := set[k]; !ok {
				set[k] = struct{}{}
				allDeps = append(allDeps, d)
			}
		}
	}

	return allDeps, nil
}

// sync invokes the Syncer for every repo in the supplied slice.
func (s *Service) sync(ctx context.Context, repos []api.RepoName) error {
	ctx, cancel := context.WithCancel(ctx)
	g, ctx := errgroup.WithContext(ctx)
	defer cancel()

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

// IndexLockfiles resolves the lockfile dependencies for a set of repository and revsisions
// and writes them the database.
//
// This method is expected to be used only from background routines controlling lockfile indexing
// scheduling. Additional users may impact the performance profile of the application as a whole.
func (s *Service) IndexLockfiles(ctx context.Context, repoRevs map[api.RepoName]types.RevSpecSet) (err error) {
	if !lockfileIndexingEnabled() {
		return nil
	}

	// Resolve the revhashes for the source repo-commit pairs
	repoCommits, _, err := s.resolveRepoCommits(ctx, repoRevs)
	if err != nil {
		return err
	}

	var allDependencies []shared.PackageDependency
	for _, repoCommit := range repoCommits {
		deps, err := s.listAndPersistLockfileDependencies(ctx, repoCommit)
		if err != nil {
			return err
		}
		allDependencies = append(allDependencies, deps...)
	}

	return s.upsertAndSyncDependencies(ctx, allDependencies)
}

func (s *Service) upsertAndSyncDependencies(ctx context.Context, deps []shared.PackageDependency) error {
	hash := func(dep Repo) string {
		return strings.Join([]string{dep.Scheme, dep.Name, dep.Version}, ":")
	}

	dependencies := make([]Repo, 0, len(deps))
	repoNamesByDependency := make(map[string]api.RepoName, len(deps))

	for _, dep := range deps {
		repo := dep.RepoName()
		scheme := dep.Scheme()
		name := dep.PackageSyntax()
		version := dep.PackageVersion()

		dep := Repo{Scheme: scheme, Name: name, Version: version}
		dependencies = append(dependencies, dep)
		repoNamesByDependency[hash(dep)] = repo
	}

	// Write dependencies to database
	newDependencies, err := s.dependenciesStore.UpsertDependencyRepos(ctx, dependencies)
	if err != nil {
		return errors.Wrap(err, "store.UpsertDependencyRepos")
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
	return s.sync(ctx, newRepos)
}

// Dependents resolves the (transitive) inverse dependencies for a set of repository and revisions.
// Both the input repoRevs and the output dependencyRevs are a map from repository names to revspecs.
func (s *Service) Dependents(ctx context.Context, repoRevs map[api.RepoName]types.RevSpecSet) (dependentsRevs map[api.RepoName]types.RevSpecSet, err error) {
	// Resolve the revhashes for the source repo-commit pairs.
	// TODO - Process unresolved commits.
	repoCommits, _, err := s.resolveRepoCommits(ctx, repoRevs)
	if err != nil {
		return nil, err
	}

	var deps []api.RepoCommit
	for _, commit := range repoCommits {
		// TODO - batch these requests in the store layer
		repoDeps, err := s.dependenciesStore.LockfileDependents(ctx, string(commit.Repo), commit.ResolvedCommit)
		if err != nil {
			return nil, errors.Wrap(err, "store.LockfileDependents")
		}
		deps = append(deps, repoDeps...)

	}

	dependentsRevs = map[api.RepoName]types.RevSpecSet{}
	for _, dep := range deps {
		if _, ok := dependentsRevs[dep.Repo]; !ok {
			dependentsRevs[dep.Repo] = types.RevSpecSet{}
		}
		dependentsRevs[dep.Repo][api.RevSpec(dep.CommitID)] = struct{}{}
	}

	if !enablePreciseQueries {
		return dependentsRevs, nil
	}

	for _, repoCommit := range repoCommits {
		// TODO - batch these requests in the store layer
		preciseDeps, err := s.dependenciesStore.PreciseDependents(ctx, string(repoCommit.Repo), repoCommit.ResolvedCommit)
		if err != nil {
			return nil, errors.Wrap(err, "store.PreciseDependents")
		}

		for repoName, commits := range preciseDeps {
			if _, ok := dependentsRevs[repoName]; !ok {
				dependentsRevs[repoName] = types.RevSpecSet{}
			}
			for commit := range commits {
				dependentsRevs[repoName][commit] = struct{}{}
			}
		}
	}

	return dependentsRevs, nil
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

func (s *Service) SelectRepoRevisionsToResolve(ctx context.Context, batchSize int, minimumCheckInterval time.Duration) (map[string][]string, error) {
	return s.dependenciesStore.SelectRepoRevisionsToResolve(ctx, batchSize, minimumCheckInterval)
}

func (s *Service) UpdateResolvedRevisions(ctx context.Context, repoRevsToResolvedRevs map[string]map[string]string) error {
	return s.dependenciesStore.UpdateResolvedRevisions(ctx, repoRevsToResolvedRevs)
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

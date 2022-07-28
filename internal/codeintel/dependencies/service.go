package dependencies

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/opentracing/opentracing-go/log"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"

	sglog "github.com/sourcegraph/log"

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
	logger             sglog.Logger
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
		logger:             observationContext.Logger,
	}
}

// QueryParams can be used to query the dependencies or the dependents of the
// given revspecs in a repository.
type QueryParams struct {
	// Repo is the repository in which to search for dependency relationships.
	Repo api.RepoName
	// IncludeTransitive determines whether or not transitive dependencies are
	// included in response. It's ignored by Dependents right now.
	IncludeTransitive bool
	// RevSpecs is the set of revspecs over which to query.
	RevSpecs types.RevSpecSet
}

// resolvedQueryParams is used internally to represent flattened QueryParams
// whose RevSpecs have been resolved by a call to gitserver to a concrete
// ResolvedCommit. One QueryParams can be resolved into N resolvedQueryParams,
// where N is maximum len(queryParams.RevSpecs).
type resolvedQueryParams struct {
	Repo              api.RepoName
	IncludeTransitive bool

	// RevSpec is the original revspec as defined in QueryParams.RevSpec.
	RevSpec api.RevSpec
	// ResolvedCommit is the commit into which RevSpec has been resolved.
	ResolvedCommit string
}

// Dependencies resolves the (transitive) dependencies for a set of repository and revisions.
// The output dependencyRevs are a map from repository names to revspecs.
func (s *Service) Dependencies(ctx context.Context, params []QueryParams) (dependencyRevs map[api.RepoName]types.RevSpecSet, notFound map[api.RepoName]types.RevSpecSet, err error) {
	ctx, _, endObservation := s.operations.dependencies.With(ctx, &err, observation.Args{LogFields: constructLogFields(params)})
	defer func() {
		endObservation(1, observation.Args{LogFields: []log.Field{
			log.Int("numDependencyRevs", len(dependencyRevs)),
		}})
	}()

	// Resolve the revhashes for the source repo-commit pairs.
	// TODO - Process unresolved commits.
	resolvedParams, _, err := s.resolveDependenciesParams(ctx, params)
	if err != nil {
		return nil, nil, err
	}

	// Load lockfile dependencies for the given repository and revision pairs
	deps, notFoundParams, err := s.resolveLockfileDependenciesFromStore(ctx, resolvedParams)
	if err != nil {
		return nil, nil, err
	}

	// Populate return value map from the given information.
	dependencyRevs = make(map[api.RepoName]types.RevSpecSet, len(params))
	for _, dep := range deps {
		repo := dep.RepoName()
		rev := api.RevSpec(dep.GitTagFromVersion())

		if _, ok := dependencyRevs[repo]; !ok {
			dependencyRevs[repo] = types.RevSpecSet{}
		}
		dependencyRevs[repo][rev] = struct{}{}
	}

	notFound = make(map[api.RepoName]types.RevSpecSet, len(notFoundParams))
	for _, param := range notFoundParams {
		if _, ok := notFound[param.Repo]; !ok {
			notFound[param.Repo] = types.RevSpecSet{}
		}
		notFound[param.Repo][param.RevSpec] = struct{}{}
	}

	if !enablePreciseQueries {
		return dependencyRevs, notFound, nil
	}

	for _, param := range resolvedParams {
		// TODO - batch these requests in the store layer
		preciseDeps, err := s.dependenciesStore.PreciseDependencies(ctx, string(param.Repo), param.ResolvedCommit)
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
			if notFoundRevs, ok := notFound[param.Repo]; ok {
				for commit := range commits {
					delete(notFoundRevs, commit)
				}

				if len(notFoundRevs) == 0 {
					delete(notFound, param.Repo)
				}
			}
		}
	}

	return dependencyRevs, notFound, nil
}

// resolveDependenciesParams flattens the given map into a slice of api.RepoCommits with two extra fields:
//
// - ResolvedCommit: the canonical 40-character commit hash of the given revlike, which is often symbolic.
// - IncludeTransitive: the boolean from the original DependenciesParam from the revlike originated.
//
// The commits that failed to resolve are returned in a separate slice.
func (s *Service) resolveDependenciesParams(ctx context.Context, params []QueryParams) ([]resolvedQueryParams, []api.RepoCommit, error) {
	n := 0
	for _, p := range params {
		n += len(p.RevSpecs)
	}

	repoCommits := make([]api.RepoCommit, 0, n)
	// repoCommitToParam maps repoCommit entries to the entry in params from
	// which they originate.
	// Example: param corresponding to repoCommits[5] is params[repoCommitToParam[5]].
	repoCommitToParam := make([]int, 0, n)
	for i, p := range params {
		for rev := range p.RevSpecs {
			repoCommits = append(repoCommits, api.RepoCommit{
				Repo:     p.Repo,
				CommitID: api.CommitID(rev),
			})
			repoCommitToParam = append(repoCommitToParam, i)
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

	resolvedParams := make([]resolvedQueryParams, 0, len(repoCommits))
	var unresolvedCommits []api.RepoCommit
	for i, repoCommit := range repoCommits {
		if commits[i] == nil {
			unresolvedCommits = append(unresolvedCommits, repoCommit)
			continue
		}

		param := params[repoCommitToParam[i]]
		resolvedParams = append(resolvedParams, resolvedQueryParams{
			Repo:              param.Repo,
			IncludeTransitive: param.IncludeTransitive,
			RevSpec:           api.RevSpec(repoCommit.CommitID),
			ResolvedCommit:    string(commits[i].ID),
		})
	}

	return resolvedParams, unresolvedCommits, nil
}

// resolveLockfileDependenciesFromStore returns a flattened list of package dependencies for each
// of the given repo-commit pairs from the database.
func (s *Service) resolveLockfileDependenciesFromStore(ctx context.Context, params []resolvedQueryParams) (deps []shared.PackageDependency, notFound []resolvedQueryParams, err error) {
	ctx, _, endObservation := s.operations.resolveLockfileDependenciesFromStore.With(ctx, &err, observation.Args{})
	defer func() {
		endObservation(1, observation.Args{LogFields: []log.Field{
			log.Int("notFound", len(notFound)),
		}})
	}()

	for _, param := range params {
		// TODO - batch these requests in the store layer
		if repoDeps, ok, err := s.dependenciesStore.LockfileDependencies(ctx, store.LockfileDependenciesOpts{
			RepoName:          string(param.Repo),
			Commit:            param.ResolvedCommit,
			IncludeTransitive: param.IncludeTransitive,
		}); err != nil {
			return nil, notFound, errors.Wrap(err, "store.LockfileDependencies")
		} else if !ok {
			notFound = append(notFound, param)
		} else {
			deps = append(deps, repoDeps...)
		}
	}

	return deps, notFound, nil
}

// listAndPersistLockfileDependencies gathers dependencies from the lockfiles service for the
// given repo-commit pair and persists the result to the database. This aids in both caching
// and building an inverted index to power dependents search.
func (s *Service) listAndPersistLockfileDependencies(ctx context.Context, param resolvedQueryParams) ([]shared.PackageDependency, error) {
	results, err := s.lockfilesSvc.ListDependencies(ctx, param.Repo, param.ResolvedCommit)
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
			string(param.Repo),
			param.ResolvedCommit,
			result.Lockfile,
			serializableRepoDeps,
			serializableGraph,
		)
		if err != nil {
			return nil, errors.Wrap(err, "store.UpsertLockfileGraph")
		}

		for _, d := range serializableRepoDeps {
			k := fmt.Sprintf("%s%s", d.PackageSyntax(), d.PackageVersion())
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
				s.logger.Warn("Failed to sync dependency repo", sglog.String("repo", string(repo)), sglog.Error(err))
			}

			return nil
		})
	}

	return g.Wait()
}

// IndexLockfiles resolves the lockfile dependencies for a set of repository and revisions
// and writes them the database.
//
// This method is expected to be used only from background routines controlling lockfile indexing
// scheduling. Additional users may impact the performance profile of the application as a whole.
//
// It ignores most fields of QueryParams and only uses Repo and the RevSpecs.
func (s *Service) IndexLockfiles(ctx context.Context, params QueryParams) (err error) {
	if !lockfileIndexingEnabled() {
		return nil
	}

	// Resolve the revhashes for the source repo-commit pairs
	resolvedParams, _, err := s.resolveDependenciesParams(ctx, []QueryParams{params})
	if err != nil {
		return err
	}

	var allDependencies []shared.PackageDependency
	for _, param := range resolvedParams {
		deps, err := s.listAndPersistLockfileDependencies(ctx, param)
		if err != nil {
			return err
		}
		allDependencies = append(allDependencies, deps...)
	}

	return s.upsertAndSyncDependencies(ctx, allDependencies)
}

func (s *Service) upsertAndSyncDependencies(ctx context.Context, deps []shared.PackageDependency) error {
	hash := func(dep Repo) string {
		return strings.Join([]string{dep.Scheme, string(dep.Name), dep.Version}, ":")
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
func (s *Service) Dependents(ctx context.Context, params []QueryParams) (dependentsRevs map[api.RepoName]types.RevSpecSet, err error) {
	// Resolve the revhashes for the source repo-commit pairs.
	// TODO - Process unresolved commits.
	resolvedParams, _, err := s.resolveDependenciesParams(ctx, params)
	if err != nil {
		return nil, err
	}

	var deps []api.RepoCommit
	for _, commit := range resolvedParams {
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

	for _, repoCommit := range resolvedParams {
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

func constructLogFields(params []QueryParams) (fields []log.Field) {
	for i, repoParam := range params {
		revStrs := make([]string, 0, len(repoParam.RevSpecs))
		for rev := range repoParam.RevSpecs {
			revStrs = append(revStrs, string(rev))
		}

		fields = append(fields,
			log.String(fmt.Sprintf("repo-%d", i), string(repoParam.Repo)),
			log.String(fmt.Sprintf("revs-%d", i), strings.Join(revStrs, ",")),
		)
	}

	return fields
}

func (s *Service) SelectRepoRevisionsToResolve(ctx context.Context, batchSize int, minimumCheckInterval time.Duration) (map[string][]string, error) {
	return s.dependenciesStore.SelectRepoRevisionsToResolve(ctx, batchSize, minimumCheckInterval)
}

func (s *Service) UpdateResolvedRevisions(ctx context.Context, repoRevsToResolvedRevs map[string]map[string]string) error {
	return s.dependenciesStore.UpdateResolvedRevisions(ctx, repoRevsToResolvedRevs)
}

type Repo = shared.Repo

type ListDependencyReposOpts struct {
	// Scheme is the moniker scheme to filter for e.g. 'gomod', 'npm' etc.
	Scheme string
	// Name is the package name to filter for e.g. '@types/node' etc.
	Name reposource.PackageName
	// After is the value predominantly used for pagination. When sorting by
	// newest first, this should be the ID of the last element in the previous
	// page, when excluding versions it should be the last package name in the
	// previous page.
	After any
	// Limit limits the size of the results set to be returned.
	Limit int
	// NewestFirst sorts by when a (package, version) was added to the list.
	// Incompatible with ExcludeVersions below.
	NewestFirst bool
	// ExcludeVersions returns one row for every package, instead of one for
	// every (package, version) tuple. Results will be sorted by name to make
	// pagination possible. Takes precedence over NewestFirst.
	ExcludeVersions bool
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

type ListLockfileIndexesOpts struct {
	RepoName string
	Commit   string
	Lockfile string

	After int
	Limit int
}

func (s *Service) ListLockfileIndexes(ctx context.Context, opts ListLockfileIndexesOpts) ([]shared.LockfileIndex, int, error) {
	return s.dependenciesStore.ListLockfileIndexes(ctx, store.ListLockfileIndexesOpts(opts))
}

type GetLockfileIndexOpts struct {
	ID       int
	RepoName string
	Commit   string
	Lockfile string
}

func (s *Service) GetLockfileIndexOpts(ctx context.Context, opts GetLockfileIndexOpts) (shared.LockfileIndex, error) {
	return s.dependenciesStore.GetLockfileIndex(ctx, store.GetLockfileIndexOpts(opts))
}

func (s *Service) DeleteLockfileIndexByID(ctx context.Context, id int) error {
	return s.dependenciesStore.DeleteLockfileIndexByID(ctx, id)
}

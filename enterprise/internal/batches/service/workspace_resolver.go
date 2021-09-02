package service

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/cockroachdb/errors"
	"github.com/gobwas/glob"
	"github.com/hashicorp/go-multierror"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	streamapi "github.com/sourcegraph/sourcegraph/internal/search/streaming/api"
	streamhttp "github.com/sourcegraph/sourcegraph/internal/search/streaming/http"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
	"github.com/sourcegraph/sourcegraph/lib/batches/template"
)

type workspaceResolver struct {
	store               *store.Store
	frontendInternalURL string
}

func (wr *workspaceResolver) ResolveWorkspacesForBatchSpec(ctx context.Context, batchSpec *batcheslib.BatchSpec, opts ResolveWorkspacesForBatchSpecOpts) (_ []*RepoWorkspace, err error) {
	seen, unsupported, err := wr.determineRepositories(ctx, batchSpec, opts)
	if err != nil {
		return nil, err
	}

	all, ignored, err := filterIgnoredRepositories(ctx, seen, opts.AllowIgnored, unsupported)
	if err != nil {
		return nil, err
	}

	final, err := findWorkspaces(ctx, batchSpec, wr, all)
	if err != nil {
		return nil, err
	}

	if unsupported.HasUnsupported() {
		return final, unsupported
	}

	if ignored.HasIgnored() {
		return final, ignored
	}

	return final, nil
}

func (wr *workspaceResolver) determineRepositories(ctx context.Context, batchSpec *batcheslib.BatchSpec, opts ResolveWorkspacesForBatchSpecOpts) (map[api.RepoID]*RepoRevision, UnsupportedRepoSet, error) {
	seen := map[api.RepoID]*RepoRevision{}
	unsupported := UnsupportedRepoSet{}

	// TODO: this could be trivially parallelised in the future.
	for _, on := range batchSpec.On {
		// TODO: Use multierror.
		repos, err := wr.resolveRepositoriesOn(ctx, &on)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "resolving %q", on.String())
		}

		for _, repo := range repos {
			// Skip repos where no branch exists.
			if !repo.HasBranch() {
				continue
			}

			if other, ok := seen[repo.Repo.ID]; !ok {
				seen[repo.Repo.ID] = repo

				if !btypes.IsKindSupported(extsvc.TypeToKind(repo.Repo.ExternalRepo.ServiceType)) {
					if !opts.AllowUnsupported {
						unsupported.Append(repo.Repo)
					}
				}
			} else {
				// If we've already seen this repository, we overwrite the
				// Commit/Branch fields with the latest value we have
				other.Commit = repo.Commit
				other.Branch = repo.Branch
			}
		}
	}

	return seen, unsupported, nil
}

func filterIgnoredRepositories(
	ctx context.Context,
	repos map[api.RepoID]*RepoRevision,
	allowIgnored bool,
	unsupported UnsupportedRepoSet,
) ([]*RepoRevision, IgnoredRepoSet, error) {
	type result struct {
		repo           *RepoRevision
		hasBatchIgnore bool
		err            error
	}

	var (
		final   = make([]*RepoRevision, 0, len(repos))
		ignored = IgnoredRepoSet{}

		input   = make(chan *RepoRevision, len(repos))
		results = make(chan result, len(repos))

		wg sync.WaitGroup
	)

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(in chan *RepoRevision, out chan result) {
			defer wg.Done()
			for repo := range in {
				hasBatchIgnore, err := hasBatchIgnoreFile(ctx, repo)
				results <- result{repo, hasBatchIgnore, err}
			}
		}(input, results)
	}

	for _, repo := range repos {
		input <- repo
	}
	close(input)

	go func(wg *sync.WaitGroup) {
		wg.Wait()
		close(results)
	}(&wg)

	var errs *multierror.Error
	for result := range results {
		if result.err != nil {
			errs = multierror.Append(errs, result.err)
			continue
		}

		if !allowIgnored && result.hasBatchIgnore {
			ignored.Append(result.repo.Repo)
		}

		if !unsupported.Includes(result.repo.Repo) && !ignored.Includes(result.repo.Repo) {
			final = append(final, result.repo)
		}
	}

	return final, ignored, errs.ErrorOrNil()
}

var ErrMalformedOnQueryOrRepository = batcheslib.NewValidationError(errors.New("malformed 'on' field; missing either a repository name or a query"))

func (wr *workspaceResolver) resolveRepositoriesOn(ctx context.Context, on *batcheslib.OnQueryOrRepository) (_ []*RepoRevision, err error) {
	tr, ctx := trace.New(ctx, "workspaceResolver.resolveRepositoriesOn", "")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	if on.RepositoriesMatchingQuery != "" {
		return wr.resolveRepositoriesMatchingQuery(ctx, on.RepositoriesMatchingQuery)
	}

	if on.Repository != "" && on.Branch != "" {
		repo, err := wr.resolveRepositoryNameAndBranch(ctx, on.Repository, on.Branch)
		if err != nil {
			return nil, err
		}
		return []*RepoRevision{repo}, nil
	}

	if on.Repository != "" {
		repo, err := wr.resolveRepositoryName(ctx, on.Repository)
		if err != nil {
			return nil, err
		}
		return []*RepoRevision{repo}, nil
	}

	// This shouldn't happen on any batch spec that has passed validation, but,
	// alas, software.
	return nil, ErrMalformedOnQueryOrRepository
}

func (wr *workspaceResolver) resolveRepositoryName(ctx context.Context, name string) (_ *RepoRevision, err error) {
	tr, ctx := trace.New(ctx, "workspaceResolver.resolveRepositoryName", "")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	repo, err := wr.store.Repos().GetByName(ctx, api.RepoName(name))
	if err != nil {
		return nil, err
	}

	return repoToRepoRevisionWithDefaultBranch(ctx, repo)
}

func (wr *workspaceResolver) resolveRepositoryNameAndBranch(ctx context.Context, name, branch string) (_ *RepoRevision, err error) {
	tr, ctx := trace.New(ctx, "workspaceResolver.resolveRepositoryNameAndBranch", "")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	repo, err := wr.store.Repos().GetByName(ctx, api.RepoName(name))
	if err != nil {
		return nil, err
	}

	commit, err := git.ResolveRevision(ctx, repo.Name, branch, git.ResolveRevisionOptions{
		NoEnsureRevision: true,
	})
	if err != nil && errors.HasType(err, &gitserver.RevisionNotFoundError{}) {
		return nil, fmt.Errorf("no branch matching %q found for repository %s", branch, name)
	}

	return &RepoRevision{
		Repo:   repo,
		Branch: branch,
		Commit: commit,
	}, nil
}

func (wr *workspaceResolver) resolveRepositoriesMatchingQuery(ctx context.Context, query string) (_ []*RepoRevision, err error) {
	tr, ctx := trace.New(ctx, "workspaceResolver.resolveRepositorySearch", "")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	query = setDefaultQueryCount(query)
	query = setDefaultQuerySelect(query)

	repoIDs := []api.RepoID{}
	wr.runSearch(ctx, query, func(matches []streamhttp.EventMatch) {
		for _, match := range matches {
			switch m := match.(type) {
			case *streamhttp.EventRepoMatch:
				repoIDs = append(repoIDs, api.RepoID(m.RepositoryID))
			case *streamhttp.EventContentMatch:
				repoIDs = append(repoIDs, api.RepoID(m.RepositoryID))
			case *streamhttp.EventPathMatch:
				repoIDs = append(repoIDs, api.RepoID(m.RepositoryID))
			case *streamhttp.EventSymbolMatch:
				repoIDs = append(repoIDs, api.RepoID(m.RepositoryID))
			}
		}
	})

	// 🚨 SECURITY: We use database.Repos.List to check whether the user has access to
	// the repositories or not.
	accessibleRepos, err := wr.store.Repos().List(ctx, database.ReposListOptions{IDs: repoIDs})
	if err != nil {
		return nil, err
	}

	revs := make([]*RepoRevision, 0, len(accessibleRepos))
	for _, repo := range accessibleRepos {
		rev, err := repoToRepoRevisionWithDefaultBranch(ctx, repo)
		if err != nil {
			return nil, err
		}
		revs = append(revs, rev)
	}

	return revs, nil
}

const internalSearchClientUserAgent = "Batch Changes repository resolver"

func (wr *workspaceResolver) runSearch(ctx context.Context, query string, onMatches func(matches []streamhttp.EventMatch)) (err error) {
	req, err := streamhttp.NewRequest(wr.frontendInternalURL, query)
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)

	// We don't set an auth token here and don't authenticate on the users
	// behalf in any way, because we will fetch the repositories from the
	// database later and check for repository permissions that way.
	req.Header.Set("User-Agent", internalSearchClientUserAgent)

	resp, err := httpcli.InternalClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	dec := streamhttp.FrontendStreamDecoder{
		OnMatches: onMatches,
		OnError: func(ee *streamhttp.EventError) {
			err = errors.New(ee.Message)
		},
		OnProgress: func(p *streamapi.Progress) {
			// TODO: Evaluate skipped for values we care about.
		},
	}
	return dec.ReadAll(resp.Body)
}

func repoToRepoRevisionWithDefaultBranch(ctx context.Context, repo *types.Repo) (_ *RepoRevision, err error) {
	tr, ctx := trace.New(ctx, "repoToRepoRevision", "")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	branch, commit, err := git.GetDefaultBranch(ctx, repo.Name)
	if err != nil {
		return nil, err
	}

	repoRev := &RepoRevision{
		Repo:   repo,
		Branch: branch,
		Commit: commit,
	}
	return repoRev, nil
}

func hasBatchIgnoreFile(ctx context.Context, r *RepoRevision) (_ bool, err error) {
	traceTitle := fmt.Sprintf("RepoID: %q", r.Repo.ID)
	tr, ctx := trace.New(ctx, "hasBatchIgnoreFile", traceTitle)
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	const path = ".batchignore"
	stat, err := git.Stat(ctx, r.Repo.Name, r.Commit, path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	if !stat.Mode().IsRegular() {
		return false, errors.Errorf("not a blob: %q", path)
	}
	return true, nil
}

var defaultQueryCountRegex = regexp.MustCompile(`\bcount:(\d+|all)\b`)

const hardCodedCount = " count:all"

func setDefaultQueryCount(query string) string {
	if defaultQueryCountRegex.MatchString(query) {
		return query
	}

	return query + hardCodedCount
}

var selectRegex = regexp.MustCompile(`\bselect:(.+)\b`)

const hardCodedSelectRepo = " select:repo"

func setDefaultQuerySelect(query string) string {
	if selectRegex.MatchString(query) {
		return query
	}

	return query + hardCodedSelectRepo
}

// TODO(mrnugget): Merge these two types (give them an "errorfmt" function,
// rename "Has*" methods to "NotEmpty" or something)

// UnsupportedRepoSet provides a set to manage repositories that are on
// unsupported code hosts. This type implements error to allow it to be
// returned directly as an error value if needed.
type UnsupportedRepoSet map[*types.Repo]struct{}

func (e UnsupportedRepoSet) Includes(r *types.Repo) bool {
	_, ok := e[r]
	return ok
}

func (e UnsupportedRepoSet) includesRepoWithID(id api.RepoID) bool {
	for r := range e {
		if r.ID == id {
			return true
		}
	}
	return false
}

func (e UnsupportedRepoSet) Error() string {
	repos := []string{}
	typeSet := map[string]struct{}{}
	for repo := range e {
		repos = append(repos, string(repo.Name))
		typeSet[repo.ExternalRepo.ServiceType] = struct{}{}
	}

	types := []string{}
	for t := range typeSet {
		types = append(types, t)
	}

	return fmt.Sprintf(
		"found repositories on unsupported code hosts: %s\nrepositories:\n\t%s",
		strings.Join(types, ", "),
		strings.Join(repos, "\n\t"),
	)
}

func (e UnsupportedRepoSet) Append(repo *types.Repo) {
	e[repo] = struct{}{}
}

func (e UnsupportedRepoSet) HasUnsupported() bool {
	return len(e) > 0
}

// IgnoredRepoSet provides a set to manage repositories that are on
// unsupported code hosts. This type implements error to allow it to be
// returned directly as an error value if needed.
type IgnoredRepoSet map[*types.Repo]struct{}

func (e IgnoredRepoSet) Includes(r *types.Repo) bool {
	_, ok := e[r]
	return ok
}

func (e IgnoredRepoSet) includesRepoWithID(id api.RepoID) bool {
	for r := range e {
		if r.ID == id {
			return true
		}
	}
	return false
}

func (e IgnoredRepoSet) Error() string {
	repos := []string{}
	for repo := range e {
		repos = append(repos, string(repo.Name))
	}

	return fmt.Sprintf(
		"found repositories containing .batchignore files:\n\t%s",
		strings.Join(repos, "\n\t"),
	)
}

func (e IgnoredRepoSet) Append(repo *types.Repo) {
	e[repo] = struct{}{}
}

func (e IgnoredRepoSet) HasIgnored() bool {
	return len(e) > 0
}

// FindDirectoriesInRepos returns a map of repositories and the locations of
// files matching the given file name in the repository.
// The locations are paths relative to the root of the directory.
// No "/" at the beginning.
// A dot (".") represents the root directory.
func (wr *workspaceResolver) FindDirectoriesInRepos(ctx context.Context, fileName string, repos ...*RepoRevision) (map[repoRevKey][]string, error) {
	findForRepoRev := func(repoRev *RepoRevision) ([]string, error) {
		query := fmt.Sprintf(`file:(^|/)%s$ repo:^%s$@%s type:path count:99999`, regexp.QuoteMeta(fileName), regexp.QuoteMeta(string(repoRev.Repo.Name)), repoRev.Commit)

		results := []string{}
		err := wr.runSearch(ctx, query, func(matches []streamhttp.EventMatch) {
			for _, match := range matches {
				switch m := match.(type) {
				case (*streamhttp.EventPathMatch):
					results = append(results, m.Path)
				}
			}
		})
		if err != nil {
			return nil, err
		}

		return results, nil
	}

	// Limit concurrency.
	sem := make(chan struct{}, 10)
	for i := 0; i < 10; i++ {
		sem <- struct{}{}
	}

	var (
		errs    error
		mu      sync.Mutex
		results = make(map[repoRevKey][]string)
	)
	for _, repoRev := range repos {
		<-sem
		go func(repoRev *RepoRevision) {
			defer func() {
				sem <- struct{}{}
			}()

			result, err := findForRepoRev(repoRev)
			if err != nil {
				errs = multierror.Append(errs, err)
				return
			}

			mu.Lock()
			results[repoRev.Key()] = result
			mu.Unlock()
		}(repoRev)
	}

	// Wait for all to finish.
	for i := 0; i < 10; i++ {
		<-sem
	}

	return results, errs
}

type directoryFinder interface {
	FindDirectoriesInRepos(ctx context.Context, fileName string, repos ...*RepoRevision) (map[repoRevKey][]string, error)
}

// findWorkspaces matches the given repos to the workspace configs and
// searches, via the Sourcegraph instance, the locations of the workspaces in
// each repository.
// The repositories that were matched by a workspace config and all repos that didn't
// match a config are returned as workspaces.
func findWorkspaces(
	ctx context.Context,
	spec *batcheslib.BatchSpec,
	finder directoryFinder,
	repoRevs []*RepoRevision,
) ([]*RepoWorkspace, error) {
	// Pre-compile all globs.
	workspaceMatchers := make(map[batcheslib.WorkspaceConfiguration]glob.Glob)
	var errs *multierror.Error
	for _, conf := range spec.Workspaces {
		g, err := glob.Compile(conf.In)
		if err != nil {
			errs = multierror.Append(errs, batcheslib.NewValidationError(errors.Errorf("failed to compile glob %q: %v", conf.In, err)))
		}
		workspaceMatchers[conf] = g
	}
	if err := errs.ErrorOrNil(); err != nil {
		return nil, err
	}

	root := []*RepoRevision{}

	// Maps workspace config indexes to repositories matching them.
	matched := map[int][]*RepoRevision{}

	for _, repoRev := range repoRevs {
		found := false

		// Try to find a workspace configuration matching this repo.
		for idx, conf := range spec.Workspaces {
			if !workspaceMatchers[conf].Match(string(repoRev.Repo.Name)) {
				continue
			}

			// Don't allow duplicate matches.
			if found {
				return nil, batcheslib.NewValidationError(errors.Errorf("repository %s matches multiple workspaces.in globs in the batch spec. glob: %q", repoRev.Repo.Name, conf.In))
			}

			matched[idx] = append(matched[idx], repoRev)
			found = true
		}

		if !found {
			root = append(root, repoRev)
		}
	}

	type repoWorkspaces struct {
		*RepoRevision
		Paths              []string
		OnlyFetchWorkspace bool
	}
	workspacesByRepoRev := map[repoRevKey]repoWorkspaces{}
	for idx, repoRevs := range matched {
		conf := spec.Workspaces[idx]
		repoRevDirs, err := finder.FindDirectoriesInRepos(ctx, conf.RootAtLocationOf, repoRevs...)
		if err != nil {
			return nil, err
		}

		repoRevsByKey := map[repoRevKey]*RepoRevision{}
		for _, repoRev := range repoRevs {
			repoRevsByKey[repoRev.Key()] = repoRev
		}

		for repoRevKey, dirs := range repoRevDirs {
			// Don't add repos that don't have any matched workspaces.
			if len(dirs) == 0 {
				continue
			}
			workspacesByRepoRev[repoRevKey] = repoWorkspaces{
				RepoRevision:       repoRevsByKey[repoRevKey],
				Paths:              dirs,
				OnlyFetchWorkspace: conf.OnlyFetchWorkspace,
			}
		}
	}

	// And add the root for repos.
	for _, repoRev := range root {
		conf, ok := workspacesByRepoRev[repoRev.Key()]
		if !ok {
			workspacesByRepoRev[repoRev.Key()] = repoWorkspaces{
				RepoRevision: repoRev,
				// Root.
				Paths:              []string{""},
				OnlyFetchWorkspace: false,
			}
			continue
		}
		conf.Paths = append(conf.Paths, "")
	}

	workspaces := make([]*RepoWorkspace, 0, len(workspacesByRepoRev))
	for _, workspace := range workspacesByRepoRev {
		steps, err := stepsForRepoRevision(spec, workspace.RepoRevision)
		if err != nil {
			return nil, err
		}

		// If the workspace doesn't have any steps we don't need to include it.
		if len(steps) == 0 {
			continue
		}

		for _, path := range workspace.Paths {
			fetchWorkspace := workspace.OnlyFetchWorkspace
			if path == "" {
				fetchWorkspace = false
			}

			workspaces = append(workspaces, &RepoWorkspace{
				RepoRevision:       workspace.RepoRevision,
				Path:               path,
				Steps:              steps,
				OnlyFetchWorkspace: fetchWorkspace,
			})
		}
	}

	// Stable sorting.
	sort.Slice(workspaces, func(i, j int) bool {
		if workspaces[i].Repo.Name == workspaces[j].Repo.Name {
			return workspaces[i].Path < workspaces[j].Path
		}
		return workspaces[i].Repo.Name < workspaces[j].Repo.Name
	})

	return workspaces, nil
}

// stepsForRepoRevision calculates the steps required to run on the given repo.
func stepsForRepoRevision(spec *batcheslib.BatchSpec, repoRev *RepoRevision) ([]batcheslib.Step, error) {
	taskSteps := []batcheslib.Step{}

	for _, step := range spec.Steps {
		// If no if condition is given, just go ahead and add the step to the list.
		if step.IfCondition() == "" {
			taskSteps = append(taskSteps, step)
			continue
		}

		batchChange := template.BatchChangeAttributes{
			Name:        spec.Name,
			Description: spec.Description,
		}
		stepCtx := &template.StepContext{
			Repository: template.TemplatingRepository{
				ID:   string(graphqlbackend.MarshalRepositoryID(repoRev.Repo.ID)),
				Name: string(repoRev.Repo.Name),
				// TODO: It's not always the default branch, this also seems to
				// be wrong in src-cli.
				DefaultBranch: template.TemplatingBranch{
					Name:      repoRev.Branch,
					TargetOID: string(repoRev.Commit),
				},
				// TODO: Reimplement.
				FileMatches: make(map[string]bool),
			},
			BatchChange: batchChange,
		}
		static, boolVal, err := template.IsStaticBool(step.IfCondition(), stepCtx)
		if err != nil {
			return nil, err
		}

		// If we could evaluate the condition statically and the resulting
		// boolean is false, we don't add that step.
		if !static {
			taskSteps = append(taskSteps, step)
		} else if boolVal {
			taskSteps = append(taskSteps, step)
		}
	}

	return taskSteps, nil
}

type repoRevKey struct {
	RepoID int32
	Branch string
	Commit string
}

func (r *RepoRevision) Key() repoRevKey {
	return repoRevKey{
		RepoID: int32(r.Repo.ID),
		Branch: r.Branch,
		Commit: string(r.Commit),
	}
}

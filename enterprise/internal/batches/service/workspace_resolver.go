package service

import (
	"context"
	"fmt"
	"os"
	"path"
	"sort"
	"sync"

	"github.com/gobwas/glob"
	"github.com/grafana/regexp"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/api/internalapi"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	streamapi "github.com/sourcegraph/sourcegraph/internal/search/streaming/api"
	streamhttp "github.com/sourcegraph/sourcegraph/internal/search/streaming/http"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
	onlib "github.com/sourcegraph/sourcegraph/lib/batches/on"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// RepoRevision describes a repository on a branch at a fixed revision.
type RepoRevision struct {
	Repo        *types.Repo
	Branch      string
	Commit      api.CommitID
	FileMatches []string
}

func (r *RepoRevision) HasBranch() bool {
	return r.Branch != ""
}

type RepoWorkspace struct {
	*RepoRevision
	Path string

	OnlyFetchWorkspace bool

	Ignored     bool
	Unsupported bool
}

type WorkspaceResolver interface {
	ResolveWorkspacesForBatchSpec(
		ctx context.Context,
		batchSpec *batcheslib.BatchSpec,
	) (
		workspaces []*RepoWorkspace,
		err error,
	)
}

type WorkspaceResolverBuilder func(tx *store.Store) WorkspaceResolver

func NewWorkspaceResolver(s *store.Store, gitserverClient gitserver.Client) WorkspaceResolver {
	return &workspaceResolver{
		store:               s,
		gitserverClient:     gitserverClient,
		frontendInternalURL: internalapi.Client.URL + "/.internal",
	}
}

type workspaceResolver struct {
	store               *store.Store
	gitserverClient     gitserver.Client
	frontendInternalURL string
}

func (wr *workspaceResolver) ResolveWorkspacesForBatchSpec(ctx context.Context, batchSpec *batcheslib.BatchSpec) (workspaces []*RepoWorkspace, err error) {
	tr, ctx := trace.New(ctx, "workspaceResolver.ResolveWorkspacesForBatchSpec", "")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	// First, find all repositories that match the batch spec on definitions.
	// This list is filtered by permissions using database.Repos.List.
	repos, err := wr.determineRepositories(ctx, batchSpec)
	if err != nil {
		return nil, err
	}

	// Next, find the repos that are ignoredRepos through a .batchignore file.
	ignoredRepos, err := findIgnoredRepositories(ctx, database.NewDBWith(wr.store), repos)
	if err != nil {
		return nil, err
	}

	// Now build the workspaces for the list of repos
	workspaces, err = findWorkspaces(ctx, batchSpec, wr, repos)
	if err != nil {
		return nil, err
	}

	// Finally, tag the workspaces if they're (a) on an unsupported code host
	// or (b) ignored.
	for _, ws := range workspaces {
		if !btypes.IsKindSupported(extsvc.TypeToKind(ws.Repo.ExternalRepo.ServiceType)) {
			ws.Unsupported = true
		}

		if _, ok := ignoredRepos[ws.Repo]; ok {
			ws.Ignored = true
		}
	}

	// Sort the workspaces so that the list of workspaces is kinda stable when
	// using `replaceBatchSpecInput`.
	sort.Slice(workspaces, func(i, j int) bool {
		if workspaces[i].Repo.Name != workspaces[j].Repo.Name {
			return workspaces[i].Repo.Name < workspaces[j].Repo.Name
		}
		if workspaces[i].Path != workspaces[j].Path {
			return workspaces[i].Path < workspaces[j].Path
		}
		return workspaces[i].Branch < workspaces[j].Branch
	})

	return workspaces, nil
}

func (wr *workspaceResolver) determineRepositories(ctx context.Context, batchSpec *batcheslib.BatchSpec) ([]*RepoRevision, error) {
	agg := onlib.NewRepoRevisionAggregator()

	var errs error
	// TODO: this could be trivially parallelised in the future.
	for _, on := range batchSpec.On {
		revs, ruleType, err := wr.resolveRepositoriesOn(ctx, &on)
		if err != nil {
			errs = errors.Append(errs, errors.Wrapf(err, "resolving %q", on.String()))
			continue
		}

		result := agg.NewRuleRevisions(ruleType)
		for _, rev := range revs {
			// Skip repos where no branch exists.
			if !rev.HasBranch() {
				continue
			}

			result.AddRepoRevision(rev.Repo.ID, rev)
		}
	}

	repoRevs := []*RepoRevision{}
	for _, rev := range agg.Revisions() {
		repoRevs = append(repoRevs, rev.(*RepoRevision))
	}
	return repoRevs, errs
}

// TODO: Use a gitserver batch method like https://sourcegraph.com/-/editor?remote_url=github.com%2Fsourcegraph%2Fsourcegraph&branch=es%2Frace-condition-search-result-paths&file=internal%2Fgitserver%2Fclient.go&editor=VSCode&version=2.2.3&start_row=666&start_col=12&end_row=666&end_col=12&utm_campaign=vscode-extension&utm_medium=direct_traffic&utm_source=vscode-extension&utm_content=vsce-commands.
func findIgnoredRepositories(ctx context.Context, db database.DB, repos []*RepoRevision) (map[*types.Repo]struct{}, error) {
	type result struct {
		repo           *RepoRevision
		hasBatchIgnore bool
		err            error
	}

	var (
		ignored = make(map[*types.Repo]struct{})

		input   = make(chan *RepoRevision, len(repos))
		results = make(chan result, len(repos))

		wg sync.WaitGroup
	)

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(in chan *RepoRevision, out chan result) {
			defer wg.Done()
			for repo := range in {
				hasBatchIgnore, err := hasBatchIgnoreFile(ctx, db, repo)
				out <- result{repo, hasBatchIgnore, err}
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

	var errs error
	for result := range results {
		if result.err != nil {
			errs = errors.Append(errs, result.err)
			continue
		}

		if result.hasBatchIgnore {
			ignored[result.repo.Repo] = struct{}{}
		}
	}

	return ignored, errs
}

var ErrMalformedOnQueryOrRepository = batcheslib.NewValidationError(errors.New("malformed 'on' field; missing either a repository name or a query"))

// resolveRepositoriesOn resolves a single on: entry in a batch spec.
func (wr *workspaceResolver) resolveRepositoriesOn(ctx context.Context, on *batcheslib.OnQueryOrRepository) (_ []*RepoRevision, _ onlib.RepositoryRuleType, err error) {
	tr, ctx := trace.New(ctx, "workspaceResolver.resolveRepositoriesOn", "")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	if on.RepositoriesMatchingQuery != "" {
		revs, err := wr.resolveRepositoriesMatchingQuery(ctx, on.RepositoriesMatchingQuery)
		return revs, onlib.RepositoryRuleTypeQuery, err
	}

	branches, err := on.GetBranches()
	if err != nil {
		return nil, onlib.RepositoryRuleTypeExplicit, err
	}

	if on.Repository != "" && len(branches) > 0 {
		revs := make([]*RepoRevision, len(branches))
		for i, branch := range branches {
			repo, err := wr.resolveRepositoryNameAndBranch(ctx, on.Repository, branch)
			if err != nil {
				return nil, onlib.RepositoryRuleTypeExplicit, err
			}

			revs[i] = repo
		}
		return revs, onlib.RepositoryRuleTypeExplicit, nil
	}

	if on.Repository != "" {
		repo, err := wr.resolveRepositoryName(ctx, on.Repository)
		if err != nil {
			return nil, onlib.RepositoryRuleTypeExplicit, err
		}
		return []*RepoRevision{repo}, onlib.RepositoryRuleTypeExplicit, nil
	}

	// This shouldn't happen on any batch spec that has passed validation, but,
	// alas, software.
	return nil, onlib.RepositoryRuleTypeExplicit, ErrMalformedOnQueryOrRepository
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

	return repoToRepoRevisionWithDefaultBranch(
		ctx,
		database.NewDBWith(wr.store),
		wr.gitserverClient,
		repo,
		// Directly resolved repos don't have any file matches.
		[]string{},
	)
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

	commit, err := wr.gitserverClient.ResolveRevision(ctx, repo.Name, branch, gitserver.ResolveRevisionOptions{
		NoEnsureRevision: true,
	})
	if err != nil && errors.HasType(err, &gitdomain.RevisionNotFoundError{}) {
		return nil, errors.Newf("no branch matching %q found for repository %s", branch, name)
	}

	return &RepoRevision{
		Repo:   repo,
		Branch: branch,
		Commit: commit,
		// Directly resolved repos don't have any file matches.
		FileMatches: []string{},
	}, nil
}

func (wr *workspaceResolver) resolveRepositoriesMatchingQuery(ctx context.Context, query string) (_ []*RepoRevision, err error) {
	tr, ctx := trace.New(ctx, "workspaceResolver.resolveRepositorySearch", "")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	query = setDefaultQueryCount(query)

	repoIDs := make(map[api.RepoID]struct{})
	repoRevisions := make(map[api.RepoID]string)
	repoTargets := make(map[api.RepoID]string)
	repoFileMatches := make(map[api.RepoID]map[string]struct{})
	addRepoFilePath := func(repoID api.RepoID, branch, commit, path string) {
		// We check if we have seen a revision for the repo before. If so, it must be the same.
		// Otherwise we deem this user error. It's currently not supported to target two revisions
		// of the same repo.
		if knownRev, ok := repoRevisions[repoID]; !ok {
			repoRevisions[repoID] = commit
		} else if ok && knownRev != commit {
			log15.Error("Got repo results for multiple commits, this is unexpected", "repo", repoID, "rev1", knownRev, "rev2", commit)
			// And live on and overwrite the known revision.
			// TODO: Can this happen when the repo has multiple searchable branches.
			repoRevisions[repoID] = commit
		}

		if branch != "" {
			if knownTarget, ok := repoTargets[repoID]; !ok {
				repoTargets[repoID] = branch
			} else if ok && knownTarget != branch {
				log15.Error("Got repo results for multiple branches, this is unexpected", "repo", repoID, "branch1", knownTarget, "branch2", branch)
				// And live on and overwrite the known revision.
				// TODO: Can this happen when the repo has multiple searchable branches.
				repoTargets[repoID] = branch
			}
		}

		repoMap, ok := repoFileMatches[repoID]
		if !ok {
			repoMap = make(map[string]struct{})
			repoFileMatches[repoID] = repoMap
		}
		if _, ok := repoMap[path]; !ok {
			repoMap[path] = struct{}{}
		}
	}
	visitRepo := func(repoID int32) {
		id := api.RepoID(repoID)
		if _, ok := repoIDs[id]; !ok {
			repoIDs[id] = struct{}{}
		}
	}
	if err := wr.runSearch(ctx, query, func(matches []streamhttp.EventMatch) {
		for _, match := range matches {
			switch m := match.(type) {
			case *streamhttp.EventRepoMatch:
				var branch string
				if len(m.Branches) > 0 {
					if len(m.Branches) > 1 {
						// This is an error, but it can currently not happen.
					}
					branch = m.Branches[0]
				}
				visitRepo(m.RepositoryID)
				// TODO: this currently overwrites.
				repoRevisions[api.RepoID(m.RepositoryID)] = branch
			case *streamhttp.EventContentMatch:
				var branch string
				if len(m.Branches) > 0 {
					if len(m.Branches) > 1 {
						// This is an error, but it can currently not happen.
					}
					branch = m.Branches[0]
				}
				visitRepo(m.RepositoryID)
				addRepoFilePath(api.RepoID(m.RepositoryID), branch, m.Commit, m.Path)
			case *streamhttp.EventPathMatch:
				var branch string
				if len(m.Branches) > 0 {
					if len(m.Branches) > 1 {
						// This is an error, but it can currently not happen.
					}
					branch = m.Branches[0]
				}
				visitRepo(m.RepositoryID)
				addRepoFilePath(api.RepoID(m.RepositoryID), branch, m.Commit, m.Path)
			case *streamhttp.EventSymbolMatch:
				var branch string
				if len(m.Branches) > 0 {
					if len(m.Branches) > 1 {
						// This is an error, but it can currently not happen.
					}
					branch = m.Branches[0]
				}
				visitRepo(m.RepositoryID)
				addRepoFilePath(api.RepoID(m.RepositoryID), branch, m.Commit, m.Path)
			}
		}
	}); err != nil {
		return nil, err
	}

	// If no repos matched the search query, we can early return.
	if len(repoIDs) == 0 {
		return []*RepoRevision{}, nil
	}

	ids := make([]api.RepoID, 0, len(repoIDs))
	for id := range repoIDs {
		ids = append(ids, id)
	}

	// 🚨 SECURITY: We use database.Repos.List to check whether the user has access to
	// the repositories or not. We also impersonate on the internal search request to
	// properly respect these permissions, so this should not filter anything out,
	// but better be safe than sorry!
	accessibleRepos, err := wr.store.Repos().List(ctx, database.ReposListOptions{IDs: ids})
	if err != nil {
		return nil, err
	}

	revs := make([]*RepoRevision, 0, len(accessibleRepos))
	for _, repo := range accessibleRepos {
		fileMatches := make([]string, 0, len(repoFileMatches[repo.ID]))
		for path := range repoFileMatches[repo.ID] {
			fileMatches = append(fileMatches, path)
		}
		sort.Strings(fileMatches)

		var rev *RepoRevision
		// TODO: Check here if the branch has been overwritten. If, use that instead of the default branch.
		// Also, validate that the SHA is part of the default branch.
		if target, ok := repoTargets[repo.ID]; ok {
			rev, err = repoToRepoRevisionWithBranch(ctx, wr.store.DatabaseDB(), wr.gitserverClient, repo, target, fileMatches)
			if err != nil {
				// There is an edge-case where a repo might be returned by a search query that does not exist in gitserver yet.
				if errcode.IsNotFound(err) {
					continue
				}
				return nil, err
			}
		} else {
			rev, err = repoToRepoRevisionWithDefaultBranch(ctx, wr.store.DatabaseDB(), wr.gitserverClient, repo, fileMatches)
			if err != nil {
				// There is an edge-case where a repo might be returned by a search query that does not exist in gitserver yet.
				if errcode.IsNotFound(err) {
					continue
				}
				return nil, err
			}
		}

		// If we have searched a different commit than the default branch commit, we overwrite it here.
		// There can be cases where indexed search lags behind latest main and then the search result paths
		// can be incorrect for the latest head of the repo.
		// TODO: This is not ideal. It does not verify the commit is actually part of the default branch.
		// What if the user searched a different branch?
		if commit, ok := repoRevisions[repo.ID]; ok {
			rev.Commit = api.CommitID(commit)
		}
		revs = append(revs, rev)
	}

	return revs, nil
}

const internalSearchClientUserAgent = "Batch Changes repository resolver"

// TODO: Run the search from within worker, instead of hitting the frontend.
// This will help with load/memory issues for user-facing requests.
func (wr *workspaceResolver) runSearch(ctx context.Context, query string, onMatches func(matches []streamhttp.EventMatch)) (err error) {
	// We impersonate as the user who initiated this search. This is to properly
	// scope repository permissions while running the search.
	a := actor.FromContext(ctx)
	if !a.IsAuthenticated() {
		return errors.New("no user set in workspaceResolver.runSearch")
	}

	req, err := streamhttp.NewRequest(wr.frontendInternalURL, query)
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)

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
	decErr := dec.ReadAll(resp.Body)
	if decErr != nil {
		return decErr
	}
	return err
}

func repoToRepoRevisionWithBranch(ctx context.Context, db database.DB, gitserverClient gitserver.Client, repo *types.Repo, branch string, fileMatches []string) (_ *RepoRevision, err error) {
	tr, ctx := trace.New(ctx, "repoToRepoRevision", "")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	// TODO: Verify this is the right method for what we want.
	// (We want to get the latest commit of the branch).
	commit, err := gitserverClient.ResolveRevision(ctx, repo.Name, branch, gitserver.ResolveRevisionOptions{NoEnsureRevision: true})
	if err != nil {
		return nil, err
	}

	repoRev := &RepoRevision{
		Repo:        repo,
		Branch:      branch,
		Commit:      commit,
		FileMatches: fileMatches,
	}
	return repoRev, nil
}

func repoToRepoRevisionWithDefaultBranch(ctx context.Context, db database.DB, gitserverClient gitserver.Client, repo *types.Repo, fileMatches []string) (_ *RepoRevision, err error) {
	tr, ctx := trace.New(ctx, "repoToRepoRevision", "")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	branch, commit, err := gitserverClient.GetDefaultBranch(ctx, repo.Name)
	if err != nil {
		return nil, err
	}

	repoRev := &RepoRevision{
		Repo:        repo,
		Branch:      branch,
		Commit:      commit,
		FileMatches: fileMatches,
	}
	return repoRev, nil
}

// hasBatchIgnoreFile returns true, if the given RepoRevision contains a .batchignore
// file at the root of the repository.
// TODO: Can we change this algorithm slightly to just check if there is a .batchignore file on latest
// default branch? If so, we could have a background worker determine changes to repos and we wouldn't
// need to spawn thousands of git stat calls for a larger workspace resolution.
// We can then look at the last changed date for a repo and the background worker remembers up to which
// time it has evaluated all repos. We then store in a DB table which repos are ignored.
// batch_change_ignored_repos (repo_id).
func hasBatchIgnoreFile(ctx context.Context, db database.DB, r *RepoRevision) (_ bool, err error) {
	traceTitle := fmt.Sprintf("RepoID: %q", r.Repo.ID)
	tr, ctx := trace.New(ctx, "hasBatchIgnoreFile", traceTitle)
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	const path = ".batchignore"
	stat, err := git.Stat(ctx, db, authz.DefaultSubRepoPermsChecker, r.Repo.Name, r.Commit, path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	// TODO: Should we really error here?
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

// FindDirectoriesInRepos returns a map of repositories and the locations of
// files matching the given file name in the repository.
// The locations are paths relative to the root of the directory.
// No "/" at the beginning.
// A dot (".") represents the root directory.
// TODO: Can we use gitserver here instead of search? If so, we should use a bulk endpoint.
func (wr *workspaceResolver) FindDirectoriesInRepos(ctx context.Context, fileName string, repos ...*RepoRevision) (map[repoRevKey][]string, error) {
	findForRepoRev := func(repoRev *RepoRevision) ([]string, error) {
		query := fmt.Sprintf(`file:(^|/)%s$ repo:^%s$@%s type:path count:99999`, regexp.QuoteMeta(fileName), regexp.QuoteMeta(string(repoRev.Repo.Name)), repoRev.Commit)

		results := []string{}
		err := wr.runSearch(ctx, query, func(matches []streamhttp.EventMatch) {
			for _, match := range matches {
				switch m := match.(type) {
				case *streamhttp.EventPathMatch:
					// We use path.Dir and not filepath.Dir here, because while
					// src-cli might be executed on Windows, we need the paths to
					// be Unix paths, since they will be used inside Docker
					// containers.
					dir := path.Dir(m.Path)

					// "." means the path is root, but in the executor we use "" to signify root.
					if dir == "." {
						dir = ""
					}

					results = append(results, dir)
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
				errs = errors.Append(errs, err)
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
	var errs error
	for _, conf := range spec.Workspaces {
		in := conf.In
		// Empty `in` should fall back to matching all, instead of nothing.
		if in == "" {
			in = "*"
		}
		g, err := glob.Compile(in)
		if err != nil {
			errs = errors.Append(errs, batcheslib.NewValidationError(errors.Errorf("failed to compile glob %q: %v", in, err)))
		}
		workspaceMatchers[conf] = g
	}
	if errs != nil {
		return nil, errs
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

			// TODO: Evaluate all matches first and return a multi error. Simpler debugging for users :)
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
		for _, path := range workspace.Paths {
			fetchWorkspace := workspace.OnlyFetchWorkspace
			if path == "" {
				fetchWorkspace = false
			}

			workspaces = append(workspaces, &RepoWorkspace{
				RepoRevision:       workspace.RepoRevision,
				Path:               path,
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

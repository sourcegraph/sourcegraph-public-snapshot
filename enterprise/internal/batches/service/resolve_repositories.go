package service

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/cockroachdb/errors"
	"github.com/hashicorp/go-multierror"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
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
)

type RepoRevision struct {
	Repo   *types.Repo
	Branch string
	Commit api.CommitID
}

func (r *RepoRevision) HasBranch() bool {
	return r.Branch != ""
}

type ResolveRepositoriesForBatchSpecOpts struct {
	AllowIgnored     bool
	AllowUnsupported bool
}

func (s *Service) ResolveRepositoriesForBatchSpec(ctx context.Context, batchSpec *batcheslib.BatchSpec, opts ResolveRepositoriesForBatchSpecOpts) (_ []*RepoRevision, err error) {
	traceTitle := fmt.Sprintf("len(On): %d", len(batchSpec.On))
	tr, ctx := trace.New(ctx, "service.ResolveRepositoriesForBatchSpec", traceTitle)
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	wr := &workspaceResolver{store: s.store, frontendInternalURL: api.InternalClient.URL + "/.internal"}
	return wr.ResolveRepositoriesForBatchSpec(ctx, batchSpec, opts)
}

type workspaceResolver struct {
	store               *store.Store
	frontendInternalURL string
}

func (wr *workspaceResolver) ResolveRepositoriesForBatchSpec(ctx context.Context, batchSpec *batcheslib.BatchSpec, opts ResolveRepositoriesForBatchSpecOpts) (_ []*RepoRevision, err error) {
	seen := map[api.RepoID]*RepoRevision{}
	unsupported := UnsupportedRepoSet{}

	// TODO: this could be trivially parallelised in the future.
	for _, on := range batchSpec.On {
		repos, err := wr.resolveRepositoriesOn(ctx, &on)
		if err != nil {
			return nil, errors.Wrapf(err, "resolving %q", on.String())
		}

		for _, repo := range repos {
			// Skip repos where no branch exists.
			if !repo.HasBranch() {
				continue
			}

			if other, ok := seen[repo.Repo.ID]; !ok {
				seen[repo.Repo.ID] = repo

				switch st := repo.Repo.ExternalRepo.ServiceType; st {
				case extsvc.TypeGitHub, extsvc.TypeGitLab, extsvc.TypeBitbucketServer:
				default:
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

	final, ignored, err := filterIgnoredRepositories(ctx, seen, opts.AllowIgnored, unsupported)
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

var ErrMalformedOnQueryOrRepository = errors.New("malformed 'on' field; missing either a repository name or a query")

func (wr *workspaceResolver) resolveRepositoriesOn(ctx context.Context, on *batcheslib.OnQueryOrRepository) (_ []*RepoRevision, err error) {
	tr, ctx := trace.New(ctx, "workspaceResolver.resolveRepositoriesOn", "")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	if on.RepositoriesMatchingQuery != "" {
		return wr.resolveRepositoriesMatchingQuery(ctx, on.RepositoriesMatchingQuery)
	} else if on.Repository != "" && on.Branch != "" {
		repo, err := wr.resolveRepositoryNameAndBranch(ctx, on.Repository, on.Branch)
		if err != nil {
			return nil, err
		}
		return []*RepoRevision{repo}, nil
	} else if on.Repository != "" {
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

	return repoToRepoRevision(ctx, repo)
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
		rev, err := repoToRepoRevision(ctx, repo)
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
		OnMatches: func(matches []streamhttp.EventMatch) {
			onMatches(matches)
		},
		OnError: func(ee *streamhttp.EventError) {
			err = errors.New(ee.Message)
		},
		OnProgress: func(p *streamapi.Progress) {
			// TODO: Evaluate skipped for values we care about.
		},
	}
	return dec.ReadAll(resp.Body)
}

func repoToRepoRevision(ctx context.Context, repo *types.Repo) (_ *RepoRevision, err error) {
	tr, ctx := trace.New(ctx, "repoToRepoRevision", "")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	repoRev := &RepoRevision{
		Repo: repo,
	}

	repoRev.Branch, repoRev.Commit, err = git.GetDefaultBranch(ctx, repo.Name)
	return repoRev, err
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

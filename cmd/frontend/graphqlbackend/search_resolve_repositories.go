package graphqlbackend

import (
	"context"
	"fmt"
	"regexp"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
	"github.com/sourcegraph/sourcegraph/schema"
)

type resolveRepoOp struct {
	repoFilters        []string
	minusRepoFilters   []string
	repoGroupFilters   []string
	versionContextName string
	noForks            bool
	onlyForks          bool
	noArchived         bool
	onlyArchived       bool
	commitAfter        string
	onlyPrivate        bool
	onlyPublic         bool
}

type repositoryResolver struct {
	resolveRepoOp

	tracer interface {
		LazyPrintf(format string, a ...interface{})
	}

	// these struct members represent the input of the repositoryResolver
	// and must be filled prior to calling the resolveRepositories method
	maxRepoListSize int

	// these struct members are intermediary variables used
	// by the various methods called by the resolveRepositories method
	includePatternRevs         []patternRevspec
	includePatterns            []string
	versionContextRepositories []string
	missingRepoRevisions       []*search.RepositoryRevisions
	repoRevisions              []*search.RepositoryRevisions
}

func resolveRepositories(ctx context.Context, op resolveRepoOp) (repoRevisions, missingRepoRevisions []*search.RepositoryRevisions, overLimit bool, err error) {
	tr, ctx := trace.New(ctx, "resolveRepositories", fmt.Sprintf("%+v", op))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	r := repositoryResolver{
		resolveRepoOp:   op,
		tracer:          tr,
		maxRepoListSize: maxReposToSearch(),
	}

	return r.resolveRepositories(ctx)
}

func (r *repositoryResolver) resolveRepositories(ctx context.Context) ([]*search.RepositoryRevisions, []*search.RepositoryRevisions, bool, error) {
	// If any repo groups are specified, take the intersection of the repo
	// groups and the set of repos specified with repo:. (If none are specified
	// with repo:, then include all from the group.)
	if groupNames := r.repoGroupFilters; len(groupNames) > 0 {
		err := r.mergeRepoWithRepoGroups(ctx, groupNames)
		if err != nil {
			return nil, nil, false, err
		}
	}

	var err error

	// note that this mutates the strings in includePatterns, stripping their
	// revision specs, if they had any.
	r.includePatterns, r.includePatternRevs, err = findPatternRevs(r.repoFilters)
	if err != nil {
		return nil, nil, false, err
	}

	// If a version context is specified, gather the list of repository names
	// to limit the results to these repositories.
	versionContext, err := r.getVersionContext()
	if err != nil {
		return nil, nil, false, err
	}

	repos, err := r.getRepos(ctx)
	if err != nil {
		return nil, nil, false, err
	}

	overLimit := len(repos) >= r.maxRepoListSize

	r.repoRevisions = make([]*search.RepositoryRevisions, 0, len(repos))
	r.tracer.LazyPrintf("Associate/validate revs - start")

	if versionContext != nil {
		r.validateAndAssociateWithVersionContext(ctx, repos, versionContext)
	} else {
		r.validateAndAssociate(ctx, repos)
	}

	r.tracer.LazyPrintf("Associate/validate revs - done")

	if r.commitAfter != "" {
		r.repoRevisions, err = filterRepoHasCommitAfter(ctx, r.repoRevisions, r.commitAfter)
	}

	return r.repoRevisions, r.missingRepoRevisions, overLimit, err
}

// If any repo groups are specified, take the intersection of the repo
// groups and the set of repos specified with repo:. (If none are specified
// with repo:, then include all from the group.)
func (r *repositoryResolver) mergeRepoWithRepoGroups(ctx context.Context, groupNames []string) error {
	groups, err := resolveRepoGroups(ctx)
	if err != nil {
		return err
	}
	var patterns []string
	for _, groupName := range groupNames {
		for _, repo := range groups[groupName] {
			patterns = append(patterns, "^"+regexp.QuoteMeta(string(repo.Name))+"$")
		}
	}

	r.repoFilters = append(r.repoFilters, unionRegExps(patterns))

	// Ensure we don't omit any repos explicitly included via a repo group.
	if len(patterns) > r.maxRepoListSize {
		r.maxRepoListSize = len(patterns)
	}

	return nil
}

// If a version context is specified, gather the list of repository names
// to limit the results to these repositories.
// If no version context was specified or if the user query contains a reference, a nil context is returned.
func (r *repositoryResolver) getVersionContext() (*schema.VersionContext, error) {
	// If a ref is specified we ignore the version context.
	if len(r.includePatternRevs) > 0 {
		return nil, nil
	}

	// if no version context is specified, we return nothing
	if r.versionContextName == "" {
		return nil, nil
	}

	var versionContext *schema.VersionContext
	for _, vc := range conf.Get().ExperimentalFeatures.VersionContexts {
		if vc.Name == r.versionContextName {
			versionContext = vc
			break
		}
	}
	if versionContext == nil {
		return nil, errors.New("version context not found")
	}

	for _, revision := range versionContext.Revisions {
		r.versionContextRepositories = append(r.versionContextRepositories, revision.Repo)
	}

	return versionContext, nil
}

// get repositories from the database or from the default repositories.
func (r *repositoryResolver) getRepos(ctx context.Context) ([]*types.Repo, error) {
	var err error
	var repos []*types.Repo

	if envvar.SourcegraphDotComMode() && len(r.includePatterns) == 0 {
		getIndexedRepos := func(ctx context.Context, revs []*search.RepositoryRevisions) (indexed, unindexed []*search.RepositoryRevisions, err error) {
			return zoektIndexedRepos(ctx, search.Indexed(), revs, nil)
		}
		repos, err := defaultRepositories(ctx, db.DefaultRepos.List, getIndexedRepos)
		if err != nil {
			return nil, errors.Wrap(err, "getting list of default repos")
		}
		if len(repos) > r.maxRepoListSize {
			repos = repos[:r.maxRepoListSize]
		}

		return repos, nil
	}

	r.tracer.LazyPrintf("Repos.List - start")
	repos, err = db.Repos.List(ctx, db.ReposListOptions{
		OnlyRepoIDs:     true,
		IncludePatterns: r.includePatterns,
		Names:           r.versionContextRepositories,
		ExcludePattern:  unionRegExps(r.minusRepoFilters),
		// List N+1 repos so we can see if there are repos omitted due to our repo limit.
		LimitOffset:  &db.LimitOffset{Limit: r.maxRepoListSize + 1},
		NoForks:      r.noForks,
		OnlyForks:    r.onlyForks,
		NoArchived:   r.noArchived,
		OnlyArchived: r.onlyArchived,
		NoPrivate:    r.onlyPublic,
		OnlyPrivate:  r.onlyPrivate,
	})
	r.tracer.LazyPrintf("Repos.List - done")
	if err != nil {
		return nil, err
	}

	return repos, nil
}

// checks if the repository actually has the revisions that the user specified and returns missing revisions.
func (r *repositoryResolver) findRepositoryRevisions(ctx context.Context, repo *search.RepositoryRevisions, revs []search.RevisionSpecifier) (revisionsFound []search.RevisionSpecifier) {
	// Check if the repository actually has the revisions that the user specified.
	for _, rev := range revs {
		if rev.RefGlob != "" || rev.ExcludeRefGlob != "" {
			// Do not validate ref patterns. A ref pattern matching 0 refs is not necessarily
			// invalid, so it's not clear what validation would even mean.
		} else if isDefaultBranch := rev.RevSpec == ""; !isDefaultBranch { // skip default branch resolution to save time
			// Validate the revspec.

			// Do not trigger a repo-updater lookup (e.g.,
			// backend.{GitRepo,Repos.ResolveRev}) because that would slow this operation
			// down by a lot (if we're looping over many repos). This means that it'll fail if a
			// repo is not on gitserver.
			//
			// TODO(sqs): make this NOT send gitserver this revspec in EnsureRevision, to avoid
			// searches like "repo:@foobar" (where foobar is an invalid revspec on most repos)
			// taking a long time because they all ask gitserver to try to fetch from the remote
			// repo.
			if _, err := git.ResolveRevision(ctx, repo.GitserverRepo(), nil, rev.RevSpec, &git.ResolveRevisionOptions{NoEnsureRevision: true}); gitserver.IsRevisionNotFound(err) || err == context.DeadlineExceeded {
				// The revspec does not exist, so don't include it, and report that it's missing.
				if rev.RevSpec == "" {
					// Report as HEAD not "" (empty string) to avoid user confusion.
					rev.RevSpec = "HEAD"
				}
				r.missingRepoRevisions = append(r.missingRepoRevisions, &search.RepositoryRevisions{
					Repo: repo.Repo,
					Revs: []search.RevisionSpecifier{{RevSpec: rev.RevSpec}},
				})
				continue
			}
			// If err != nil and is not one of the err values checked for above, cloning and other errors will be handled later, so just ignore an error
			// if there is one.
		}

		revisionsFound = append(revisionsFound, rev)
	}

	return
}

// validate and associate repositories with revisions
func (r *repositoryResolver) validateAndAssociate(ctx context.Context, repos []*types.Repo) {
	for _, repo := range repos {
		var repoRev search.RepositoryRevisions
		var revs []search.RevisionSpecifier
		var clashingRevs []search.RevisionSpecifier
		revs, clashingRevs = getRevsForMatchedRepo(repo.Name, r.includePatternRevs)
		repoRev.Repo = repo
		// if multiple specified revisions clash, report this usefully:
		if len(revs) == 0 && clashingRevs != nil {
			r.missingRepoRevisions = append(r.missingRepoRevisions, &search.RepositoryRevisions{
				Repo: repo,
				Revs: clashingRevs,
			})
		}

		// We do in place filtering to reduce allocations. Common path is no
		// filtering of revs.
		if len(revs) > 0 {
			repoRev.Revs = revs[:0]
		}

		// Check if the repository actually has the revisions that the user specified.
		revsFound := r.findRepositoryRevisions(ctx, &repoRev, revs)
		repoRev.Revs = append(repoRev.Revs, revsFound...)
		r.repoRevisions = append(r.repoRevisions, &repoRev)
	}
}

// validate and associate repositories and revisions within a version context
func (r *repositoryResolver) validateAndAssociateWithVersionContext(ctx context.Context, repos []*types.Repo, versionContext *schema.VersionContext) {
	for _, repo := range repos {
		var repoRev search.RepositoryRevisions
		var revs []search.RevisionSpecifier
		// versionContext will be nil if the query contains revision specifiers
		for _, vcRepoRef := range versionContext.Revisions {
			if vcRepoRef.Repo == string(repo.Name) {
				repoRev.Repo = repo
				revs = append(revs, search.RevisionSpecifier{RevSpec: vcRepoRef.Ref})
				break
			}
		}

		// We do in place filtering to reduce allocations. Common path is no
		// filtering of revs.
		if len(revs) > 0 {
			repoRev.Revs = revs[:0]
		}

		// Check if the repository actually has the revisions that the user specified.
		revsFound := r.findRepositoryRevisions(ctx, &repoRev, revs)
		repoRev.Revs = append(repoRev.Revs, revsFound...)
		r.repoRevisions = append(r.repoRevisions, &repoRev)
	}
}

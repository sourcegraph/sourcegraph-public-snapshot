package graphqlbackend

import (
	"context"
	"strconv"
	"sync"
	"time"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/envvar"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/search2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"

	"github.com/pkg/errors"
)

type diffSearchResult struct {
	diff    *diff
	preview *highlightedString
}

func (r *diffSearchResult) Diff() *diff                 { return r.diff }
func (r *diffSearchResult) Preview() *highlightedString { return r.preview }

type commitSearchResult struct {
	commit      *commitInfoResolver
	diffPreview *highlightedString
}

func (r *commitSearchResult) Commit() *commitInfoResolver     { return r.commit }
func (r *commitSearchResult) DiffPreview() *highlightedString { return r.diffPreview }

var mockSearchCommitDiffsInRepo func(ctx context.Context, repoName, rev string, info *patternInfo, combinedQuery resolvedQuery) (results []*commitSearchResult, limitHit bool, err error)

func searchCommitDiffsInRepo(ctx context.Context, repoName, rev string, info *patternInfo, combinedQuery resolvedQuery) (results []*commitSearchResult, limitHit bool, err error) {
	if mockSearchRepo != nil {
		return mockSearchCommitDiffsInRepo(ctx, repoName, rev, info, combinedQuery)
	}

	repo, err := localstore.Repos.GetByURI(ctx, repoName)
	if err != nil {
		return nil, false, err
	}
	// 🚨 SECURITY: DO NOT REMOVE THIS CHECK! ResolveRev is responsible for ensuring 🚨
	// the user has permissions to access the repository.
	if _, err := backend.Repos.ResolveRev(ctx, &sourcegraph.ReposResolveRevOp{Repo: repo.ID, Rev: rev}); err != nil {
		return nil, false, err
	}

	vcsrepo, err := localstore.RepoVCS.Open(ctx, repo.ID)
	if err != nil {
		return nil, false, err
	}

	const maxResults = 20 // TODO(sqs): arbitrary
	args := []string{
		"--unified=0",
		"--no-prefix",
		"--max-count=" + strconv.Itoa(maxResults+1),
	}
	if !combinedQuery.isCaseSensitive() {
		args = append(args, "--regexp-ignore-case")
	}

	for _, s := range combinedQuery.fieldValues[searchFieldBefore].Values() {
		args = append(args, "--until="+s)
	}
	for _, s := range combinedQuery.fieldValues[searchFieldAfter].Values() {
		args = append(args, "--since="+s)
	}

	// Helper for adding git log flags --grep, --author, and --committer, which all behave similarly.
	var hasSeenGrepLikeFields, hasSeenInvertedGrepLikeFields bool
	addGrepLikeFlags := func(args *[]string, gitLogFlag string, field search2.Field) error {
		values := combinedQuery.fieldValues[field].Values()
		minusValues := combinedQuery.fieldValues[minusField(field)].Values()

		hasSeenGrepLikeFields = hasSeenGrepLikeFields || len(values) > 0
		hasSeenInvertedGrepLikeFields = hasSeenInvertedGrepLikeFields || len(minusValues) > 0

		if hasSeenGrepLikeFields && hasSeenInvertedGrepLikeFields {
			// TODO(sqs): this is a limitation of `git log` flags, but we could overcome this
			// with post-filtering
			return errors.New("query not supported: combining message:/author:/committer: and -message/-author:/-committer: filters")
		}
		if len(values) > 0 || len(minusValues) > 0 {
			// To be consistent with how other filters work, always treat additional
			// filters as further constraining the result set, not widening it.
			*args = append(*args, "--all-match")

			if len(minusValues) > 0 {
				*args = append(*args, "--invert-grep")
			}

			// Only one of these for-loops will have any values to iterate over.
			for _, s := range values {
				*args = append(*args, gitLogFlag+"="+s)
			}
			for _, s := range minusValues {
				*args = append(*args, gitLogFlag+"="+s)
			}
		}
		return nil
	}
	if err := addGrepLikeFlags(&args, "--grep", searchFieldMessage); err != nil {
		return nil, false, err
	}
	if err := addGrepLikeFlags(&args, "--author", searchFieldAuthor); err != nil {
		return nil, false, err
	}
	if err := addGrepLikeFlags(&args, "--committer", searchFieldCommitter); err != nil {
		return nil, false, err
	}

	strv := func(s *string) string {
		if s == nil {
			return ""
		}
		return *s
	}

	// TODO(sqs): set extra strict timeout to avoid runaway resource consumption
	// during testing of this feature
	var timeout time.Duration
	if envvar.DeploymentOnPrem() {
		timeout = 5 * time.Second
	} else {
		timeout = 2500 * time.Millisecond
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	rawResults, complete, err := vcsrepo.RawLogDiffSearch(ctx, vcs.RawLogDiffSearchOptions{
		Query: vcs.TextSearchOptions{
			Pattern:         info.Pattern,
			IsRegExp:        info.IsRegExp,
			IsCaseSensitive: info.IsCaseSensitive,
		},
		Paths: vcs.PathOptions{
			IncludePatterns: info.IncludePatterns,
			ExcludePattern:  strv(info.ExcludePattern),
			IsCaseSensitive: info.PathPatternsAreCaseSensitive,
			IsRegExp:        info.PathPatternsAreRegExps,
			// TODO(sqs): use ArgsHint for better perf
		},
		OnlyMatchingHunks: true,
		Args:              args,
	})
	if err != nil {
		return nil, false, err
	}

	limitHit = limitHit || !complete
	if len(rawResults) > maxResults {
		limitHit = true
		rawResults = rawResults[:maxResults]
	}

	results = make([]*commitSearchResult, len(rawResults))
	for i, rawResult := range rawResults {
		commit := rawResult.Commit
		results[i] = &commitSearchResult{
			commit: &commitInfoResolver{
				repository: &repositoryResolver{repo: repo},
				oid:        gitObjectID(commit.ID),
				author: &signatureResolver{
					person: &personResolver{
						name:  commit.Author.Name,
						email: commit.Author.Email,
					},
					date: commit.Author.Date.String(),
				},
				committer: &signatureResolver{
					person: &personResolver{
						name:  commit.Author.Name,
						email: commit.Author.Email,
					},
					date: commit.Committer.Date.String(),
				},
				message: commit.Message,
			},
			diffPreview: &highlightedString{
				value:      rawResult.Diff.Raw,
				highlights: fromVCSHighlights(rawResult.Highlights),
			},
		}
	}
	return results, limitHit, nil
}

var mockSearchCommitDiffsInRepos func(args *repoSearchArgs, combinedQuery resolvedQuery) ([]*searchResult, *searchResultsCommon, error)

// searchCommitDiffsInRepos searches a set of repos for matching commit diffs.
func searchCommitDiffsInRepos(ctx context.Context, args *repoSearchArgs, combinedQuery resolvedQuery) ([]*searchResult, *searchResultsCommon, error) {
	if mockSearchCommitDiffsInRepos != nil {
		return mockSearchCommitDiffsInRepos(args, combinedQuery)
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var (
		err         error
		wg          sync.WaitGroup
		mu          sync.Mutex
		unflattened [][]*commitSearchResult
		common      = &searchResultsCommon{}
	)
	for _, repoRev := range args.Repositories {
		wg.Add(1)
		go func(repoRev repositoryRevision) {
			defer wg.Done()
			var rev string
			if repoRev.Rev != nil {
				rev = *repoRev.Rev
			}
			results, repoLimitHit, searchErr := searchCommitDiffsInRepo(ctx, repoRev.Repo, rev, args.Query, combinedQuery)
			if ctx.Err() != nil {
				// Our request has been canceled, we can just ignore searchRepo for this repo.
				return
			}
			mu.Lock()
			defer mu.Unlock()
			if fatalErr := handleRepoSearchResult(common, repoRev, repoLimitHit, searchErr); fatalErr != nil {
				err = errors.Wrapf(searchErr, "failed to search diffs %s", repoRev.String())
				cancel()
			}
			if len(results) > 0 {
				unflattened = append(unflattened, results)
			}
		}(*repoRev)
	}
	wg.Wait()
	if err != nil {
		return nil, nil, err
	}

	var flattened []*commitSearchResult
	for _, results := range unflattened {
		flattened = append(flattened, results...)
	}
	return commitSearchResultsToSearchResults(flattened), common, nil
}

func commitSearchResultsToSearchResults(results []*commitSearchResult) []*searchResult {
	results2 := make([]*searchResult, len(results))
	for i, result := range results {
		results2[i] = &searchResult{diff: result}
	}
	return results2
}

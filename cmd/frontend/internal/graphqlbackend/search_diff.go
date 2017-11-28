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
	commit  *commitInfoResolver
	preview *highlightedString
}

func (r *commitSearchResult) Commit() *commitInfoResolver { return r.commit }
func (r *commitSearchResult) Preview() *highlightedString { return r.preview }

var mockSearchDiffsInRepo func(ctx context.Context, repoName, rev string, info *patternInfo, combinedQuery resolvedQuery) (results []*commitSearchResult, limitHit bool, err error)

func searchDiffsInRepo(ctx context.Context, repoName, rev string, info *patternInfo, combinedQuery resolvedQuery) (results []*commitSearchResult, limitHit bool, err error) {
	if mockSearchRepo != nil {
		return mockSearchDiffsInRepo(ctx, repoName, rev, info, combinedQuery)
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
	for _, s := range combinedQuery.fieldValues[searchFieldBefore].Values() {
		args = append(args, "--until="+s)
	}
	for _, s := range combinedQuery.fieldValues[searchFieldAfter].Values() {
		args = append(args, "--since="+s)
	}
	for _, s := range combinedQuery.fieldValues[searchFieldAuthor].Values() {
		args = append(args, "--author="+s)
	}
	for _, s := range combinedQuery.fieldValues[searchFieldCommitter].Values() {
		args = append(args, "--committer="+s)
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
			preview: &highlightedString{
				value:      rawResult.Diff.Raw,
				highlights: fromVCSHighlights(rawResult.Highlights),
			},
		}
	}
	return results, limitHit, nil
}

var mockSearchDiffsInRepos func(args *repoSearchArgs, combinedQuery resolvedQuery) ([]*searchResult, *searchResultsCommon, error)

// searchDiffsInRepos searches a set of repos for matching diffs.
func searchDiffsInRepos(ctx context.Context, args *repoSearchArgs, combinedQuery resolvedQuery) ([]*searchResult, *searchResultsCommon, error) {
	if mockSearchDiffsInRepos != nil {
		return mockSearchDiffsInRepos(args, combinedQuery)
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
			results, repoLimitHit, searchErr := searchDiffsInRepo(ctx, repoRev.Repo, rev, args.Query, combinedQuery)
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

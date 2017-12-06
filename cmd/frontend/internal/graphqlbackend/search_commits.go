package graphqlbackend

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/searchquery"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"

	"github.com/pkg/errors"
)

var (
	gitLogSearchTimeout       = mustParseDuration(env.Get("GIT_LOG_SEARCH_TIMEOUT", "5s", "maximum duration for type:commit and type:diff queries before incomplete results are returned"))
	maxGitLogSearchResults, _ = strconv.Atoi(env.Get("GIT_LOG_MAX_RESULTS", "20", "maximum number of results for type:commit and type:diff queries"))
)

func mustParseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		log.Fatal(err)
	}
	return d
}

type diffSearchResult struct {
	diff    *diff
	preview *highlightedString
}

func (r *diffSearchResult) Diff() *diff                 { return r.diff }
func (r *diffSearchResult) Preview() *highlightedString { return r.preview }

type commitSearchResult struct {
	commit         *commitInfoResolver
	refs           []*gitRefResolver
	sourceRefs     []*gitRefResolver
	messagePreview *highlightedString
	diffPreview    *highlightedString
}

func (r *commitSearchResult) Commit() *commitInfoResolver        { return r.commit }
func (r *commitSearchResult) Refs() []*gitRefResolver            { return r.refs }
func (r *commitSearchResult) SourceRefs() []*gitRefResolver      { return r.sourceRefs }
func (r *commitSearchResult) MessagePreview() *highlightedString { return r.messagePreview }
func (r *commitSearchResult) DiffPreview() *highlightedString    { return r.diffPreview }

var mockSearchCommitDiffsInRepo func(ctx context.Context, repoRevs repositoryRevisions, info *patternInfo, combinedQuery searchquery.Query) (results []*commitSearchResult, limitHit bool, err error)

func searchCommitDiffsInRepo(ctx context.Context, repoRevs repositoryRevisions, info *patternInfo, combinedQuery searchquery.Query) (results []*commitSearchResult, limitHit bool, err error) {
	if mockSearchCommitDiffsInRepo != nil {
		return mockSearchCommitDiffsInRepo(ctx, repoRevs, info, combinedQuery)
	}

	textSearchOptions := vcs.TextSearchOptions{
		Pattern:         info.Pattern,
		IsRegExp:        info.IsRegExp,
		IsCaseSensitive: info.IsCaseSensitive,
	}
	return searchCommitsInRepo(ctx, repoRevs, info, combinedQuery, true, textSearchOptions, nil)
}

var mockSearchCommitLogInRepo func(ctx context.Context, repoRevs repositoryRevisions, info *patternInfo, combinedQuery searchquery.Query) (results []*commitSearchResult, limitHit bool, err error)

func searchCommitLogInRepo(ctx context.Context, repoRevs repositoryRevisions, info *patternInfo, combinedQuery searchquery.Query) (results []*commitSearchResult, limitHit bool, err error) {
	if mockSearchCommitLogInRepo != nil {
		return mockSearchCommitLogInRepo(ctx, repoRevs, info, combinedQuery)
	}

	var terms []string
	if info.Pattern != "" {
		terms = append(terms, info.Pattern)
	}
	return searchCommitsInRepo(ctx, repoRevs, info, combinedQuery, false, vcs.TextSearchOptions{}, terms)
}

func searchCommitsInRepo(ctx context.Context, repoRevs repositoryRevisions, info *patternInfo, combinedQuery searchquery.Query, diff bool, textSearchOptions vcs.TextSearchOptions, extraMessageValues []string) (results []*commitSearchResult, limitHit bool, err error) {
	repo, err := localstore.Repos.GetByURI(ctx, repoRevs.repo)
	if err != nil {
		return nil, false, err
	}
	// 🚨 SECURITY: DO NOT REMOVE THIS CHECK! ResolveRev is responsible for ensuring 🚨
	// the user has permissions to access the repository. (It does not actually need to
	// resolve the rev.)
	if _, err := backend.Repos.ResolveRev(ctx, &sourcegraph.ReposResolveRevOp{Repo: repo.ID}); err != nil {
		return nil, false, err
	}

	vcsrepo, err := localstore.RepoVCS.Open(ctx, repo.ID)
	if err != nil {
		return nil, false, err
	}

	args := []string{
		"--unified=0",
		"--no-prefix",
		"--max-count=" + strconv.Itoa(maxGitLogSearchResults+1),
	}
	if info.IsRegExp {
		args = append(args, "--extended-regexp")
	}
	if !combinedQuery.IsCaseSensitive() {
		args = append(args, "--regexp-ignore-case")
	}

	for _, rev := range repoRevs.revs {
		switch {
		case rev.revspec != "":
			if strings.HasPrefix(rev.revspec, "-") {
				// A revspec starting with "-" would be interpreted as a `git log` flag.
				// It would not be a security vulnerability because the flags are checked
				// against a whitelist, but it could cause unexpected errors by (e.g.)
				// changing the format of `git log` to a format that our parser doesn't
				// expect.
				return nil, false, fmt.Errorf("invalid revspec: %q", rev.revspec)
			}
			args = append(args, rev.revspec)

		case rev.refGlob != "":
			args = append(args, "--glob="+rev.refGlob)

		case rev.excludeRefGlob != "":
			args = append(args, "--exclude="+rev.excludeRefGlob)
		}
	}

	beforeValues, _ := combinedQuery.StringValues(searchquery.FieldBefore)
	for _, s := range beforeValues {
		args = append(args, "--until="+s)
	}
	afterValues, _ := combinedQuery.StringValues(searchquery.FieldAfter)
	for _, s := range afterValues {
		args = append(args, "--since="+s)
	}

	// Helper for adding git log flags --grep, --author, and --committer, which all behave similarly.
	var hasSeenGrepLikeFields, hasSeenInvertedGrepLikeFields bool
	addGrepLikeFlags := func(args *[]string, gitLogFlag string, field string, extraValues []string) error {
		values, minusValues := combinedQuery.RegexpPatterns(field)
		values = append(values, extraValues...)

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
	if err := addGrepLikeFlags(&args, "--grep", searchquery.FieldMessage, extraMessageValues); err != nil {
		return nil, false, err
	}
	if err := addGrepLikeFlags(&args, "--author", searchquery.FieldAuthor, nil); err != nil {
		return nil, false, err
	}
	if err := addGrepLikeFlags(&args, "--committer", searchquery.FieldCommitter, nil); err != nil {
		return nil, false, err
	}

	strv := func(s *string) string {
		if s == nil {
			return ""
		}
		return *s
	}

	ctx, cancel := context.WithTimeout(ctx, gitLogSearchTimeout)
	defer cancel()

	rawResults, complete, err := vcsrepo.RawLogDiffSearch(ctx, vcs.RawLogDiffSearchOptions{
		Query: textSearchOptions,
		Paths: vcs.PathOptions{
			IncludePatterns: info.IncludePatterns,
			ExcludePattern:  strv(info.ExcludePattern),
			IsCaseSensitive: info.PathPatternsAreCaseSensitive,
			IsRegExp:        info.PathPatternsAreRegExps,
			// TODO(sqs): use ArgsHint for better perf
		},
		Diff:              diff,
		OnlyMatchingHunks: true,
		Args:              args,
	})
	if err != nil {
		return nil, false, err
	}

	limitHit = limitHit || !complete
	if len(rawResults) > maxGitLogSearchResults {
		limitHit = true
		rawResults = rawResults[:maxGitLogSearchResults]
	}

	repoResolver := &repositoryResolver{repo: repo}
	results = make([]*commitSearchResult, len(rawResults))
	for i, rawResult := range rawResults {
		commit := rawResult.Commit
		results[i] = &commitSearchResult{
			commit: &commitInfoResolver{
				repository: repoResolver,
				oid:        gitObjectID(commit.ID),
				author:     *toSignatureResolver(&commit.Author),
				committer:  toSignatureResolver(commit.Committer),
				message:    commit.Message,
			},
		}

		addRefs := func(dst *[]*gitRefResolver, src []string) {
			for _, ref := range src {
				*dst = append(*dst, &gitRefResolver{
					repo: repoResolver,
					name: ref,
				})
			}
		}
		addRefs(&results[i].refs, rawResult.Refs)
		addRefs(&results[i].sourceRefs, rawResult.SourceRefs)

		// TODO(sqs): properly combine message: and term values for type:commit searches
		if !diff {
			var patString string
			if len(extraMessageValues) > 0 {
				patString = extraMessageValues[0]
				if !combinedQuery.IsCaseSensitive() {
					patString = "(?i:" + patString + ")"
				}
				pat, err := regexp.Compile(patString)
				if err == nil {
					results[i].messagePreview = highlightMatches(pat, []byte(commit.Message))
				}
			} else {
				results[i].messagePreview = &highlightedString{value: string(commit.Message)}
			}
		}

		if rawResult.Diff != nil {
			results[i].diffPreview = &highlightedString{
				value:      rawResult.Diff.Raw,
				highlights: fromVCSHighlights(rawResult.DiffHighlights),
			}
		}
	}
	return results, limitHit, nil
}

func highlightMatches(pattern *regexp.Regexp, data []byte) *highlightedString {
	const maxMatchesPerLine = 25 // arbitrary

	var highlights []*highlightedRange
	for i, line := range bytes.Split(data, []byte("\n")) {
		for _, match := range pattern.FindAllIndex(bytes.ToLower(line), maxMatchesPerLine) {
			highlights = append(highlights, &highlightedRange{
				line:      int32(i + 1),
				character: int32(match[0]),
				length:    int32(match[1] - match[0]),
			})
		}
	}
	return &highlightedString{
		value:      string(data),
		highlights: highlights,
	}
}

var mockSearchCommitDiffsInRepos func(args *repoSearchArgs, combinedQuery searchquery.Query) ([]*searchResult, *searchResultsCommon, error)

// searchCommitDiffsInRepos searches a set of repos for matching commit diffs.
func searchCommitDiffsInRepos(ctx context.Context, args *repoSearchArgs, combinedQuery searchquery.Query) ([]*searchResult, *searchResultsCommon, error) {
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
	for _, repoRev := range args.repos {
		wg.Add(1)
		go func(repoRev repositoryRevisions) {
			defer wg.Done()
			results, repoLimitHit, searchErr := searchCommitDiffsInRepo(ctx, repoRev, args.query, combinedQuery)
			if ctx.Err() != nil {
				// Our request has been canceled, we can just ignore searchRepo for this repo.
				return
			}
			mu.Lock()
			defer mu.Unlock()
			if fatalErr := handleRepoSearchResult(common, repoRev, repoLimitHit, searchErr); fatalErr != nil {
				err = errors.Wrapf(searchErr, "failed to search commit diffs %s", repoRev.String())
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

var mockSearchCommitLogInRepos func(args *repoSearchArgs, combinedQuery searchquery.Query) ([]*searchResult, *searchResultsCommon, error)

// searchCommitLogInRepos searches a set of repos for matching commits.
func searchCommitLogInRepos(ctx context.Context, args *repoSearchArgs, combinedQuery searchquery.Query) ([]*searchResult, *searchResultsCommon, error) {
	if mockSearchCommitLogInRepos != nil {
		return mockSearchCommitLogInRepos(args, combinedQuery)
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
	for _, repoRev := range args.repos {
		wg.Add(1)
		go func(repoRev repositoryRevisions) {
			defer wg.Done()
			results, repoLimitHit, searchErr := searchCommitLogInRepo(ctx, repoRev, args.query, combinedQuery)
			if ctx.Err() != nil {
				// Our request has been canceled, we can just ignore searchRepo for this repo.
				return
			}
			mu.Lock()
			defer mu.Unlock()
			if fatalErr := handleRepoSearchResult(common, repoRev, repoLimitHit, searchErr); fatalErr != nil {
				err = errors.Wrapf(searchErr, "failed to search commit log %s", repoRev.String())
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

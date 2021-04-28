package graphqlbackend

import (
	"context"
	"math"
	"regexp"
	"sync"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"

	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
)

var mockSearchRepositories func(args *search.TextParameters) ([]SearchResultResolver, *streaming.Stats, error)

// searchRepositories searches for repositories by name.
//
// For a repository to match a query, the repository's name must match all of the repo: patterns AND the
// default patterns (i.e., the patterns that are not prefixed with any search field).
func searchRepositories(ctx context.Context, db dbutil.DB, args *search.TextParameters, limit int32, stream Sender) error {
	if mockSearchRepositories != nil {
		results, stats, err := mockSearchRepositories(args)
		stream.Send(SearchEvent{
			Results: results,
			Stats:   statsDeref(stats),
		})
		return err
	}

	ctx, stream, cancel := WithLimit(ctx, stream, int(limit))
	defer cancel()

	fieldAllowlist := map[string]struct{}{
		query.FieldRepo:               {},
		query.FieldRepoGroup:          {},
		query.FieldContext:            {},
		query.FieldType:               {},
		query.FieldDefault:            {},
		query.FieldIndex:              {},
		query.FieldCount:              {},
		query.FieldTimeout:            {},
		query.FieldFork:               {},
		query.FieldArchived:           {},
		query.FieldVisibility:         {},
		query.FieldCase:               {},
		query.FieldRepoHasFile:        {},
		query.FieldRepoHasCommitAfter: {},
		query.FieldPatternType:        {},
		query.FieldSelect:             {},
	}
	// Don't return repo results if the search contains fields that aren't on the allowlist.
	// Matching repositories based whether they contain files at a certain path (etc.) is not yet implemented.
	for field := range args.Query.Fields() {
		if _, ok := fieldAllowlist[field]; !ok {
			return nil
		}
	}

	patternRe := args.PatternInfo.Pattern
	if !args.Query.IsCaseSensitive() {
		patternRe = "(?i)" + patternRe
	}

	pattern, err := regexp.Compile(patternRe)
	if err != nil {
		return err
	}

	// Filter args.Repos by matching their names against the query pattern.
	resolved, err := getRepos(ctx, args.RepoPromise)
	if err != nil {
		return err
	}

	results := make(chan []*search.RepositoryRevisions)
	go func() {
		defer close(results)
		matchRepos(pattern, resolved, results)
	}()

	// Filter the repos if there is a repohasfile: or -repohasfile field.
	if len(args.PatternInfo.FilePatternsReposMustExclude) > 0 || len(args.PatternInfo.FilePatternsReposMustInclude) > 0 {
		// Fallback to batch for reposToAdd
		var repos []*search.RepositoryRevisions
		for matched := range results {
			repos = append(repos, matched...)
		}
		repos, err = reposToAdd(ctx, db, args, repos)
		if err != nil {
			return err
		}
		stream.Send(SearchEvent{
			Results: repoRevsToSearchResultResolver(ctx, db, repos),
		})
		return nil
	}

	for repos := range results {
		stream.Send(SearchEvent{
			Results: repoRevsToSearchResultResolver(ctx, db, repos),
		})
	}

	return nil
}

func repoRevsToSearchResultResolver(ctx context.Context, db dbutil.DB, repos []*search.RepositoryRevisions) []SearchResultResolver {
	results := make([]SearchResultResolver, 0, len(repos))
	for _, r := range repos {
		revs, err := r.ExpandedRevSpecs(ctx)
		if err != nil { // fallback to just return revspecs
			revs = r.RevSpecs()
		}
		for _, rev := range revs {
			rr := NewRepositoryResolver(db, r.Repo.ToRepo())
			rr.RepoMatch.Rev = rev
			results = append(results, rr)
		}
	}
	return results
}

func matchRepos(pattern *regexp.Regexp, resolved []*search.RepositoryRevisions, results chan<- []*search.RepositoryRevisions) {
	/*
		Local benchmarks showed diminishing returns for higher levels of concurrency.
		5 workers seems to be a good trade-off for now. We might want to revisit this
		benchmark over time.

		go test -cpu 1,2,3,4,5,6,7,8,9,10 -count=5 -bench=SearchRepo .

		   goos: darwin
		   goarch: amd64
		   pkg: github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend
		   BenchmarkSearchRepositories       	      13	 132088878 ns/op
		   BenchmarkSearchRepositories-2     	      16	  69968357 ns/op
		   BenchmarkSearchRepositories-3     	      24	  48294832 ns/op
		   BenchmarkSearchRepositories-4     	      28	  42497674 ns/op
		   BenchmarkSearchRepositories-5     	      27	  42851670 ns/op
		   BenchmarkSearchRepositories-6     	      28	  39327860 ns/op
		   BenchmarkSearchRepositories-7     	      27	  38198665 ns/op
		   BenchmarkSearchRepositories-8     	      28	  38877182 ns/op
		   BenchmarkSearchRepositories-9     	      26	  42457771 ns/op
		   BenchmarkSearchRepositories-10    	      26	  40519692 ns/op
	*/
	step := len(resolved) / 5 // for benchmarking, replace 5 with runtime.GOMAXPROCS(0)
	if step == 0 {
		step = len(resolved)
	} else {
		step += 1
	}

	var wg sync.WaitGroup
	offset := 0
	for offset < len(resolved) {
		next := offset + step
		if next > len(resolved) {
			next = len(resolved)
		}
		wg.Add(1)
		go func(repos []*search.RepositoryRevisions) {
			defer wg.Done()

			var matched []*search.RepositoryRevisions
			for _, r := range repos {
				if pattern.MatchString(string(r.Repo.Name)) {
					matched = append(matched, r)
				}
			}
			if len(matched) > 0 {
				results <- matched
			}
		}(resolved[offset:next])
		offset = next
	}

	wg.Wait()
}

// reposToAdd determines which repositories should be included in the result set based on whether they fit in the subset
// of repostiories specified in the query's `repohasfile` and `-repohasfile` fields if they exist.
func reposToAdd(ctx context.Context, db dbutil.DB, args *search.TextParameters, repos []*search.RepositoryRevisions) ([]*search.RepositoryRevisions, error) {
	// matchCounts will contain the count of repohasfile patterns that matched.
	// For negations, we will explicitly set this to -1 if it matches.
	matchCounts := make(map[api.RepoID]int)
	if len(args.PatternInfo.FilePatternsReposMustInclude) > 0 {
		for _, pattern := range args.PatternInfo.FilePatternsReposMustInclude {
			// The high FileMatchLimit here is to make sure we get all the repo matches we can. Setting it to
			// len(repos) could mean we miss some repos since there could be for example len(repos) file matches in
			// the first repo and some more in other repos.
			p := search.TextPatternInfo{IsRegExp: true, FileMatchLimit: math.MaxInt32, IncludePatterns: []string{pattern}, PathPatternsAreCaseSensitive: false, PatternMatchesContent: true, PatternMatchesPath: true}
			q, err := query.ParseLiteral("file:" + pattern)
			if err != nil {
				return nil, err
			}
			newArgs := *args
			newArgs.PatternInfo = &p
			newArgs.RepoPromise = (&search.Promise{}).Resolve(repos)
			newArgs.Query = q
			newArgs.UseFullDeadline = true
			matches, _, err := searchFilesInReposBatch(ctx, db, &newArgs)
			if err != nil {
				return nil, err
			}

			// deduplicate repo results
			matchedIDs := make(map[api.RepoID]struct{})
			for _, m := range matches {
				matchedIDs[m.Repo.ID] = struct{}{}
			}

			// increment the count for all seen repos
			for id := range matchedIDs {
				matchCounts[id] += 1
			}
		}
	} else {
		// Default to including all the repos, then excluding some of them below.
		for _, r := range repos {
			matchCounts[r.Repo.ID] = 0
		}
	}

	if len(args.PatternInfo.FilePatternsReposMustExclude) > 0 {
		for _, pattern := range args.PatternInfo.FilePatternsReposMustExclude {
			p := search.TextPatternInfo{IsRegExp: true, FileMatchLimit: math.MaxInt32, IncludePatterns: []string{pattern}, PathPatternsAreCaseSensitive: false, PatternMatchesContent: true, PatternMatchesPath: true}
			q, err := query.ParseLiteral("file:" + pattern)
			if err != nil {
				return nil, err
			}
			newArgs := *args
			newArgs.PatternInfo = &p
			rp := (&search.Promise{}).Resolve(repos)
			newArgs.RepoPromise = rp
			newArgs.Query = q
			newArgs.UseFullDeadline = true
			matches, _, err := searchFilesInReposBatch(ctx, db, &newArgs)
			if err != nil {
				return nil, err
			}
			for _, m := range matches {
				matchCounts[m.Repo.ID] = -1
			}
		}
	}

	var rsta []*search.RepositoryRevisions
	for _, r := range repos {
		if count, ok := matchCounts[r.Repo.ID]; ok && count == len(args.PatternInfo.FilePatternsReposMustInclude) {
			rsta = append(rsta, r)
		}
	}

	return rsta, nil
}

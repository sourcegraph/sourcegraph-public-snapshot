package run

import (
	"context"
	"math"
	"regexp"
	"runtime"

	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/sourcegraph/sourcegraph/internal/types"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/search/unindexed"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

var MockSearchRepositories func(args *search.TextParameters) ([]result.Match, *streaming.Stats, error)

// SearchRepositories searches for repositories by name.
//
// For a repository to match a query, the repository's name must match all of the repo: patterns AND the
// default patterns (i.e., the patterns that are not prefixed with any search field).
func SearchRepositories(ctx context.Context, args *search.TextParameters, limit int32, stream streaming.Sender) (err error) {
	if MockSearchRepositories != nil {
		results, stats, err := MockSearchRepositories(args)
		stream.Send(streaming.SearchEvent{
			Results: results,
			Stats:   stats.Deref(),
		})
		return err
	}

	tr, ctx := trace.New(ctx, "run.SearchRepositories", "")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	ctx, stream, cancel := streaming.WithLimit(ctx, stream, int(limit))
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
			tr.LazyPrintf("contains dissallowed field: %s", field)
			return nil
		}
	}

	patternRe := args.PatternInfo.Pattern
	if !args.Query.IsCaseSensitive() {
		patternRe = "(?i)" + patternRe
	}

	tr.LogFields(
		otlog.String("pattern", patternRe),
		otlog.Int32("limit", limit))

	pattern, err := regexp.Compile(patternRe)
	if err != nil {
		return err
	}

	// Filter args.Repos by matching their names against the query pattern.
	tr.LogFields(otlog.Int("resolved.len", args.Repos.Len()))

	results := make(chan []*types.RepoName)
	go func() {
		defer close(results)
		matchRepos(pattern, args.Repos, results)
	}()

	// Filter the repos if there is a repohasfile: or -repohasfile field.
	if len(args.PatternInfo.FilePatternsReposMustExclude) > 0 || len(args.PatternInfo.FilePatternsReposMustInclude) > 0 {
		// Fallback to batch for reposToAdd
		var matched []*types.RepoName
		for repos := range results {
			matched = append(matched, repos...)
		}

		matched, err = reposToAdd(ctx, args, matched)
		if err != nil {
			return err
		}

		stream.Send(streaming.SearchEvent{
			Results: toRepoMatches(args.Repos.RepoRevs, matched),
		})
		return nil
	}

	count := 0
	for matched := range results {
		count += len(matched)
		stream.Send(streaming.SearchEvent{
			Results: toRepoMatches(args.Repos.RepoRevs, matched),
		})
	}
	tr.LogFields(otlog.Int("matched.len", count))

	return nil
}

func toRepoMatches(repoRevs map[api.RepoName]search.RevSpecs, matched []*types.RepoName) []result.Match {
	matches := make([]result.Match, 0, len(matched))
	for _, r := range matched {
		revs := repoRevs[r.Name]
		if len(revs) == 0 {
			revs = search.DefaultRevSpecs
		}
		for _, rev := range revs {
			if !rev.IsGlob() {
				matches = append(matches, &result.RepoMatch{
					Name: r.Name,
					ID:   r.ID,
					Rev:  rev.RevSpec,
				})
			}
		}
	}
	return matches
}

func matchRepos(pattern *regexp.Regexp, resolved *search.Repos, results chan<- []*types.RepoName) {
	/*
		goos: linux
		goarch: amd64
		pkg: github.com/sourcegraph/sourcegraph/internal/search/run
		cpu: Intel(R) Core(TM) i9-8950HK CPU @ 2.90GHz
		BenchmarkSearchRepositories-8   	      39	  31514267 ns/op
		BenchmarkSearchRepositories-8   	      26	  38898255 ns/op
		BenchmarkSearchRepositories-8   	      36	  31482727 ns/op
		BenchmarkSearchRepositories-8   	      33	  30513691 ns/op
		BenchmarkSearchRepositories-8   	      30	  37038388 ns/op
		BenchmarkSearchRepositories-8   	      30	  38095363 ns/op
		BenchmarkSearchRepositories-8   	      36	  39347784 ns/op
		BenchmarkSearchRepositories-8   	      28	  41431416 ns/op
		BenchmarkSearchRepositories-8   	      30	  41695426 ns/op
		BenchmarkSearchRepositories-8   	      28	  39782412 ns/op
		PASS
		ok  	github.com/sourcegraph/sourcegraph/internal/search/run	18.729s
	*/
	workers := runtime.NumCPU()
	for _, repos := range [2][]*types.RepoName{resolved.Private.Repos, resolved.Public.Repos} {
		limit := len(repos) / workers

		last := make(chan struct{})
		close(last)

		for i := 0; i < workers; i++ {
			page := repos[i*limit : i*limit+limit]
			if i == workers-1 {
				page = repos[i*limit:]
			}

			wait := last
			done := make(chan struct{})
			last = done

			go func() {
				defer close(done)

				var matched []*types.RepoName
				for _, r := range page {
					if !resolved.Excludes(r.ID) && pattern.MatchString(string(r.Name)) {
						matched = append(matched, r)
					}
				}

				// Wait for the previous chunk to send its matches
				// before sending ours.
				<-wait
				results <- matched
			}()
		}

		<-last
	}
}

func reposContainingPath(ctx context.Context, args *search.TextParameters, repos []*types.RepoName, pattern string) ([]*result.FileMatch, error) {
	// Use a max FileMatchLimit to ensure we get all the repo matches we
	// can. Setting it to len(repos) could mean we miss some repos since
	// there could be for example len(repos) file matches in the first repo
	// and some more in other repos. deduplicate repo results
	p := search.TextPatternInfo{
		IsRegExp:                     true,
		FileMatchLimit:               math.MaxInt32,
		IncludePatterns:              []string{pattern},
		PathPatternsAreCaseSensitive: false,
		PatternMatchesContent:        true,
		PatternMatchesPath:           true,
	}
	q, err := query.ParseLiteral("file:" + pattern)
	if err != nil {
		return nil, err
	}
	newArgs := *args
	newArgs.PatternInfo = &p
	newArgs.Repos = search.NewRepos(repos...)
	newArgs.Query = q
	newArgs.UseFullDeadline = true
	matches, _, err := unindexed.SearchFilesInReposBatch(ctx, &newArgs)
	if err != nil {
		return nil, err
	}
	return matches, err
}

// reposToAdd determines which repositories should be included in the result set based on whether they fit in the subset
// of repostiories specified in the query's `repohasfile` and `-repohasfile` fields if they exist.
func reposToAdd(ctx context.Context, args *search.TextParameters, repos []*types.RepoName) ([]*types.RepoName, error) {
	// matchCounts will contain the count of repohasfile patterns that matched.
	// For negations, we will explicitly set this to -1 if it matches.
	matchCounts := make(map[api.RepoID]int)
	if len(args.PatternInfo.FilePatternsReposMustInclude) > 0 {
		for _, pattern := range args.PatternInfo.FilePatternsReposMustInclude {
			matches, err := reposContainingPath(ctx, args, repos, pattern)
			if err != nil {
				return nil, err
			}

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
			matchCounts[r.ID] = 0
		}
	}

	if len(args.PatternInfo.FilePatternsReposMustExclude) > 0 {
		for _, pattern := range args.PatternInfo.FilePatternsReposMustExclude {
			matches, err := reposContainingPath(ctx, args, repos, pattern)
			if err != nil {
				return nil, err
			}
			for _, m := range matches {
				matchCounts[m.Repo.ID] = -1
			}
		}
	}

	var rsta []*types.RepoName
	for _, r := range repos {
		if count, ok := matchCounts[r.ID]; ok && count == len(args.PatternInfo.FilePatternsReposMustInclude) {
			rsta = append(rsta, r)
		}
	}
	return rsta, nil
}

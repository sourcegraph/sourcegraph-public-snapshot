package graphqlbackend

import (
	"fmt"
	"strconv"

	"github.com/sourcegraph/sourcegraph/internal/search"
)

func number(i int) string {
	if i < 1000 {
		return strconv.Itoa(i)
	}
	if i < 10000 {
		return fmt.Sprintf("%d,%d", i/1000, i%1000)
	}
	return fmt.Sprintf("%dk", i/1000)
}

func plural(one, many string, n int) string {
	if n == 1 {
		return one
	}
	return many
}

func repositoryCloningHandler(resultsResolver *SearchResultsResolver) (search.Skipped, bool) {
	repos := resultsResolver.Cloning()
	if len(repos) == 0 {
		return search.Skipped{}, false
	}

	amount := number(len(repos))
	return search.Skipped{
		Reason: search.RepositoryCloning,
		Title:  fmt.Sprintf("%s cloning", amount),
		// TODO sample of repos in message
		Message:  fmt.Sprintf("%s %s could not be searched since %s still cloning. Try searching again or reducing the scope of your query with `repo:`,  `repogroup:` or other filters.", amount, plural("repository", "repositories", len(repos)), plural("it is", "they are", len(repos))),
		Severity: search.SeverityInfo,
	}, true
}

func repositoryMissingHandler(resultsResolver *SearchResultsResolver) (search.Skipped, bool) {
	repos := resultsResolver.Missing()
	if len(repos) == 0 {
		return search.Skipped{}, false
	}

	amount := number(len(repos))
	return search.Skipped{
		Reason: search.RepositoryMissing,
		Title:  fmt.Sprintf("%s missing", amount),
		// TODO sample of repos in message
		Message:  fmt.Sprintf("%s %s could not be searched. Try reducing the scope of your query with `repo:`, `repogroup:` or other filters.", amount, plural("repository", "repositories", len(repos))),
		Severity: search.SeverityInfo,
	}, true
}

func shardTimeoutHandler(resultsResolver *SearchResultsResolver) (search.Skipped, bool) {
	// This is not the same, but once we expose this more granular details
	// from our backend it will be shard specific.
	repos := resultsResolver.Timedout()
	if len(repos) == 0 {
		return search.Skipped{}, false
	}

	amount := number(len(repos))
	return search.Skipped{
		Reason: search.ShardTimeout,
		Title:  fmt.Sprintf("%s timedout", amount),
		// TODO sample of repos in message
		Message:  fmt.Sprintf("%s %s could not be searched in time. Try reducing the scope of your query with `repo:`, `repogroup:` or other filters.", amount, plural("repository", "repositories", len(repos))),
		Severity: search.SeverityWarn,
	}, true
}

func shardMatchLimitHandler(resultsResolver *SearchResultsResolver) (search.Skipped, bool) {
	// We don't have the details of repo vs shard vs document limits yet. So
	// we just pretend all our shard limits.
	if !resultsResolver.LimitHit() {
		return search.Skipped{}, false
	}

	return search.Skipped{
		Reason:   search.ShardMatchLimit,
		Title:    "result limit hit",
		Message:  "Not all results have been returned due to hitting a match limit. Sourcegraph has limits for the number of results returned from a line, document and repository.",
		Severity: search.SeverityInfo,
	}, true
}

func excludedForkHandler(resultsResolver *SearchResultsResolver) (search.Skipped, bool) {
	forks := resultsResolver.excluded.forks
	if forks == 0 {
		return search.Skipped{}, false
	}

	amount := number(forks)
	return search.Skipped{
		Reason:   search.ExcludedFork,
		Title:    fmt.Sprintf("%s forked", amount),
		Message:  "By default we exclude forked repositories. Include them with `fork:yes` in your query.",
		Severity: search.SeverityInfo,
		Suggested: &search.SkippedSuggested{
			Title:           "include forked",
			QueryExpression: "fork:yes",
		},
	}, true
}

func excludedArchiveHandler(resultsResolver *SearchResultsResolver) (search.Skipped, bool) {
	archived := resultsResolver.excluded.archived
	if archived == 0 {
		return search.Skipped{}, false
	}

	amount := number(archived)
	return search.Skipped{
		Reason:   search.ExcludedArchive,
		Title:    fmt.Sprintf("%s archived", amount),
		Message:  "By default we exclude archived repositories. Include them with `archived:yes` in your query.",
		Severity: search.SeverityInfo,
		Suggested: &search.SkippedSuggested{
			Title:           "include archived",
			QueryExpression: "archived:yes",
		},
	}, true
}

// TODO implement all skipped reasons
var skippedHandlers = []func(*SearchResultsResolver) (search.Skipped, bool){
	repositoryMissingHandler,
	repositoryCloningHandler,
	// documentMatchLimitHandler,
	shardMatchLimitHandler,
	// repositoryLimitHandler,
	shardTimeoutHandler,
	excludedForkHandler,
	excludedArchiveHandler,
}

// Progress builds a progress event from a final results resolver.
func (sr *SearchResultsResolver) Progress() search.Progress {
	skipped := []search.Skipped{}

	for _, handler := range skippedHandlers {
		if sk, ok := handler(sr); ok {
			skipped = append(skipped, sk)
		}
	}

	return search.Progress{
		RepositoriesCount: intPtr(int(sr.RepositoriesCount())),
		MatchCount:        int(sr.MatchCount()),
		DurationMs:        int(sr.ElapsedMilliseconds()),
		Skipped:           skipped,
	}
}

func intPtr(i int) *int {
	return &i
}

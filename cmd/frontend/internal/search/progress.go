package search

import (
	"fmt"
	"strconv"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
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

func repositoryCloningHandler(resultsResolver *graphqlbackend.SearchResultsResolver) (search.Skipped, bool) {
	repos := resultsResolver.Cloning()
	if len(repos) == 0 {
		return search.Skipped{}, false
	}

	amount := number(len(repos))
	return search.Skipped{
		Reason: search.RepositoryCloning,
		Title:  fmt.Sprintf("%s cloning", amount),
		// TODO sample of repos in message
		Message:  fmt.Sprintf("%s repositories could not be searched since they are still cloning. Try searching again or reducing the scope of your query with repo: repogroup: or other filters.", amount),
		Severity: search.SeverityWarn,
	}, true
}

func repositoryMissingHandler(resultsResolver *graphqlbackend.SearchResultsResolver) (search.Skipped, bool) {
	repos := resultsResolver.Missing()
	if len(repos) == 0 {
		return search.Skipped{}, false
	}

	amount := number(len(repos))
	return search.Skipped{
		Reason: search.RepositoryMissing,
		Title:  fmt.Sprintf("%s missing", amount),
		// TODO sample of repos in message
		Message:  fmt.Sprintf("%s repositories could not be searched. Try reducing the scope of your query with repo: repogroup: or other filters.", amount),
		Severity: search.SeverityWarn,
	}, true
}

func shardTimeoutHandler(resultsResolver *graphqlbackend.SearchResultsResolver) (search.Skipped, bool) {
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
		Message:  fmt.Sprintf("%s repositories could not be searched in time. Try reducing the scope of your query with repo: repogroup: or other filters.", amount),
		Severity: search.SeverityWarn,
	}, true
}

func shardMatchLimitHandler(resultsResolver *graphqlbackend.SearchResultsResolver) (search.Skipped, bool) {
	// We don't have the details of repo vs shard vs document limits yet. So
	// we just pretend all our shard limits.
	if !resultsResolver.LimitHit() {
		return search.Skipped{}, false
	}

	return search.Skipped{
		Reason:   search.ShardMatchLimit,
		Title:    "result limit hit",
		Message:  "Not all results have been returned due to hitting a match limit. Sourcegraph has limits for the number of results returned from a line, document and repository.",
		Severity: search.SeverityWarn,
	}, true
}

// TODO implement all skipped reasons
var skippedHandlers = []func(*graphqlbackend.SearchResultsResolver) (search.Skipped, bool){
	repositoryMissingHandler,
	repositoryCloningHandler,
	// documentMatchLimitHandler,
	shardMatchLimitHandler,
	// repositoryLimitHandler,
	shardTimeoutHandler,
	// excludedForkHandler,
	// excludedArchiveHandler,
}

// progressFromResolver builds a progress event from a final results resolver.
func progressFromResolver(resultsResolver *graphqlbackend.SearchResultsResolver) search.Progress {
	skipped := []search.Skipped{}

	for _, handler := range skippedHandlers {
		if sk, ok := handler(resultsResolver); ok {
			skipped = append(skipped, sk)
		}
	}

	return search.Progress{
		Done:              true,
		RepositoriesCount: intPtr(int(resultsResolver.RepositoriesCount())),
		MatchCount:        int(resultsResolver.MatchCount()),
		DurationMs:        int(resultsResolver.ElapsedMilliseconds()),
		Skipped:           skipped,
	}
}

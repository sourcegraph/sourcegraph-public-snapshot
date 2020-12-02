package search

import (
	"fmt"
	"strconv"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

type eventProgress struct {
	Done              bool           `json:"done"`
	RepositoriesCount *int           `json:"repositoriesCount"`
	MatchCount        int            `json:"matchCount"`
	DurationMs        int            `json:"durationMs"`
	Skipped           []eventSkipped `json:"skipped"`
}

type eventSkipped struct {
	Reason    skippedReason   `json:"reason"`
	Title     string          `json:"title"`
	Message   string          `json:"message"`
	Severity  severityType    `json:"severity"`
	Suggested *eventSuggested `json:"suggested,omitempty"`
}

type eventSuggested struct {
	Title           string `json:"title"`
	QueryExpression string `json:"queryExpression"`
}

type skippedReason string

const (
	documentMatchLimit skippedReason = "document-match-limit"
	shardMatchLimit                  = "shard-match-limit"
	repositoryLimit                  = "repository-limit"
	shardTimeout                     = "shard-timeout"
	repositoryCloning                = "repository-cloning"
	repositoryMissing                = "repository-missing"
	excludedFork                     = "repository-fork"
	excludedArchive                  = "excluded-archive"
)

type severityType string

const (
	severityInfo severityType = "info"
	severityWarn              = "warn"
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

func repositoryCloningHandler(resultsResolver *graphqlbackend.SearchResultsResolver) (eventSkipped, bool) {
	repos := resultsResolver.Cloning()
	if len(repos) == 0 {
		return eventSkipped{}, false
	}

	amount := number(len(repos))
	return eventSkipped{
		Reason: repositoryCloning,
		Title:  fmt.Sprintf("%s cloning", amount),
		// TODO sample of repos in message
		Message:  fmt.Sprintf("%s repositories could not be searched since they are still cloning. Try searching again or reducing the scope of your query with repo: repogroup: or other filters.", amount),
		Severity: severityWarn,
	}, true
}

func repositoryMissingHandler(resultsResolver *graphqlbackend.SearchResultsResolver) (eventSkipped, bool) {
	repos := resultsResolver.Missing()
	if len(repos) == 0 {
		return eventSkipped{}, false
	}

	amount := number(len(repos))
	return eventSkipped{
		Reason: repositoryMissing,
		Title:  fmt.Sprintf("%s missing", amount),
		// TODO sample of repos in message
		Message:  fmt.Sprintf("%s repositories could not be searched. Try reducing the scope of your query with repo: repogroup: or other filters.", amount),
		Severity: severityWarn,
	}, true
}

func shardTimeoutHandler(resultsResolver *graphqlbackend.SearchResultsResolver) (eventSkipped, bool) {
	// This is not the same, but once we expose this more granular details
	// from our backend it will be shard specific.
	repos := resultsResolver.Timedout()
	if len(repos) == 0 {
		return eventSkipped{}, false
	}

	amount := number(len(repos))
	return eventSkipped{
		Reason: shardTimeout,
		Title:  fmt.Sprintf("%s timedout", amount),
		// TODO sample of repos in message
		Message:  fmt.Sprintf("%s repositories could not be searched in time. Try reducing the scope of your query with repo: repogroup: or other filters.", amount),
		Severity: severityWarn,
	}, true
}

func shardMatchLimitHandler(resultsResolver *graphqlbackend.SearchResultsResolver) (eventSkipped, bool) {
	// We don't have the details of repo vs shard vs document limits yet. So
	// we just pretend all our shard limits.
	if !resultsResolver.LimitHit() {
		return eventSkipped{}, false
	}

	return eventSkipped{
		Reason:   shardMatchLimit,
		Title:    "result limit hit",
		Message:  "Not all results have been returned due to hitting a match limit. Sourcegraph has limits for the number of results returned from a line, document and repository.",
		Severity: severityWarn,
	}, true
}

// TODO implement all skipped reasons
var skippedHandlers = []func(*graphqlbackend.SearchResultsResolver) (eventSkipped, bool){
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
func progressFromResolver(resultsResolver *graphqlbackend.SearchResultsResolver) eventProgress {
	skipped := []eventSkipped{}

	for _, handler := range skippedHandlers {
		if sk, ok := handler(resultsResolver); ok {
			skipped = append(skipped, sk)
		}
	}

	return eventProgress{
		Done:              true,
		RepositoriesCount: intPtr(int(resultsResolver.RepositoriesCount())),
		MatchCount:        int(resultsResolver.MatchCount()),
		DurationMs:        int(resultsResolver.ElapsedMilliseconds()),
		Skipped:           skipped,
	}
}

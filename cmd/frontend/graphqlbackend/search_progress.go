package graphqlbackend

import (
	"fmt"
	"strconv"
	"strings"

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

func skippedReposHandler(repos []*RepositoryResolver, titleVerb, messageReason string, base search.Skipped) (search.Skipped, bool) {
	if len(repos) == 0 {
		return search.Skipped{}, false
	}

	amount := number(len(repos))
	base.Title = fmt.Sprintf("%s %s", amount, titleVerb)

	if len(repos) == 1 {
		base.Message = fmt.Sprintf("`%s` %s. Try searching again or reducing the scope of your query with `repo:`,  `repogroup:` or other filters.", repos[0].Name(), messageReason)
	} else {
		sampleSize := 10
		if sampleSize > len(repos) {
			sampleSize = len(repos)
		}

		var b strings.Builder
		_, _ = fmt.Fprintf(&b, "%s repositories %s. Try searching again or reducing the scope of your query with `repo:`, `repogroup:` or other filters.", amount, messageReason)
		for _, repo := range repos[:sampleSize] {
			_, _ = fmt.Fprintf(&b, "\n* `%s`", repo.Name())
		}
		if sampleSize < len(repos) {
			b.WriteString("\n* ...")
		}
		base.Message = b.String()
	}

	return base, true
}

func repositoryCloningHandler(resultsResolver *SearchResultsResolver) (search.Skipped, bool) {
	repos := resultsResolver.Cloning()
	messageReason := fmt.Sprintf("could not be searched since %s still cloning", plural("it is", "they are", len(repos)))
	return skippedReposHandler(repos, "cloning", messageReason, search.Skipped{
		Reason:   search.RepositoryCloning,
		Severity: search.SeverityInfo,
	})
}

func repositoryMissingHandler(resultsResolver *SearchResultsResolver) (search.Skipped, bool) {
	return skippedReposHandler(resultsResolver.Missing(), "missing", "could not be searched", search.Skipped{
		Reason:   search.RepositoryMissing,
		Severity: search.SeverityInfo,
	})
}

func shardTimeoutHandler(resultsResolver *SearchResultsResolver) (search.Skipped, bool) {
	// This is not the same, but once we expose this more granular details
	// from our backend it will be shard specific.
	return skippedReposHandler(resultsResolver.Timedout(), "timed out", "could not be searched in time", search.Skipped{
		Reason:   search.ShardTimeout,
		Severity: search.SeverityWarn,
	})
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

package api

import (
	"fmt"
	"strconv"
	"strings"
)

// BuildProgressEvent builds a progress event from a final results resolver.
func BuildProgressEvent(stats ProgressStats) Progress {
	skipped := []Skipped{}

	for _, handler := range skippedHandlers {
		if sk, ok := handler(stats); ok {
			skipped = append(skipped, sk)
		}
	}

	return Progress{
		RepositoriesCount: stats.RepositoriesCount,
		MatchCount:        stats.MatchCount,
		DurationMs:        stats.ElapsedMilliseconds,
		Skipped:           skipped,
		Trace:             stats.Trace,
	}
}

type Namer interface {
	Name() string
}

type ProgressStats struct {
	MatchCount          int
	ElapsedMilliseconds int
	RepositoriesCount   *int
	ExcludedArchived    int
	ExcludedForks       int

	Timedout []Namer
	Missing  []Namer
	Cloning  []Namer

	LimitHit bool

	// SuggestedLimit is what to suggest to the user for count if needed.
	SuggestedLimit int

	Trace string // only filled if requested
}

func skippedReposHandler(repos []Namer, titleVerb, messageReason string, base Skipped) (Skipped, bool) {
	if len(repos) == 0 {
		return Skipped{}, false
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

func repositoryCloningHandler(resultsResolver ProgressStats) (Skipped, bool) {
	repos := resultsResolver.Cloning
	messageReason := fmt.Sprintf("could not be searched since %s still cloning", plural("it is", "they are", len(repos)))
	return skippedReposHandler(repos, "cloning", messageReason, Skipped{
		Reason:   RepositoryCloning,
		Severity: SeverityInfo,
	})
}

func repositoryMissingHandler(resultsResolver ProgressStats) (Skipped, bool) {
	return skippedReposHandler(resultsResolver.Missing, "missing", "could not be searched", Skipped{
		Reason:   RepositoryMissing,
		Severity: SeverityInfo,
	})
}

func shardTimeoutHandler(resultsResolver ProgressStats) (Skipped, bool) {
	// This is not the same, but once we expose this more granular details
	// from our backend it will be shard specific.
	return skippedReposHandler(resultsResolver.Timedout, "timed out", "could not be searched in time", Skipped{
		Reason:   ShardTimeout,
		Severity: SeverityWarn,
	})
}

func shardMatchLimitHandler(resultsResolver ProgressStats) (Skipped, bool) {
	// We don't have the details of repo vs shard vs document limits yet. So
	// we just pretend all our shard limits.
	if !resultsResolver.LimitHit {
		return Skipped{}, false
	}

	var suggest *SkippedSuggested
	if resultsResolver.SuggestedLimit > 0 {
		suggest = &SkippedSuggested{
			Title:           "increase limit",
			QueryExpression: fmt.Sprintf("count:%d", resultsResolver.SuggestedLimit),
		}
	}

	return Skipped{
		Reason:    ShardMatchLimit,
		Title:     "result limit hit",
		Message:   "Not all results have been returned due to hitting a match limit. Sourcegraph has limits for the number of results returned from a line, document and repository.",
		Severity:  SeverityInfo,
		Suggested: suggest,
	}, true
}

func excludedForkHandler(resultsResolver ProgressStats) (Skipped, bool) {
	forks := resultsResolver.ExcludedForks
	if forks == 0 {
		return Skipped{}, false
	}

	amount := number(forks)
	return Skipped{
		Reason:   ExcludedFork,
		Title:    fmt.Sprintf("%s forked", amount),
		Message:  "By default we exclude forked repositories. Include them with `fork:yes` in your query.",
		Severity: SeverityInfo,
		Suggested: &SkippedSuggested{
			Title:           "include forked",
			QueryExpression: "fork:yes",
		},
	}, true
}

func excludedArchiveHandler(resultsResolver ProgressStats) (Skipped, bool) {
	archived := resultsResolver.ExcludedArchived
	if archived == 0 {
		return Skipped{}, false
	}

	amount := number(archived)
	return Skipped{
		Reason:   ExcludedArchive,
		Title:    fmt.Sprintf("%s archived", amount),
		Message:  "By default we exclude archived repositories. Include them with `archived:yes` in your query.",
		Severity: SeverityInfo,
		Suggested: &SkippedSuggested{
			Title:           "include archived",
			QueryExpression: "archived:yes",
		},
	}, true
}

// TODO implement all skipped reasons
var skippedHandlers = []func(stats ProgressStats) (Skipped, bool){
	repositoryMissingHandler,
	repositoryCloningHandler,
	// documentMatchLimitHandler,
	shardMatchLimitHandler,
	// repositoryLimitHandler,
	shardTimeoutHandler,
	excludedForkHandler,
	excludedArchiveHandler,
}

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

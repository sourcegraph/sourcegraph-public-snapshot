package main

import (
	"fmt"
	"time"

	"github.com/buildkite/go-buildkite/v3/buildkite"
)

// findConsecutiveFailures scans the given set of builds for a series of consecutive
// failures. If returns all failed builds encountered as soon as it finds a passed build.
//
// Assumes builds are ordered from neweset to oldest.
func findConsecutiveFailures(
	builds []buildkite.Build,
	threshold int,
	timeout time.Duration,
) (
	failedCommits []CommitInfo,
	thresholdExceeded bool,
	buildsScanned int,
) {
	var consecutiveFailures int
	var build buildkite.Build
	for buildsScanned, build = range builds {
		if isBuildPassed(build) {
			// If we find a passed build we are done
			return
		} else if !isBuildFailed(build, timeout) {
			// we're only safe if non-failures are actually passed, otherwise
			// keep looking
			continue
		}

		var author string
		if build.Author != nil {
			author = fmt.Sprintf("%s (%s)", build.Author.Name, build.Author.Email)
		}

		// Process this build as a failure
		consecutiveFailures += 1
		commit := CommitInfo{
			Author: author,
			Commit: maybeString(build.Commit),
		}
		if build.Number != nil {
			commit.BuildNumber = *build.Number
			commit.BuildURL = maybeString(build.URL)
		}
		if build.CreatedAt != nil {
			commit.BuildCreated = build.CreatedAt.Time
		}
		failedCommits = append(failedCommits, commit)
		if consecutiveFailures >= threshold {
			thresholdExceeded = true
		}
	}

	return
}

func maybeString(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}

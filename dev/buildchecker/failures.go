package main

import (
	"fmt"
	"time"

	"github.com/buildkite/go-buildkite/v3/buildkite"

	"github.com/sourcegraph/sourcegraph/lib/pointers"
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
) {
	var consecutiveFailures int
	var build buildkite.Build
	for _, build = range builds {
		if isBuildScheduled(build) {
			// a Scheduled build should not be considered as part of the set that determines whether
			// main is locked.
			// An exmaple of a scheduled build is the nightly release healthcheck build at:
			// https://buildkite.com/sourcegraph/sourcegraph/settings/schedules/d0b2e4ea-e2df-4fb5-b90e-db88fddb1b76
			continue
		}
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
			Commit: pointers.DerefZero(build.Commit),
		}
		if build.Number != nil {
			commit.BuildNumber = *build.Number
			commit.BuildURL = pointers.DerefZero(build.WebURL)
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

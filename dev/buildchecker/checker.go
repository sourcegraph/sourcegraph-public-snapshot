package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/buildkite/go-buildkite/v3/buildkite"
)

type CheckOptions struct {
	FailuresThreshold int
	BuildTimeout      time.Duration
}

type CommitInfo struct {
	Commit      string
	SlackUserID string
	Author      string
}

type CheckResults struct {
	// LockBranch indicates whether or not the Action will lock the branch.
	LockBranch bool
	// Action is a callback to actually execute changes.
	Action func() (err error)
	// FailedCommits lists the commits with failed builds that were detected.
	FailedCommits []CommitInfo
}

// CheckBuilds is the main buildchecker program. It checks the given builds for relevant
// failures and runs lock/unlock operations on the given branch.
func CheckBuilds(ctx context.Context, branch BranchLocker, slackUser SlackUserResolver, builds []buildkite.Build, opts CheckOptions) (results *CheckResults, err error) {
	results = &CheckResults{}

	// Scan for first build with a meaningful state
	var firstFailedBuildIndex int
	for i, b := range builds {
		if isBuildPassed(b) {
			fmt.Printf("most recent finished build %d passed\n", *b.Number)
			results.Action, err = branch.Unlock(ctx)
			if err != nil {
				return nil, fmt.Errorf("unlockBranch: %w", err)
			}
			return
		}
		if isBuildFailed(b, opts.BuildTimeout) {
			fmt.Printf("most recent finished build %d failed\n", *b.Number)
			firstFailedBuildIndex = i
			break
		}

		// Otherwise, keep looking for a completed (failed or passed) build
	}

	// if failed, check if failures are consecutive
	var exceeded bool
	results.FailedCommits, exceeded = checkConsecutiveFailures(
		builds[max(firstFailedBuildIndex-1, 0):], // Check builds starting with the one we found
		opts.FailuresThreshold,
		opts.BuildTimeout)
	if !exceeded {
		fmt.Println("threshold not exceeded")
		results.Action, err = branch.Unlock(ctx)
		if err != nil {
			return nil, fmt.Errorf("unlockBranch: %w", err)
		}
		return
	}

	fmt.Println("threshold exceeded, this is a big deal!")

	// annotate the failures with their author (Github handle), so we can reach them
	// over Slack.
	for i, info := range results.FailedCommits {
		results.FailedCommits[i].SlackUserID, err = slackUser.ResolveByCommit(ctx, info.Commit)
		if err != nil {
			// If we can't resolve the user, do not interrupt the process.
			fmt.Println(fmt.Errorf("slackUserResolve: %w", err))
		}
	}

	results.LockBranch = true
	results.Action, err = branch.Lock(ctx, results.FailedCommits, "dev-experience")
	if err != nil {
		return nil, fmt.Errorf("lockBranch: %w", err)
	}
	return
}

func isBuildPassed(build buildkite.Build) bool {
	return build.State != nil && *build.State == "passed"
}

func isBuildFailed(build buildkite.Build, timeout time.Duration) bool {
	// Has state and is failed
	if build.State != nil && (*build.State == "failed" || *build.State == "cancelled") {
		return true
	}
	// Created, but not done
	if build.CreatedAt != nil && build.FinishedAt == nil {
		// Failed if exceeded timeout
		return time.Now().After(build.CreatedAt.Add(timeout))
	}
	return false
}

func buildSummary(build buildkite.Build) string {
	summary := []string{*build.Commit}
	if build.State != nil {
		summary = append(summary, *build.State)
	}
	if build.CreatedAt != nil {
		summary = append(summary, "started: "+build.CreatedAt.String())
	}
	if build.FinishedAt != nil {
		summary = append(summary, "finished: "+build.FinishedAt.String())
	}
	return strings.Join(summary, ", ")
}

func checkConsecutiveFailures(builds []buildkite.Build, threshold int, timeout time.Duration) (failedCommits []CommitInfo, thresholdExceeded bool) {
	failedCommits = []CommitInfo{}

	var consecutiveFailures int
	for _, b := range builds {
		if !isBuildFailed(b, timeout) {
			fmt.Printf("build %d not failed: %+v\n", *b.Number, buildSummary(b))

			if !isBuildPassed(b) {
				// we're only safe if non-failures are actually passed, otherwise
				// keep looking
				continue
			}
			return
		}

		var author string
		if b.Author != nil {
			author = fmt.Sprintf("%s (%s)", b.Author.Name, b.Author.Email)
		}

		consecutiveFailures += 1
		failedCommits = append(failedCommits, CommitInfo{
			Commit: *b.Commit,
			Author: author,
		})
		fmt.Printf("build %d is a failure: count %d, %s\n", *b.Number, consecutiveFailures, buildSummary(b))
		if consecutiveFailures >= threshold {
			return failedCommits, true
		}
	}

	return
}

func max(x, y int) int {
	if x < y {
		return y
	}
	return x
}

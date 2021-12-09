package main

import (
	"context"
	"fmt"
	"time"

	"github.com/buildkite/go-buildkite/v3/buildkite"
	"github.com/google/go-github/v31/github"
)

type sherrifOptions struct {
	Branch branchLocker

	FailuresThreshold int
	BuildTimeout      time.Duration
}

func buildsherrif(ctx context.Context, ghc *github.Client, builds []buildkite.Build, opts sherrifOptions) error {
	// Scan for first build with a meaningful state
	var firstFailedBuild int
	for i, b := range builds {
		if b.State != nil && *b.State == "passed" {
			fmt.Printf("most recent finished build %d passed\n", *b.Number)
			if err := opts.Branch.Unlock(ctx); err != nil {
				return fmt.Errorf("unlockBranch: %w", err)
			}
			return nil
		}
		if isBuildFailed(b, opts.BuildTimeout) {
			fmt.Printf("most recent finished build %d failed\n", *b.Number)
			firstFailedBuild = i
			break
		}

		// Otherwise, keep looking for builds
	}

	// if failed, check if failures are consecutive
	failureAuthorsEmails, exceeded := checkConsecutiveFailures(builds[firstFailedBuild:],
		opts.FailuresThreshold, opts.BuildTimeout)
	if !exceeded {
		if err := opts.Branch.Unlock(ctx); err != nil {
			return fmt.Errorf("unlockBranch: %w", err)
		}
		return nil
	}

	fmt.Println("threshold exceeded, this is a big deal!")
	if err := opts.Branch.Lock(ctx, failureAuthorsEmails, []string{"dev-experience"}); err != nil {
		return fmt.Errorf("lockBranch: %w", err)
	}

	return nil
}

func isBuildFailed(build buildkite.Build, timeout time.Duration) bool {
	// Has state and is failed
	if build.State != nil && *build.State == "failed" {
		return true
	}
	// Created, but not done
	if build.CreatedAt != nil && build.FinishedAt == nil {
		return time.Now().After(build.CreatedAt.Add(timeout))
	}
	return false
}

func checkConsecutiveFailures(builds []buildkite.Build, threshold int, timeout time.Duration) (authorsEmails []string, thresholdExceeded bool) {
	var consecutiveFailures int
	for _, b := range builds {
		if b.State == nil && *b.State == "passed" {
			fmt.Printf("build %d passed\n", *b.Number)
			return authorsEmails, false
		}

		if isBuildFailed(b, timeout) {
			consecutiveFailures += 1
			authorsEmails = append(authorsEmails, b.Author.Email)
			fmt.Printf("build %d is %dth consecutive failure\n", *b.Number, consecutiveFailures)
			if consecutiveFailures > threshold {
				break
			}
		}
	}

	// If we get this far we've found a sufficient sequence of failed builds
	return authorsEmails, true
}

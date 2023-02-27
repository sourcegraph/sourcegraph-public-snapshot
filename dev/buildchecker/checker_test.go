package main

import (
	"context"
	"testing"
	"time"

	"github.com/buildkite/go-buildkite/v3/buildkite"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/dev/team"
)

type mockBranchLocker struct {
	calledUnlock int
	calledLock   int
}

func (m *mockBranchLocker) Unlock(context.Context) (func() error, error) {
	m.calledUnlock += 1
	return func() error { return nil }, nil
}
func (m *mockBranchLocker) Lock(context.Context, []CommitInfo, string) (func() error, error) {
	m.calledLock += 1
	return func() error { return nil }, nil
}

func TestCheckBuilds(t *testing.T) {
	// Simple end-to-end tests of the buildchecker entrypoint with mostly fixed parameters
	ctx := context.Background()
	slackUser := team.NewMockTeammateResolver()
	slackUser.ResolveByCommitAuthorFunc.SetDefaultReturn(&team.Teammate{SlackID: "bobheadxi"}, nil)
	testOptions := CheckOptions{
		FailuresThreshold: 2,
		BuildTimeout:      time.Hour,
	}

	// Triggers a pass
	passBuild := buildkite.Build{
		Number: buildkite.Int(1),
		Commit: buildkite.String("a"),
		State:  buildkite.String("passed"),
	}
	// Triggers a fail
	failSet := []buildkite.Build{{
		Number: buildkite.Int(2),
		Commit: buildkite.String("a"),
		State:  buildkite.String("failed"),
	}, {
		Number: buildkite.Int(3),
		Commit: buildkite.String("b"),
		State:  buildkite.String("failed"),
	}}
	runningBuild := buildkite.Build{
		Number:    buildkite.Int(4),
		Commit:    buildkite.String("a"),
		State:     buildkite.String("running"),
		StartedAt: buildkite.NewTimestamp(time.Now()),
	}
	scheduledBuild := buildkite.Build{
		Number:    buildkite.Int(5),
		Commit:    buildkite.String("a"),
		State:     buildkite.String("failed"),
		StartedAt: buildkite.NewTimestamp(time.Now()),
		Source:    buildkite.String("scheduled"),
	}

	tests := []struct {
		name       string
		builds     []buildkite.Build
		wantLocked bool
	}{{
		name:       "passed, should not lock",
		builds:     []buildkite.Build{passBuild},
		wantLocked: false,
	}, {
		name:       "not enough failed, should not lock",
		builds:     []buildkite.Build{failSet[0]},
		wantLocked: false,
	}, {
		name:       "should lock",
		builds:     failSet,
		wantLocked: true,
	}, {
		name:       "should skip leading running builds to pass",
		builds:     []buildkite.Build{runningBuild, passBuild},
		wantLocked: false,
	}, {
		name:       "should skip leading running builds to lock",
		builds:     append([]buildkite.Build{runningBuild}, failSet...),
		wantLocked: true,
	}, {
		name:       "should not locked because of scheduled build",
		builds:     []buildkite.Build{failSet[0], scheduledBuild},
		wantLocked: false,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var lock = &mockBranchLocker{}
			res, err := CheckBuilds(ctx, lock, slackUser, tt.builds, testOptions)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantLocked, res.LockBranch, "should lock")
			// Mock always returns an action, check it's always assigned correctly
			assert.NotNil(t, res.Action, "Action")
			// Lock/Unlock should not be called repeatedly
			assert.LessOrEqual(t, lock.calledUnlock, 1, "calledUnlock")
			assert.LessOrEqual(t, lock.calledLock, 1, "calledLock")
			// Don't return >N failed commits
			assert.LessOrEqual(t, len(res.FailedCommits), testOptions.FailuresThreshold, "FailedCommits count")
		})
	}

	t.Run("only return oldest N failed commits", func(t *testing.T) {
		var lock = &mockBranchLocker{}
		res, err := CheckBuilds(ctx, lock, slackUser, append(failSet,
			// 2 builds == FailuresThreshold
			buildkite.Build{
				Number: buildkite.Int(10),
				Commit: buildkite.String("b"),
				State:  buildkite.String("failed"),
			}, buildkite.Build{
				Number: buildkite.Int(20),
				Commit: buildkite.String("b"),
				State:  buildkite.String("failed"),
			}),
			testOptions)
		assert.NoError(t, err)
		assert.True(t, res.LockBranch, "should lock")

		assert.Len(t, res.FailedCommits, testOptions.FailuresThreshold, "FailedCommits count")
		gotBuildNumbers := []int{}
		for _, c := range res.FailedCommits {
			gotBuildNumbers = append(gotBuildNumbers, c.BuildNumber)
		}
		assert.Equal(t, []int{10, 20}, gotBuildNumbers, "FailedCommits build numbers")
	})
}

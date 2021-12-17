package main

import (
	"context"
	"testing"
	"time"

	"github.com/buildkite/go-buildkite/v3/buildkite"
	"github.com/stretchr/testify/assert"
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
		Number: buildkite.Int(1),
		Commit: buildkite.String("a"),
		State:  buildkite.String("failed"),
	}, {
		Number: buildkite.Int(2),
		Commit: buildkite.String("b"),
		State:  buildkite.String("failed"),
	}}
	runningBuild := buildkite.Build{
		Number:    buildkite.Int(1),
		Commit:    buildkite.String("a"),
		State:     buildkite.String("running"),
		StartedAt: buildkite.NewTimestamp(time.Now()),
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
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var lock = &mockBranchLocker{}
			res, err := CheckBuilds(ctx, lock, tt.builds, testOptions)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantLocked, res.LockBranch)
			// Mock always returns an action, check it's always assigned correctly
			assert.NotNil(t, res.Action)
			// Lock/Unlock should not be called repeatedly
			assert.LessOrEqual(t, lock.calledUnlock, 1, "calledUnlock")
			assert.LessOrEqual(t, lock.calledLock, 1, "calledLock")
		})
	}
}

func TestCheckConsecutiveFailures(t *testing.T) {
	type args struct {
		builds    []buildkite.Build
		threshold int
		timeout   time.Duration
	}
	tests := []struct {
		name                  string
		args                  args
		wantCommits           []string
		wantThresholdExceeded bool
	}{{
		name: "not exceeded: passed",
		args: args{
			builds: []buildkite.Build{{
				Number: buildkite.Int(1),
				Commit: buildkite.String("a"),
				State:  buildkite.String("passed"),
			}},
			threshold: 3, timeout: time.Hour,
		},
		wantCommits:           []string{},
		wantThresholdExceeded: false,
	}, {
		name: "not exceeded: failed",
		args: args{
			builds: []buildkite.Build{{
				Number: buildkite.Int(1),
				Commit: buildkite.String("a"),
				State:  buildkite.String("failed"),
			}},
			threshold: 3, timeout: time.Hour,
		},
		wantCommits:           []string{"a"},
		wantThresholdExceeded: false,
	}, {
		name: "not exceeded: failed, passed",
		args: args{
			builds: []buildkite.Build{{
				Number: buildkite.Int(1),
				Commit: buildkite.String("a"),
				State:  buildkite.String("failed"),
			}, {
				Number: buildkite.Int(2),
				Commit: buildkite.String("b"),
				State:  buildkite.String("passed"),
			}},
			threshold: 3, timeout: time.Hour,
		},
		wantCommits:           []string{"a"},
		wantThresholdExceeded: false,
	}, {
		name: "not exceeded: failed, passed, failed",
		args: args{
			builds: []buildkite.Build{{
				Number: buildkite.Int(1),
				Commit: buildkite.String("a"),
				State:  buildkite.String("failed"),
			}, {
				Number: buildkite.Int(2),
				Commit: buildkite.String("b"),
				State:  buildkite.String("passed"),
			}, {
				Number: buildkite.Int(3),
				Commit: buildkite.String("c"),
				State:  buildkite.String("failed"),
			}},
			threshold: 2, timeout: time.Hour,
		},
		wantCommits:           []string{"a"},
		wantThresholdExceeded: false,
	}, {
		name: "exceeded: failed == threshold",
		args: args{
			builds: []buildkite.Build{{
				Number: buildkite.Int(1),
				Commit: buildkite.String("a"),
				State:  buildkite.String("failed"),
			}},
			threshold: 1, timeout: time.Hour,
		},
		wantCommits:           []string{"a"},
		wantThresholdExceeded: true,
	}, {
		name: "exceeded: failed == threshold",
		args: args{
			builds: []buildkite.Build{{
				Number: buildkite.Int(1),
				Commit: buildkite.String("a"),
				State:  buildkite.String("failed"),
			}},
			threshold: 1, timeout: time.Hour,
		},
		wantCommits:           []string{"a"},
		wantThresholdExceeded: true,
	}, {
		name: "exceeded: failed, timeout, failed",
		args: args{
			builds: []buildkite.Build{{
				Number: buildkite.Int(1),
				Commit: buildkite.String("a"),
				State:  buildkite.String("failed"),
			}, {
				Number:    buildkite.Int(2),
				Commit:    buildkite.String("b"),
				State:     buildkite.String("running"),
				CreatedAt: buildkite.NewTimestamp(time.Now().Add(-2 * time.Hour)),
			}, {
				Number: buildkite.Int(3),
				Commit: buildkite.String("c"),
				State:  buildkite.String("failed"),
			}},
			threshold: 3, timeout: time.Hour,
		},
		wantCommits:           []string{"a", "b", "c"},
		wantThresholdExceeded: true,
	}, {
		name: "exceeded: failed, running, failed",
		args: args{
			builds: []buildkite.Build{{
				Number: buildkite.Int(1),
				Commit: buildkite.String("a"),
				State:  buildkite.String("failed"),
			}, {
				Number:    buildkite.Int(2),
				Commit:    buildkite.String("b"),
				State:     buildkite.String("running"),
				CreatedAt: buildkite.NewTimestamp(time.Now()),
			}, {
				Number: buildkite.Int(3),
				Commit: buildkite.String("c"),
				State:  buildkite.String("failed"),
			}},
			threshold: 2, timeout: time.Hour,
		},
		wantCommits:           []string{"a", "c"},
		wantThresholdExceeded: true,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCommits, gotThresholdExceeded := checkConsecutiveFailures(tt.args.builds, tt.args.threshold, tt.args.timeout)
			assert.Equal(t, tt.wantThresholdExceeded, gotThresholdExceeded, "thresholdExceeded")

			wantCommits := []CommitInfo{}
			for _, c := range tt.wantCommits {
				wantCommits = append(wantCommits, CommitInfo{Commit: c})
			}
			assert.Equal(t, wantCommits, gotCommits, "commits")
		})
	}
}

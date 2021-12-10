package main

import (
	"testing"
	"time"

	"github.com/buildkite/go-buildkite/v3/buildkite"
	"github.com/stretchr/testify/assert"
)

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
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCommits, gotThresholdExceeded := checkConsecutiveFailures(tt.args.builds, tt.args.threshold, tt.args.timeout)
			assert.Equal(t, tt.wantCommits, gotCommits, "commits")
			assert.Equal(t, tt.wantThresholdExceeded, gotThresholdExceeded, "thresholdExceeded")
		})
	}
}

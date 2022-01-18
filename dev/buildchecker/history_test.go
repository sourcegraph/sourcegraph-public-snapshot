package main

import (
	"testing"
	"time"

	"github.com/buildkite/go-buildkite/v3/buildkite"
	"github.com/stretchr/testify/assert"
)

func TestGenerateHistory(t *testing.T) {
	day := time.Date(2006, 01, 02, 0, 0, 0, 0, time.UTC)
	dayString := day.Format("2006/01/02")

	tests := []struct {
		name                    string
		builds                  []buildkite.Build
		wantFlakes              map[string]int
		wantConsecutiveFailures map[string]int
	}{{
		name: "consecutive failures",
		builds: []buildkite.Build{{
			CreatedAt: &buildkite.Timestamp{Time: day.Add(2 * time.Hour)},
			State:     buildkite.String("failed"),
		}, {
			CreatedAt: &buildkite.Timestamp{Time: day.Add(1 * time.Hour)},
			State:     buildkite.String("failed"),
		}, {
			CreatedAt: &buildkite.Timestamp{Time: day.Add(5 * time.Minute)},
			State:     buildkite.String("failed"),
		}, {
			CreatedAt: &buildkite.Timestamp{Time: day},
			State:     buildkite.String("failed"),
		}},
		wantFlakes: map[string]int{},
		wantConsecutiveFailures: map[string]int{
			dayString: 60 * 2,
		},
	}, {
		name: "passed, then consecutive failures",
		builds: []buildkite.Build{{
			CreatedAt: &buildkite.Timestamp{Time: day.Add(2 * time.Hour)},
			State:     buildkite.String("passed"),
		}, {
			CreatedAt: &buildkite.Timestamp{Time: day.Add(1 * time.Hour)},
			State:     buildkite.String("failed"),
		}, {
			CreatedAt: &buildkite.Timestamp{Time: day.Add(30 * time.Minute)},
			State:     buildkite.String("failed"),
		}, {
			CreatedAt: &buildkite.Timestamp{Time: day},
			State:     buildkite.String("failed"),
		}},
		wantFlakes: map[string]int{},
		wantConsecutiveFailures: map[string]int{
			dayString: 60 * 2,
		},
	}, {
		name: "mixed flakes",
		builds: []buildkite.Build{{
			CreatedAt: &buildkite.Timestamp{Time: day.Add(2 * time.Hour)},
			State:     buildkite.String("passed"),
		}, {
			CreatedAt: &buildkite.Timestamp{Time: day.Add(1 * time.Hour)},
			State:     buildkite.String("failed"),
		}, {
			CreatedAt: &buildkite.Timestamp{Time: day.Add(30 * time.Minute)},
			State:     buildkite.String("passed"),
		}, {
			CreatedAt: &buildkite.Timestamp{Time: day.Add(5 * time.Minute)},
			State:     buildkite.String("failed"),
		}, {
			CreatedAt: &buildkite.Timestamp{Time: day},
			State:     buildkite.String("failed"),
		}},
		wantFlakes: map[string]int{
			dayString: 3,
		},
		wantConsecutiveFailures: map[string]int{},
	}, {
		name: "flake -> consecutive",
		builds: []buildkite.Build{{
			CreatedAt: &buildkite.Timestamp{Time: day.Add(2 * time.Hour)},
			State:     buildkite.String("failed"),
		}, {
			CreatedAt: &buildkite.Timestamp{Time: day.Add(1 * time.Hour)},
			State:     buildkite.String("passed"),
		}, {
			CreatedAt: &buildkite.Timestamp{Time: day.Add(30 * time.Minute)},
			State:     buildkite.String("failed"),
		}, {
			CreatedAt: &buildkite.Timestamp{Time: day.Add(5 * time.Minute)},
			State:     buildkite.String("failed"),
		}, {
			CreatedAt: &buildkite.Timestamp{Time: day},
			State:     buildkite.String("failed"),
		}},
		wantFlakes: map[string]int{
			dayString: 1,
		},
		wantConsecutiveFailures: map[string]int{
			dayString: 60,
		},
	}, {
		name: "consecutive -> flake",
		builds: []buildkite.Build{{
			CreatedAt: &buildkite.Timestamp{Time: day.Add(2 * time.Hour)},
			State:     buildkite.String("failed"),
		}, {
			CreatedAt: &buildkite.Timestamp{Time: day.Add(1 * time.Hour)},
			State:     buildkite.String("failed"),
		}, {
			CreatedAt: &buildkite.Timestamp{Time: day.Add(30 * time.Minute)},
			State:     buildkite.String("failed"),
		}, {
			CreatedAt: &buildkite.Timestamp{Time: day.Add(5 * time.Minute)},
			State:     buildkite.String("passed"),
		}, {
			CreatedAt: &buildkite.Timestamp{Time: day},
			State:     buildkite.String("failed"),
		}},
		wantFlakes: map[string]int{
			dayString: 1,
		},
		wantConsecutiveFailures: map[string]int{
			dayString: 90,
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			println("--- " + t.Name())
			_, gotFlakes, gotConsecutiveFailures := generateHistory(tt.builds, day.Add(2*time.Hour), CheckOptions{
				FailuresThreshold: 3,
				BuildTimeout:      0, // disable timeout check
			})
			assert.Equal(t, gotFlakes, tt.wantFlakes, "flakes")
			assert.Equal(t, gotConsecutiveFailures, tt.wantConsecutiveFailures, "consecutive failures")
		})
	}
}

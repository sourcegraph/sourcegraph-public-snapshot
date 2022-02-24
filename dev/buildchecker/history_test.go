package main

import (
	"testing"
	"time"

	"github.com/buildkite/go-buildkite/v3/buildkite"
	"github.com/stretchr/testify/assert"
)

func TestGenerateHistory(t *testing.T) {
	day := time.Date(2006, 01, 02, 0, 0, 0, 0, time.UTC)
	dayString := day.Format("2006-01-02")

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
			_, gotFlakes, gotConsecutiveFailures := generateHistory(tt.builds, day.Add(2*time.Hour), CheckOptions{
				FailuresThreshold: 3,
				BuildTimeout:      0, // disable timeout check
			})
			assert.Equal(t, gotFlakes, tt.wantFlakes, "flakes")
			assert.Equal(t, gotConsecutiveFailures, tt.wantConsecutiveFailures, "consecutive failures")
		})
	}
}

func TestMapToRecords(t *testing.T) {
	tests := []struct {
		name        string
		arg         map[string]int
		wantRecords [][]string
	}{{
		name: "sorted",
		arg: map[string]int{
			"2022-01-02": 2,
			"2022-01-01": 1,
			"2022-01-03": 3,
		},
		wantRecords: [][]string{
			{"2022-01-01", "1"},
			{"2022-01-02", "2"},
			{"2022-01-03", "3"},
		},
	}, {
		name: "gaps filled in",
		arg: map[string]int{
			"2022-01-01": 1,
			"2022-01-03": 3,
			"2022-01-06": 6,
			"2022-01-07": 7,
		},
		wantRecords: [][]string{
			{"2022-01-01", "1"},
			{"2022-01-02", "0"},
			{"2022-01-03", "3"},
			{"2022-01-04", "0"},
			{"2022-01-05", "0"},
			{"2022-01-06", "6"},
			{"2022-01-07", "7"},
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRecords := mapToRecords(tt.arg)
			assert.Equal(t, tt.wantRecords, gotRecords)
		})
	}
}

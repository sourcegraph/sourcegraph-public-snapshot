package ratelimit

import (
	"testing"
	"time"
)

func TestLimit(t *testing.T) {
	baseTime := time.Now()

	secondsAgo := func(t time.Duration) time.Time {
		return baseTime.Add(-t * time.Second)
	}

	testLimit := limit{
		maxActions:          12,
		window:              1 * time.Minute,
		uniquePenaltyFactor: 2,
		maxUniquePenalty:    10,
	}
	tests := []struct {
		name    string
		actions []action
		want    time.Duration
	}{
		{
			name: "under_limit_by_one",
			actions: []action{
				{secondsAgo(1), "a.go"},
				{secondsAgo(1), "a.go"},
				{secondsAgo(1), "a.go"},
				{secondsAgo(2), "a.go"},
				{secondsAgo(2), "a.go"},
				{secondsAgo(3), "a.go"},
				{secondsAgo(3), "a.go"},
				{secondsAgo(3), "a.go"},
				{secondsAgo(4), "a.go"},
				{secondsAgo(4), "a.go"},
				{secondsAgo(5), "a.go"},
			},
			want: 0 * time.Second,
		},
		{
			name: "limit_hit",
			actions: []action{
				{secondsAgo(1), "a.go"},
				{secondsAgo(1), "a.go"},
				{secondsAgo(1), "a.go"},
				{secondsAgo(2), "a.go"},
				{secondsAgo(2), "a.go"},
				{secondsAgo(3), "a.go"},
				{secondsAgo(3), "a.go"},
				{secondsAgo(3), "a.go"},
				{secondsAgo(4), "a.go"},
				{secondsAgo(4), "a.go"},
				{secondsAgo(5), "a.go"},
				{secondsAgo(6), "a.go"},
			},
			want: 54 * time.Second,
		},
		{
			name: "2_unique_files",
			actions: []action{
				{secondsAgo(1), "a.go"},
				{secondsAgo(1), "b.go"},
				{secondsAgo(1), "b.go"},
				{secondsAgo(2), "b.go"},
				{secondsAgo(2), "b.go"},
				{secondsAgo(13), "b.go"},
			},
			want: 47 * time.Second,
		},
		{
			name: "3_unique_files",
			actions: []action{
				{secondsAgo(1), "a.go"},
				{secondsAgo(2), "b.go"},
				{secondsAgo(3), "c.go"},
			},
			want: 57 * time.Second,
		},
		{
			name: "4_unique_files",
			actions: []action{
				{secondsAgo(1), "a.go"},
				{secondsAgo(1), "b.go"},
				{secondsAgo(1), "c.go"},
				{secondsAgo(2), "d.go"},
				{secondsAgo(58), "d.go"},
				{secondsAgo(2), "d.go"},
			},
			want: 2 * time.Second,
		},
	}
	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			got := testLimit.calculateWaitTime(tst.actions).Round(time.Second)
			if got != tst.want {
				t.Fatalf("got %v want %v", got, tst.want)
			}
		})
	}
}

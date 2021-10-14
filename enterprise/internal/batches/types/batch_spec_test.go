package types

import (
	"testing"
)

func TestComputeBatchSpecState(t *testing.T) {
	tests := []struct {
		stats BatchSpecStats
		want  string
	}{
		{
			stats: BatchSpecStats{Workspaces: 5},
			want:  "PENDING",
		},
		{
			stats: BatchSpecStats{Workspaces: 5, Executions: 3, Queued: 3},
			want:  "QUEUED",
		},
		{
			stats: BatchSpecStats{Workspaces: 5, Executions: 3, Queued: 2, Processing: 1},
			want:  "PROCESSING",
		},
		{
			stats: BatchSpecStats{Workspaces: 5, Executions: 3, Queued: 1, Processing: 1, Completed: 1},
			want:  "PROCESSING",
		},
		{
			stats: BatchSpecStats{Workspaces: 5, Executions: 3, Queued: 1, Processing: 0, Completed: 2},
			want:  "PROCESSING",
		},
		{
			stats: BatchSpecStats{Workspaces: 5, Executions: 3, Queued: 0, Processing: 0, Completed: 3},
			want:  "COMPLETED",
		},
		{
			stats: BatchSpecStats{Workspaces: 5, Executions: 3, Queued: 1, Processing: 1, Failed: 1},
			want:  "PROCESSING",
		},
		{
			stats: BatchSpecStats{Workspaces: 5, Executions: 3, Queued: 1, Processing: 0, Failed: 2},
			want:  "PROCESSING",
		},
		{
			stats: BatchSpecStats{Workspaces: 5, Executions: 3, Queued: 0, Processing: 0, Failed: 3},
			want:  "FAILED",
		},
		{
			stats: BatchSpecStats{Workspaces: 5, Executions: 3, Queued: 0, Completed: 1, Failed: 2},
			want:  "FAILED",
		},
		{
			stats: BatchSpecStats{Workspaces: 5, Executions: 3, Canceling: 3},
			want:  "CANCELING",
		},
		{
			stats: BatchSpecStats{Workspaces: 5, Executions: 3, Canceling: 2, Completed: 1},
			want:  "CANCELING",
		},
		{
			stats: BatchSpecStats{Workspaces: 5, Executions: 3, Canceling: 2, Failed: 1},
			want:  "CANCELING",
		},
		{
			stats: BatchSpecStats{Workspaces: 5, Executions: 3, Canceling: 1, Queued: 2},
			want:  "PROCESSING",
		},
		{
			stats: BatchSpecStats{Workspaces: 5, Executions: 3, Canceling: 1, Processing: 2},
			want:  "PROCESSING",
		},
		{
			stats: BatchSpecStats{Workspaces: 5, Executions: 3, Canceled: 3},
			want:  "CANCELED",
		},
		{
			stats: BatchSpecStats{Workspaces: 5, Executions: 3, Canceled: 1, Failed: 2},
			want:  "CANCELED",
		},
		{
			stats: BatchSpecStats{Workspaces: 5, Executions: 3, Canceled: 1, Completed: 2},
			want:  "CANCELED",
		},
		{
			stats: BatchSpecStats{Workspaces: 5, Executions: 3, Canceled: 1, Canceling: 2},
			want:  "CANCELING",
		},
		{
			stats: BatchSpecStats{Workspaces: 5, Executions: 3, Canceled: 1, Canceling: 1, Queued: 1},
			want:  "PROCESSING",
		},
		{
			stats: BatchSpecStats{Workspaces: 5, Executions: 3, Canceled: 1, Processing: 2},
			want:  "PROCESSING",
		},
		{
			stats: BatchSpecStats{Workspaces: 5, Executions: 3, Canceled: 1, Canceling: 1, Processing: 1},
			want:  "PROCESSING",
		},
		{
			stats: BatchSpecStats{Workspaces: 5, Executions: 3, Canceled: 1, Queued: 2},
			want:  "PROCESSING",
		},
	}

	for idx, tt := range tests {
		have := ComputeBatchSpecState(tt.stats)

		if have != tt.want {
			t.Errorf("test %d/%d: unexpected batch spec state. want=%s, have=%s", idx+1, len(tests), tt.want, have)
		}
	}
}

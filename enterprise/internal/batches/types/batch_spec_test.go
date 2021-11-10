package types

import (
	"testing"
)

func TestComputeBatchSpecState(t *testing.T) {
	uploadedSpec := &BatchSpec{CreatedFromRaw: false}
	createdFromRawSpec := &BatchSpec{CreatedFromRaw: true}

	tests := []struct {
		stats BatchSpecStats
		spec  *BatchSpec
		want  BatchSpecState
	}{
		{
			stats: BatchSpecStats{ResolutionDone: false},
			spec:  uploadedSpec,
			want:  BatchSpecStateCompleted,
		},
		{
			stats: BatchSpecStats{ResolutionDone: false},
			spec:  createdFromRawSpec,
			want:  BatchSpecStatePending,
		},
		{
			stats: BatchSpecStats{ResolutionDone: true, Workspaces: 5},
			spec:  createdFromRawSpec,
			want:  BatchSpecStatePending,
		},
		{
			stats: BatchSpecStats{ResolutionDone: true, Workspaces: 5, SkippedWorkspaces: 5},
			spec:  createdFromRawSpec,
			want:  BatchSpecStateCompleted,
		},
		{
			stats: BatchSpecStats{ResolutionDone: true, Workspaces: 5, SkippedWorkspaces: 3},
			spec:  createdFromRawSpec,
			want:  BatchSpecStatePending,
		},
		{
			stats: BatchSpecStats{ResolutionDone: true, Workspaces: 5, SkippedWorkspaces: 2, Executions: 3, Queued: 3},
			spec:  createdFromRawSpec,
			want:  BatchSpecStateQueued,
		},
		{
			stats: BatchSpecStats{ResolutionDone: true, Workspaces: 5, Executions: 3, Queued: 3},
			spec:  createdFromRawSpec,
			want:  BatchSpecStateQueued,
		},
		{
			stats: BatchSpecStats{ResolutionDone: true, Workspaces: 5, Executions: 3, Queued: 2, Processing: 1},
			spec:  createdFromRawSpec,
			want:  BatchSpecStateProcessing,
		},
		{
			stats: BatchSpecStats{ResolutionDone: true, Workspaces: 5, Executions: 3, Queued: 1, Processing: 1, Completed: 1},
			spec:  createdFromRawSpec,
			want:  BatchSpecStateProcessing,
		},
		{
			stats: BatchSpecStats{ResolutionDone: true, Workspaces: 5, Executions: 3, Queued: 1, Processing: 0, Completed: 2},
			spec:  createdFromRawSpec,
			want:  BatchSpecStateProcessing,
		},
		{
			stats: BatchSpecStats{ResolutionDone: true, Workspaces: 5, Executions: 3, Queued: 0, Processing: 0, Completed: 3},
			spec:  createdFromRawSpec,
			want:  BatchSpecStateCompleted,
		},
		{
			stats: BatchSpecStats{ResolutionDone: true, Workspaces: 5, Executions: 3, Queued: 1, Processing: 1, Failed: 1},
			spec:  createdFromRawSpec,
			want:  BatchSpecStateProcessing,
		},
		{
			stats: BatchSpecStats{ResolutionDone: true, Workspaces: 5, Executions: 3, Queued: 1, Processing: 0, Failed: 2},
			spec:  createdFromRawSpec,
			want:  BatchSpecStateProcessing,
		},
		{
			stats: BatchSpecStats{ResolutionDone: true, Workspaces: 5, Executions: 3, Queued: 0, Processing: 0, Failed: 3},
			spec:  createdFromRawSpec,
			want:  BatchSpecStateFailed,
		},
		{
			stats: BatchSpecStats{ResolutionDone: true, Workspaces: 5, Executions: 3, Queued: 0, Completed: 1, Failed: 2},
			spec:  createdFromRawSpec,
			want:  BatchSpecStateFailed,
		},
		{
			stats: BatchSpecStats{ResolutionDone: true, Workspaces: 5, Executions: 3, Canceling: 3},
			spec:  createdFromRawSpec,
			want:  BatchSpecStateCanceling,
		},
		{
			stats: BatchSpecStats{ResolutionDone: true, Workspaces: 5, Executions: 3, Canceling: 2, Completed: 1},
			spec:  createdFromRawSpec,
			want:  BatchSpecStateCanceling,
		},
		{
			stats: BatchSpecStats{ResolutionDone: true, Workspaces: 5, Executions: 3, Canceling: 2, Failed: 1},
			spec:  createdFromRawSpec,
			want:  BatchSpecStateCanceling,
		},
		{
			stats: BatchSpecStats{ResolutionDone: true, Workspaces: 5, Executions: 3, Canceling: 1, Queued: 2},
			spec:  createdFromRawSpec,
			want:  BatchSpecStateProcessing,
		},
		{
			stats: BatchSpecStats{ResolutionDone: true, Workspaces: 5, Executions: 3, Canceling: 1, Processing: 2},
			spec:  createdFromRawSpec,
			want:  BatchSpecStateProcessing,
		},
		{
			stats: BatchSpecStats{ResolutionDone: true, Workspaces: 5, Executions: 3, Canceled: 3},
			spec:  createdFromRawSpec,
			want:  BatchSpecStateCanceled,
		},
		{
			stats: BatchSpecStats{ResolutionDone: true, Workspaces: 5, Executions: 3, Canceled: 1, Failed: 2},
			spec:  createdFromRawSpec,
			want:  BatchSpecStateCanceled,
		},
		{
			stats: BatchSpecStats{ResolutionDone: true, Workspaces: 5, Executions: 3, Canceled: 1, Completed: 2},
			spec:  createdFromRawSpec,
			want:  BatchSpecStateCanceled,
		},
		{
			stats: BatchSpecStats{ResolutionDone: true, Workspaces: 5, Executions: 3, Canceled: 1, Canceling: 2},
			spec:  createdFromRawSpec,
			want:  BatchSpecStateCanceling,
		},
		{
			stats: BatchSpecStats{ResolutionDone: true, Workspaces: 5, Executions: 3, Canceled: 1, Canceling: 1, Queued: 1},
			spec:  createdFromRawSpec,
			want:  BatchSpecStateProcessing,
		},
		{
			stats: BatchSpecStats{ResolutionDone: true, Workspaces: 5, Executions: 3, Canceled: 1, Processing: 2},
			spec:  createdFromRawSpec,
			want:  BatchSpecStateProcessing,
		},
		{
			stats: BatchSpecStats{ResolutionDone: true, Workspaces: 5, Executions: 3, Canceled: 1, Canceling: 1, Processing: 1},
			spec:  createdFromRawSpec,
			want:  BatchSpecStateProcessing,
		},
		{
			stats: BatchSpecStats{ResolutionDone: true, Workspaces: 5, Executions: 3, Canceled: 1, Queued: 2},
			spec:  createdFromRawSpec,
			want:  BatchSpecStateProcessing,
		},
		{
			stats: BatchSpecStats{ResolutionDone: true, Workspaces: 0, Executions: 0},
			spec:  createdFromRawSpec,
			want:  BatchSpecStateCompleted,
		},
	}

	for idx, tt := range tests {
		have := ComputeBatchSpecState(tt.spec, tt.stats)

		if have != tt.want {
			t.Errorf("test %d/%d: unexpected batch spec state. want=%s, have=%s", idx+1, len(tests), tt.want, have)
		}
	}
}

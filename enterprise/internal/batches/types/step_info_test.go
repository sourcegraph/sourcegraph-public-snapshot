package types

import (
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/executor"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func TestParseJSONLogsFromOutput(t *testing.T) {
	t.Parallel()

	now := timeutil.Now().Truncate(time.Second)
	event1 := &batcheslib.LogEvent{
		Operation: batcheslib.LogEventOperationExecutingTasks,
		Status:    batcheslib.LogEventStatusStarted,
		Metadata:  &batcheslib.ExecutingTasksMetadata{Tasks: []batcheslib.JSONLinesTask{}},
		Timestamp: now.Add(-5 * 60 * time.Minute),
	}
	event2 := &batcheslib.LogEvent{
		Operation: batcheslib.LogEventOperationExecutingTasks,
		Status:    batcheslib.LogEventStatusSuccess,
		Metadata:  &batcheslib.ExecutingTasksMetadata{},
		Timestamp: now.Add(-4 * 60 * time.Minute),
	}
	event3 := &batcheslib.LogEvent{
		Operation: batcheslib.LogEventOperationUploadingChangesetSpecs,
		Status:    batcheslib.LogEventStatusStarted,
		Metadata:  &batcheslib.UploadingChangesetSpecsMetadata{Total: 1},
		Timestamp: now.Add(-3 * 60 * time.Minute),
	}
	event4 := &batcheslib.LogEvent{
		Operation: batcheslib.LogEventOperationUploadingChangesetSpecs,
		Status:    batcheslib.LogEventStatusProgress,
		Metadata:  &batcheslib.UploadingChangesetSpecsMetadata{Total: 1, Done: 1},
		Timestamp: now.Add(-2 * 60 * time.Minute),
	}
	event5 := &batcheslib.LogEvent{
		Operation: batcheslib.LogEventOperationUploadingChangesetSpecs,
		Status:    batcheslib.LogEventStatusSuccess,
		Metadata:  &batcheslib.UploadingChangesetSpecsMetadata{IDs: []string{"Q2hhbmdlc2V0U3BlYzoiNkxIYWN5dkI3WDYi"}},
		Timestamp: now.Add(-1 * 60 * time.Minute),
	}

	tests := []struct {
		name       string
		output     []string
		wantEvents []*batcheslib.LogEvent
	}{
		{
			name: "success",
			output: []string{
				`stdout: {"operation":"EXECUTING_TASKS","timestamp":"` + event1.Timestamp.Format(time.RFC3339) + `","status":"STARTED","metadata":{"tasks":[]}}`,
				`stdout: {"operation":"EXECUTING_TASKS","timestamp":"` + event2.Timestamp.Format(time.RFC3339) + `","status":"SUCCESS"}`,
				`stdout: {"operation":"UPLOADING_CHANGESET_SPECS","timestamp":"` + event3.Timestamp.Format(time.RFC3339) + `","status":"STARTED","metadata":{"total":1}}`,
				`stdout: {"operation":"UPLOADING_CHANGESET_SPECS","timestamp":"` + event4.Timestamp.Format(time.RFC3339) + `","status":"PROGRESS","metadata":{"done":1,"total":1}}`,
				`stdout: {"operation":"UPLOADING_CHANGESET_SPECS","timestamp":"` + event5.Timestamp.Format(time.RFC3339) + `","status":"SUCCESS","metadata":{"ids":["Q2hhbmdlc2V0U3BlYzoiNkxIYWN5dkI3WDYi"]}}`,
			},
			wantEvents: []*batcheslib.LogEvent{
				event1,
				event2,
				event3,
				event4,
				event5,
			},
		},
		{
			name:       "no log lines",
			wantEvents: []*batcheslib.LogEvent{},
			output:     []string{},
		},
		{
			name: "with stderr messages",
			output: []string{
				`stdout: {"operation":"EXECUTING_TASKS","timestamp":"` + event1.Timestamp.Format(time.RFC3339) + `","status":"STARTED","metadata":{"tasks":[]}}`,
				`stdout: {"operation":"EXECUTING_TASKS","timestamp":"` + event2.Timestamp.Format(time.RFC3339) + `","status":"SUCCESS"}`,
				`stderr: HORSE`,
				`stdout: {"operation":"UPLOADING_CHANGESET_SPECS","timestamp":"` + event3.Timestamp.Format(time.RFC3339) + `","status":"STARTED","metadata":{"total":1}}`,
				`stderr: HORSE`,
				`stdout: {"operation":"UPLOADING_CHANGESET_SPECS","timestamp":"` + event4.Timestamp.Format(time.RFC3339) + `","status":"PROGRESS","metadata":{"done":1,"total":1}}`,
				`stderr: HORSE`,
				`stdout: {"operation":"UPLOADING_CHANGESET_SPECS","timestamp":"` + event5.Timestamp.Format(time.RFC3339) + `","status":"SUCCESS","metadata":{"ids":["Q2hhbmdlc2V0U3BlYzoiNkxIYWN5dkI3WDYi"]}}`,
			},
			wantEvents: []*batcheslib.LogEvent{
				event1,
				event2,
				event3,
				event4,
				event5,
			},
		},
		{
			name: "invalid json in between",
			output: []string{
				`stdout: {"operation":"EXECUTING_TASKS","timestamp":"` + event1.Timestamp.Format(time.RFC3339) + `","status":"STARTED","metadata":{"tasks":[]}}`,
				`stdout: {HOOOORSE}`,
				`stdout: {HORSE}`,
				`stdout: {HORSE}`,
			},
			wantEvents: []*batcheslib.LogEvent{
				event1,
			},
		},
		{
			name: "non-json output inbetween valid json",
			output: []string{
				`stdout: {"operation":"EXECUTING_TASKS","timestamp":"` + event1.Timestamp.Format(time.RFC3339) + `","status":"STARTED","metadata":{"tasks":[]}}`,
				`stdout: No changeset specs created`,
				`stdout: {"operation":"EXECUTING_TASKS","timestamp":"` + event2.Timestamp.Format(time.RFC3339) + `","status":"SUCCESS"}`,
			},
			wantEvents: []*batcheslib.LogEvent{
				event1,
				event2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			have := ParseJSONLogsFromOutput(strings.Join(tt.output, "\n"))
			if diff := cmp.Diff(tt.wantEvents, have); diff != "" {
				t.Fatalf("wrong IDs extracted: %s", diff)
			}
		})
	}
}

func TestParseLogLines(t *testing.T) {
	t.Parallel()

	time1 := timeutil.Now().Add(-15 * 60 * time.Second)
	time2 := timeutil.Now().Add(-10 * 60 * time.Second)
	time3 := timeutil.Now().Add(-5 * 60 * time.Second)
	zero := 0
	nonZero := 1
	diff := []byte(`all-must-change`)

	tcs := []struct {
		name  string
		entry executor.ExecutionLogEntry
		lines []*batcheslib.LogEvent
		want  map[int]*StepInfo
	}{
		{
			name: "Skipped",
			lines: []*batcheslib.LogEvent{
				{Metadata: &batcheslib.TaskStepSkippedMetadata{
					Step: 1,
				}},
			},
			want: map[int]*StepInfo{1: {Skipped: true}},
		},
		{
			name: "Multiple skipped",
			lines: []*batcheslib.LogEvent{
				{Metadata: &batcheslib.TaskSkippingStepsMetadata{
					// TODO: Verify how this works.
					StartStep: 3,
				}},
			},
			want: map[int]*StepInfo{
				1: {Skipped: true},
				2: {Skipped: true},
			},
		},
		{
			name: "Started preparation",
			lines: []*batcheslib.LogEvent{
				{
					Timestamp: time1,
					Status:    batcheslib.LogEventStatusStarted,
					Metadata: &batcheslib.TaskPreparingStepMetadata{
						Step: 1,
					},
				},
			},
			want: map[int]*StepInfo{
				1: {StartedAt: time1},
			},
		},
		{
			name: "Started with env",
			lines: []*batcheslib.LogEvent{
				{
					Timestamp: time1,
					Status:    batcheslib.LogEventStatusStarted,
					Metadata: &batcheslib.TaskPreparingStepMetadata{
						Step: 1,
					},
				},
				{
					Timestamp: time2,
					Status:    batcheslib.LogEventStatusStarted,
					Metadata: &batcheslib.TaskStepMetadata{
						Step: 1,
						Env:  map[string]string{"env": "var"},
					},
				},
			},
			want: map[int]*StepInfo{
				1: {
					StartedAt:   time1,
					Environment: map[string]string{"env": "var"},
				},
			},
		},
		{
			name: "With logs",
			lines: []*batcheslib.LogEvent{
				{
					Timestamp: time1,
					Status:    batcheslib.LogEventStatusStarted,
					Metadata: &batcheslib.TaskPreparingStepMetadata{
						Step: 1,
					},
				},
				{
					Timestamp: time2,
					Status:    batcheslib.LogEventStatusStarted,
					Metadata: &batcheslib.TaskStepMetadata{
						Step: 1,
						Env:  map[string]string{},
					},
				},
				{
					Timestamp: time2,
					Status:    batcheslib.LogEventStatusProgress,
					Metadata: &batcheslib.TaskStepMetadata{
						Step: 1,
						Out:  "stdout: log1\nstdout: log2\n",
					},
				},
				{
					Timestamp: time2,
					Status:    batcheslib.LogEventStatusProgress,
					Metadata: &batcheslib.TaskStepMetadata{
						Step: 1,
						Out:  "stderr: log3\n",
					},
				},
				{
					Timestamp: time2,
					Status:    batcheslib.LogEventStatusProgress,
					Metadata: &batcheslib.TaskStepMetadata{
						Step: 1,
						Out:  "stdout: log4\n",
					},
				},
			},
			want: map[int]*StepInfo{
				1: {
					StartedAt:   time1,
					Environment: map[string]string{},
					OutputLines: []string{"stdout: log1", "stdout: log2", "stderr: log3", "stdout: log4"},
				},
			},
		},
		{
			name:  "Started but timeout",
			entry: executor.ExecutionLogEntry{StartTime: time1, ExitCode: pointers.Ptr(-1), DurationMs: pointers.Ptr(500)},
			lines: []*batcheslib.LogEvent{
				{
					Timestamp: time1,
					Status:    batcheslib.LogEventStatusStarted,
					Metadata:  &batcheslib.TaskPreparingStepMetadata{Step: 1},
				},
				{
					Timestamp: time1,
					Status:    batcheslib.LogEventStatusSuccess,
					Metadata:  &batcheslib.TaskPreparingStepMetadata{Step: 1},
				},
				{
					Timestamp: time2,
					Status:    batcheslib.LogEventStatusStarted,
					Metadata: &batcheslib.TaskStepMetadata{
						Step: 1,
						Env:  map[string]string{"env": "var"},
					},
				},
			},
			want: map[int]*StepInfo{
				1: {
					StartedAt:   time1,
					FinishedAt:  time1.Add(500 * time.Millisecond),
					ExitCode:    pointers.Ptr(-1),
					Environment: map[string]string{"env": "var"},
				},
			},
		},
		{
			name: "Finished with error",
			lines: []*batcheslib.LogEvent{
				{
					Timestamp: time1,
					Status:    batcheslib.LogEventStatusStarted,
					Metadata: &batcheslib.TaskPreparingStepMetadata{
						Step: 1,
					},
				},
				{
					Timestamp: time2,
					Status:    batcheslib.LogEventStatusStarted,
					Metadata: &batcheslib.TaskStepMetadata{
						Step: 1,
						Env:  map[string]string{},
					},
				},
				{
					Timestamp: time3,
					Status:    batcheslib.LogEventStatusFailure,
					Metadata: &batcheslib.TaskStepMetadata{
						Step:     1,
						Error:    "very bad error",
						ExitCode: nonZero,
					},
				},
			},
			want: map[int]*StepInfo{
				1: {
					StartedAt:   time1,
					FinishedAt:  time3,
					Environment: make(map[string]string),
					ExitCode:    &nonZero,
				},
			},
		},
		{
			name: "Finished with success",
			lines: []*batcheslib.LogEvent{
				{
					Timestamp: time1,
					Status:    batcheslib.LogEventStatusStarted,
					Metadata: &batcheslib.TaskPreparingStepMetadata{
						Step: 1,
					},
				},
				{
					Timestamp: time2,
					Status:    batcheslib.LogEventStatusStarted,
					Metadata: &batcheslib.TaskStepMetadata{
						Step: 1,
					},
				},
				{
					Timestamp: time3,
					Status:    batcheslib.LogEventStatusSuccess,
					Metadata: &batcheslib.TaskStepMetadata{
						Step:     1,
						ExitCode: zero,
						Diff:     diff,
						Outputs:  map[string]any{"test": 1},
					},
				},
			},
			want: map[int]*StepInfo{
				1: {
					StartedAt:       time1,
					FinishedAt:      time3,
					Environment:     make(map[string]string),
					OutputVariables: map[string]any{"test": 1},
					ExitCode:        &zero,
					Diff:            diff,
					DiffFound:       true,
				},
			},
		},
		{
			name: "Complex",
			lines: []*batcheslib.LogEvent{
				{
					Timestamp: time1,
					Status:    batcheslib.LogEventStatusStarted,
					Metadata: &batcheslib.TaskPreparingStepMetadata{
						Step: 2,
					},
				},
				{
					Timestamp: time1,
					Status:    batcheslib.LogEventStatusStarted,
					Metadata: &batcheslib.TaskPreparingStepMetadata{
						Step: 1,
					},
				},
				{
					Timestamp: time2,
					Status:    batcheslib.LogEventStatusStarted,
					Metadata: &batcheslib.TaskStepMetadata{
						Step: 1,
						Env:  map[string]string{"env": "var"},
					},
				},
				{
					Timestamp: time2,
					Status:    batcheslib.LogEventStatusStarted,
					Metadata: &batcheslib.TaskStepMetadata{
						Step: 2,
						Env:  map[string]string{},
					},
				},
				{
					Timestamp: time2,
					Status:    batcheslib.LogEventStatusProgress,
					Metadata: &batcheslib.TaskStepMetadata{
						Step: 2,
						Out:  "stdout: log1\nstdout: log2\n",
					},
				},
				{
					Timestamp: time2,
					Status:    batcheslib.LogEventStatusProgress,
					Metadata: &batcheslib.TaskStepMetadata{
						Step: 1,
						Out:  "stdout: log1\nstdout: log2\n",
					},
				},
				{
					Timestamp: time2,
					Status:    batcheslib.LogEventStatusProgress,
					Metadata: &batcheslib.TaskStepMetadata{
						Step: 1,
						Out:  "stderr: log3\n",
					},
				},
				{
					Timestamp: time2,
					Status:    batcheslib.LogEventStatusProgress,
					Metadata: &batcheslib.TaskStepMetadata{
						Step: 2,
						Out:  "stderr: log3\n",
					},
				},
				{
					Timestamp: time3,
					Status:    batcheslib.LogEventStatusSuccess,
					Metadata: &batcheslib.TaskStepMetadata{
						Step:     1,
						ExitCode: zero,
						Diff:     diff,
						Outputs:  map[string]any{"test": 1},
					},
				},
				{
					Timestamp: time3,
					Status:    batcheslib.LogEventStatusFailure,
					Metadata: &batcheslib.TaskStepMetadata{
						Step:     2,
						ExitCode: nonZero,
						Error:    "very bad error",
					},
				},
			},
			want: map[int]*StepInfo{
				1: {
					StartedAt:       time1,
					FinishedAt:      time3,
					Environment:     map[string]string{"env": "var"},
					OutputVariables: map[string]any{"test": 1},
					OutputLines:     []string{"stdout: log1", "stdout: log2", "stderr: log3"},
					ExitCode:        &zero,
					Diff:            diff,
					DiffFound:       true,
				},
				2: {
					StartedAt:   time1,
					FinishedAt:  time3,
					Environment: make(map[string]string),
					OutputLines: []string{"stdout: log1", "stdout: log2", "stderr: log3"},
					ExitCode:    &nonZero,
					// TODO: Where do we expose error?
				},
			},
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			have := ParseLogLines(tc.entry, tc.lines)
			if diff := cmp.Diff(have, tc.want); diff != "" {
				t.Errorf("invalid steps returned %s", diff)
			}
		})
	}
}

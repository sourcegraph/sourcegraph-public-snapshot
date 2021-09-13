package background

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	ct "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/testing"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
)

func TestLoadAndExtractChangesetSpecIDs(t *testing.T) {
	db := dbtest.NewDB(t, "")
	user := ct.CreateTestUser(t, db, true)

	repo, _ := ct.CreateTestRepo(t, context.Background(), db)

	s := store.New(db, &observation.TestContext, nil)
	workStore := dbworkerstore.NewWithMetrics(s.Handle(), batchSpecWorkspaceExecutionWorkerStoreOptions, &observation.TestContext)

	batchSpec := &btypes.BatchSpec{UserID: user.ID, NamespaceUserID: user.ID, RawSpec: "horse"}
	if err := s.CreateBatchSpec(context.Background(), batchSpec); err != nil {
		t.Fatal(err)
	}

	workspace := &btypes.BatchSpecWorkspace{
		BatchSpecID: batchSpec.ID,
		RepoID:      repo.ID,
		Steps:       []batcheslib.Step{},
	}
	if err := s.CreateBatchSpecWorkspace(context.Background(), workspace); err != nil {
		t.Fatal(err)
	}

	var (
		changesetSpecIDs        []int64
		changesetSpecGraphQLIDs []string
	)

	for i := 0; i < 3; i++ {
		spec := &btypes.ChangesetSpec{RepoID: repo.ID, UserID: user.ID}
		if err := s.CreateChangesetSpec(context.Background(), spec); err != nil {
			t.Fatal(err)
		}
		changesetSpecIDs = append(changesetSpecIDs, spec.ID)
		changesetSpecGraphQLIDs = append(changesetSpecGraphQLIDs, fmt.Sprintf("%q", relay.MarshalID("doesnotmatter", spec.RandID)))
	}

	jsonArray := `[` + strings.Join(changesetSpecGraphQLIDs, ",") + `]`

	t.Run("success", func(t *testing.T) {
		job := &btypes.BatchSpecWorkspaceExecutionJob{
			State:                btypes.BatchSpecWorkspaceExecutionJobStateProcessing,
			BatchSpecWorkspaceID: workspace.ID,
		}

		if err := s.CreateBatchSpecWorkspaceExecutionJob(context.Background(), job); err != nil {
			t.Fatal(err)
		}

		entries := []workerutil.ExecutionLogEntry{
			{
				Key:        "setup.firecracker.start",
				Command:    []string{"ignite", "run"},
				StartTime:  time.Now().Add(-5 * time.Second),
				Out:        `stdout: cool`,
				DurationMs: intptr(200),
			},
			{
				Key:       "step.src.0",
				Command:   []string{"src", "batch", "preview", "-f", "spec.yml", "-text-only"},
				StartTime: time.Now().Add(-5 * time.Second),
				Out: `stdout: {"operation":"PARSING_BATCH_SPEC","timestamp":"2021-07-06T09:38:51.481Z","status":"STARTED"}
stdout: {"operation":"PARSING_BATCH_SPEC","timestamp":"2021-07-06T09:38:51.481Z","status":"SUCCESS"}
stdout: {"operation":"UPLOADING_CHANGESET_SPECS","timestamp":"2021-09-09T13:20:32.942Z","status":"STARTED","metadata":{"total":1}}
stdout: {"operation":"UPLOADING_CHANGESET_SPECS","timestamp":"2021-09-09T13:20:32.95Z","status":"PROGRESS","metadata":{"done":1,"total":1}}
stdout: {"operation":"UPLOADING_CHANGESET_SPECS","timestamp":"2021-09-09T13:20:32.95Z","status":"SUCCESS","metadata":{"ids":` + jsonArray + `}}
`,
				DurationMs: intptr(200),
			},
		}

		for i, e := range entries {
			entryID, err := workStore.AddExecutionLogEntry(context.Background(), int(job.ID), e, dbworkerstore.ExecutionLogEntryOptions{})
			if err != nil {
				t.Fatal(err)
			}
			if entryID != i+1 {
				t.Fatalf("AddExecutionLogEntry returned wrong entryID. want=%d, have=%d", i+1, entryID)
			}
		}

		have, err := loadAndExtractChangesetSpecIDs(context.Background(), s, job.ID)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if diff := cmp.Diff(changesetSpecIDs, have); diff != "" {
			t.Fatalf("wrong diff: %s", diff)
		}
	})

	t.Run("without log entry", func(t *testing.T) {
		job := &btypes.BatchSpecWorkspaceExecutionJob{
			State:                btypes.BatchSpecWorkspaceExecutionJobStateProcessing,
			BatchSpecWorkspaceID: workspace.ID,
		}

		if err := s.CreateBatchSpecWorkspaceExecutionJob(context.Background(), job); err != nil {
			t.Fatal(err)
		}

		_, err := loadAndExtractChangesetSpecIDs(context.Background(), s, job.ID)
		if err == nil {
			t.Fatalf("expected error but got none")
		}

		if err.Error() != "no execution logs" {
			t.Fatalf("wrong error: %q", err.Error())
		}
	})
}

func TestExtractChangesetSpecIDs(t *testing.T) {
	tests := []struct {
		name        string
		entries     []workerutil.ExecutionLogEntry
		wantRandIDs []string
		wantErr     error
	}{
		{
			name: "success",
			entries: []workerutil.ExecutionLogEntry{
				{Key: "setup.firecracker.start"},
				// Reduced log output because we don't care about _all_ lines
				{
					Key: "step.src.0",
					Out: `
stdout: {"operation":"EXECUTING_TASKS","timestamp":"2021-09-09T13:20:32.942Z","status":"SUCCESS"}
stdout: {"operation":"UPLOADING_CHANGESET_SPECS","timestamp":"2021-09-09T13:20:32.942Z","status":"STARTED","metadata":{"total":1}}
stdout: {"operation":"UPLOADING_CHANGESET_SPECS","timestamp":"2021-09-09T13:20:32.95Z","status":"PROGRESS","metadata":{"done":1,"total":1}}
stdout: {"operation":"UPLOADING_CHANGESET_SPECS","timestamp":"2021-09-09T13:20:32.95Z","status":"SUCCESS","metadata":{"ids":["Q2hhbmdlc2V0U3BlYzoiNkxIYWN5dkI3WDYi"]}}

`,
				},
			},
			// Run `echo "QmF0Y2hTcGVjOiJBZFBMTDU5SXJmWCI=" |base64 -d` to get this
			wantRandIDs: []string{"6LHacyvB7X6"},
		},
		{

			name:    "no step.src.0 log entry",
			entries: []workerutil.ExecutionLogEntry{},
			wantErr: ErrNoChangesetSpecIDs,
		},

		{

			name: "no upload step in the output",
			entries: []workerutil.ExecutionLogEntry{
				{
					Key: "step.src.0",
					Out: `stdout: {"operation":"PARSING_BATCH_SPEC","timestamp":"2021-07-06T09:38:51.481Z","status":"STARTED"}
stdout: {"operation":"PARSING_BATCH_SPEC","timestamp":"2021-07-06T09:38:51.481Z","status":"SUCCESS"}
`,
				},
			},
			wantErr: ErrNoChangesetSpecIDs,
		},
		{
			name: "empty array the output",
			entries: []workerutil.ExecutionLogEntry{
				{
					Key: "step.src.0",
					Out: `
stdout: {"operation":"EXECUTING_TASKS","timestamp":"2021-09-09T13:20:32.942Z","status":"SUCCESS"}
stdout: {"operation":"UPLOADING_CHANGESET_SPECS","timestamp":"2021-09-09T13:20:32.942Z","status":"STARTED","metadata":{"total":1}}
stdout: {"operation":"UPLOADING_CHANGESET_SPECS","timestamp":"2021-09-09T13:20:32.95Z","status":"PROGRESS","metadata":{"done":1,"total":1}}
stdout: {"operation":"UPLOADING_CHANGESET_SPECS","timestamp":"2021-09-09T13:20:32.95Z","status":"SUCCESS","metadata":{"ids":[]}}

`,
				},
			},
			wantErr: ErrNoChangesetSpecIDs,
		},

		{
			name: "additional text in log output",
			entries: []workerutil.ExecutionLogEntry{
				{
					Key: "step.src.0",
					Out: `stdout: {"operation":"EXECUTING_TASKS","timestamp":"2021-09-09T13:20:32.941Z","status":"STARTED","metadata":{"tasks":[]}}
stdout: {"operation":"EXECUTING_TASKS","timestamp":"2021-09-09T13:20:32.942Z","status":"SUCCESS"}
stderr: HORSE
stdout: {"operation":"UPLOADING_CHANGESET_SPECS","timestamp":"2021-09-09T13:20:32.942Z","status":"STARTED","metadata":{"total":1}}
stderr: HORSE
stdout: {"operation":"UPLOADING_CHANGESET_SPECS","timestamp":"2021-09-09T13:20:32.95Z","status":"PROGRESS","metadata":{"done":1,"total":1}}
stderr: HORSE
stdout: {"operation":"UPLOADING_CHANGESET_SPECS","timestamp":"2021-09-09T13:20:32.95Z","status":"SUCCESS","metadata":{"ids":["Q2hhbmdlc2V0U3BlYzoiNkxIYWN5dkI3WDYi"]}}
`,
				},
			},
			wantRandIDs: []string{"6LHacyvB7X6"},
		},

		{
			name: "invalid json",
			entries: []workerutil.ExecutionLogEntry{
				{
					Key: "step.src.0",
					Out: `stdout: {"operation":"PARSING_BATCH_SPEC","timestamp":"2021-07-06T09:38:51.481Z","status":"STARTED"}
stdout: {HOOOORSE}
stdout: {HORSE}
stdout: {HORSE}
`,
				},
			},
			wantErr: ErrNoChangesetSpecIDs,
		},

		{
			name: "non-json output inbetween valid json",
			entries: []workerutil.ExecutionLogEntry{
				{
					Key: "step.src.0",
					Out: `stdout: {"operation":"PARSING_BATCH_SPEC","timestamp":"2021-07-12T12:25:33.965Z","status":"STARTED"}
stdout: No changeset specs created
stdout: {"operation":"UPLOADING_CHANGESET_SPECS","timestamp":"2021-09-09T13:20:32.95Z","status":"SUCCESS","metadata":{"ids":["Q2hhbmdlc2V0U3BlYzoiNkxIYWN5dkI3WDYi"]}}`,
				},
			},
			wantRandIDs: []string{"6LHacyvB7X6"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			have, err := extractChangesetSpecRandIDs(tt.entries)
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Fatalf("wrong error. want=%s, got=%s", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if diff := cmp.Diff(tt.wantRandIDs, have); diff != "" {
				t.Fatalf("wrong IDs extracted: %s", diff)
			}
		})
	}
}

package background

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/google/go-cmp/cmp"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	ct "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/testing"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
)

func TestBatchSpecWorkspaceExecutionWorkerStore_MarkComplete(t *testing.T) {
	ctx := context.Background()
	db := dbtest.NewDB(t, "")
	user := ct.CreateTestUser(t, db, true)

	repo, _ := ct.CreateTestRepo(t, ctx, db)

	s := store.New(db, &observation.TestContext, nil)
	workStore := dbworkerstore.NewWithMetrics(s.Handle(), batchSpecWorkspaceExecutionWorkerStoreOptions, &observation.TestContext)

	// Setup all the associations
	batchSpec := &btypes.BatchSpec{UserID: user.ID, NamespaceUserID: user.ID, RawSpec: "horse"}
	if err := s.CreateBatchSpec(ctx, batchSpec); err != nil {
		t.Fatal(err)
	}

	workspace := &btypes.BatchSpecWorkspace{BatchSpecID: batchSpec.ID, RepoID: repo.ID, Steps: []batcheslib.Step{}}
	if err := s.CreateBatchSpecWorkspace(ctx, workspace); err != nil {
		t.Fatal(err)
	}

	job := &btypes.BatchSpecWorkspaceExecutionJob{BatchSpecWorkspaceID: workspace.ID}
	if err := ct.CreateBatchSpecWorkspaceExecutionJob(ctx, s, store.ScanBatchSpecWorkspaceExecutionJob, job); err != nil {
		t.Fatal(err)
	}

	// Create changeset_specs to simulate the job resulting in changeset_specs being created
	var (
		changesetSpecIDs        []int64
		changesetSpecGraphQLIDs []string
	)

	for i := 0; i < 3; i++ {
		spec := &btypes.ChangesetSpec{RepoID: repo.ID, UserID: user.ID}
		if err := s.CreateChangesetSpec(ctx, spec); err != nil {
			t.Fatal(err)
		}
		changesetSpecIDs = append(changesetSpecIDs, spec.ID)
		changesetSpecGraphQLIDs = append(changesetSpecGraphQLIDs, fmt.Sprintf("%q", relay.MarshalID("doesnotmatter", spec.RandID)))
	}

	// Add a log entry that contains the changeset spec IDs
	jsonArray := `[` + strings.Join(changesetSpecGraphQLIDs, ",") + `]`
	entry := workerutil.ExecutionLogEntry{
		Key:        "step.src.0",
		Command:    []string{"src", "batch", "preview", "-f", "spec.yml", "-text-only"},
		StartTime:  time.Now().Add(-5 * time.Second),
		Out:        `stdout: {"operation":"UPLOADING_CHANGESET_SPECS","timestamp":"2021-09-09T13:20:32.95Z","status":"SUCCESS","metadata":{"ids":` + jsonArray + `}} `,
		DurationMs: intptr(200),
	}

	_, err := workStore.AddExecutionLogEntry(ctx, int(job.ID), entry, dbworkerstore.ExecutionLogEntryOptions{})
	if err != nil {
		t.Fatal(err)
	}

	executionStore := &batchSpecWorkspaceExecutionWorkerStore{Store: workStore, observationContext: &observation.TestContext}
	opts := dbworkerstore.MarkFinalOptions{WorkerHostname: "worker-1"}

	setProcessing := func(t *testing.T) {
		t.Helper()
		job.State = btypes.BatchSpecWorkspaceExecutionJobStateProcessing
		job.WorkerHostname = opts.WorkerHostname
		ct.UpdateJobState(t, ctx, s, job)
	}

	attachAccessToken := func(t *testing.T) int64 {
		t.Helper()
		tokenID, _, err := database.AccessTokens(db).CreateInternal(ctx, user.ID, []string{"user:all"}, "testing", user.ID)
		if err != nil {
			t.Fatal(err)
		}
		if err := s.SetBatchSpecWorkspaceExecutionJobAccessToken(ctx, job.ID, tokenID); err != nil {
			t.Fatal(err)
		}
		return tokenID
	}

	assertJobState := func(t *testing.T, want btypes.BatchSpecWorkspaceExecutionJobState) {
		t.Helper()
		reloadedJob, err := s.GetBatchSpecWorkspaceExecutionJob(ctx, store.GetBatchSpecWorkspaceExecutionJobOpts{ID: job.ID})
		if err != nil {
			t.Fatalf("failed to reload job: %s", err)
		}

		if have := reloadedJob.State; have != want {
			t.Fatalf("wrong job state: want=%s, have=%s", want, have)
		}
	}

	t.Run("success", func(t *testing.T) {
		setProcessing(t)
		tokenID := attachAccessToken(t)

		ok, err := executionStore.MarkComplete(context.Background(), int(job.ID), opts)
		if !ok || err != nil {
			t.Fatalf("MarkComplete failed. ok=%t, err=%s", ok, err)
		}

		// Now reload the involved entities and make sure they've been updated correctly
		assertJobState(t, btypes.BatchSpecWorkspaceExecutionJobStateCompleted)

		reloadedWorkspace, err := s.GetBatchSpecWorkspace(ctx, store.GetBatchSpecWorkspaceOpts{ID: workspace.ID})
		if err != nil {
			t.Fatalf("failed to reload workspace: %s", err)
		}
		if diff := cmp.Diff(changesetSpecIDs, reloadedWorkspace.ChangesetSpecIDs); diff != "" {
			t.Fatalf("reloaded workspace has wrong changeset spec IDs: %s", diff)
		}

		reloadedSpecs, _, err := s.ListChangesetSpecs(ctx, store.ListChangesetSpecsOpts{LimitOpts: store.LimitOpts{Limit: 0}, IDs: changesetSpecIDs})
		if err != nil {
			t.Fatalf("failed to reload changeset specs: %s", err)
		}
		for _, reloadedSpec := range reloadedSpecs {
			if reloadedSpec.BatchSpecID != batchSpec.ID {
				t.Fatalf("reloaded changeset spec does not have correct batch spec id: %d", reloadedSpec.BatchSpecID)
			}
		}

		_, err = database.AccessTokens(db).GetByID(ctx, tokenID)
		if err != database.ErrAccessTokenNotFound {
			t.Fatalf("access token was not deleted")
		}
	})

	t.Run("no token set", func(t *testing.T) {
		setProcessing(t)

		ok, err := executionStore.MarkComplete(context.Background(), int(job.ID), opts)
		if !ok || err != nil {
			t.Fatalf("MarkComplete failed. ok=%t, err=%s", ok, err)
		}

		assertJobState(t, btypes.BatchSpecWorkspaceExecutionJobStateCompleted)
	})

	t.Run("token set but deletion fails", func(t *testing.T) {
		setProcessing(t)
		tokenID := attachAccessToken(t)

		database.Mocks.AccessTokens.HardDeleteByID = func(id int64) error {
			if id != tokenID {
				t.Fatalf("wrong token deleted")
			}
			return errors.New("internal database error")
		}
		defer func() { database.Mocks.AccessTokens.HardDeleteByID = nil }()

		ok, err := executionStore.MarkComplete(context.Background(), int(job.ID), opts)
		if !ok || err != nil {
			t.Fatalf("MarkComplete failed. ok=%t, err=%s", ok, err)
		}

		assertJobState(t, btypes.BatchSpecWorkspaceExecutionJobStateFailed)
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
			// Run `echo "QmF0Y2hTcGVjOiJBZFBMTDU5SXJmWCI=" | base64 -d` to get this
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

func intptr(i int) *int { return &i }

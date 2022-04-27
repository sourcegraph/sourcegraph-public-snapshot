package store

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	ct "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/testing"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
	"github.com/sourcegraph/sourcegraph/lib/batches/execution"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestBatchSpecWorkspaceExecutionWorkerStore_MarkComplete(t *testing.T) {
	ctx := context.Background()
	db := database.NewDB(dbtest.NewDB(t))
	user := ct.CreateTestUser(t, db, true)

	repo, _ := ct.CreateTestRepo(t, ctx, db)
	s := New(db, &observation.TestContext, nil)
	workStore := dbworkerstore.NewWithMetrics(s.Handle(), batchSpecWorkspaceExecutionWorkerStoreOptions, &observation.TestContext)

	// Setup all the associations
	batchSpec := &btypes.BatchSpec{UserID: user.ID, NamespaceUserID: user.ID, RawSpec: "horse", Spec: &batcheslib.BatchSpec{
		ChangesetTemplate: &batcheslib.ChangesetTemplate{},
	}}
	if err := s.CreateBatchSpec(ctx, batchSpec); err != nil {
		t.Fatal(err)
	}

	workspace := &btypes.BatchSpecWorkspace{BatchSpecID: batchSpec.ID, RepoID: repo.ID}
	if err := s.CreateBatchSpecWorkspace(ctx, workspace); err != nil {
		t.Fatal(err)
	}

	job := &btypes.BatchSpecWorkspaceExecutionJob{BatchSpecWorkspaceID: workspace.ID}
	if err := ct.CreateBatchSpecWorkspaceExecutionJob(ctx, s, ScanBatchSpecWorkspaceExecutionJob, job); err != nil {
		t.Fatal(err)
	}

	// See the `output` var below
	cacheEntryKeys := []string{
		"Nsw12JxoLSHN4ta6D3G7FQ",
		"JkC7Q0OOCZZ3Acv79QfwSA-step-0",
		"0ydsSXJ77syIPdwNrsGlzQ-step-1",
		"utgLpuQ3njDtLe3eztArAQ-step-2",
		"RoG8xSgpganc5BJ0_D3XGA-step-3",
		"Nsw12JxoLSHN4ta6D3G7FQ-step-4",
	}

	// Log entries with cache entries that'll be used to build the changeset specs.
	output := `
stdout: {"operation":"CACHE_RESULT","timestamp":"2021-11-04T12:43:19.551Z","status":"SUCCESS","metadata":{"key":"Nsw12JxoLSHN4ta6D3G7FQ","value":{"diff":"diff --git README.md README.md\nindex 1914491..d6782d3 100644\n--- README.md\n+++ README.md\n@@ -3,4 +3,7 @@ This repository is used to test opening and closing pull request with Automation\n \n (c) Copyright Sourcegraph 2013-2020.\n (c) Copyright Sourcegraph 2013-2020.\n-(c) Copyright Sourcegraph 2013-2020.\n\\ No newline at end of file\n+(c) Copyright Sourcegraph 2013-2020.this is step 2\n+this is step 3\n+this is step 4\n+previous_step.modified_files=[README.md]\ndiff --git README.txt README.txt\nnew file mode 100644\nindex 0000000..888e1ec\n--- /dev/null\n+++ README.txt\n@@ -0,0 +1 @@\n+this is step 1\ndiff --git my-output.txt my-output.txt\nnew file mode 100644\nindex 0000000..257ae8e\n--- /dev/null\n+++ my-output.txt\n@@ -0,0 +1 @@\n+this is step 5\n","changedFiles":{"modified":["README.md"],"added":["README.txt","my-output.txt"],"deleted":null,"renamed":null},"outputs":{"myOutput":"my-output.txt"},"Path":""}}}
stdout: {"operation":"CACHE_AFTER_STEP_RESULT","timestamp":"2021-11-04T12:43:19.551Z","status":"SUCCESS","metadata":{"key":"JkC7Q0OOCZZ3Acv79QfwSA-step-0","value":{"stepIndex":0,"diff":"ZGlmZiAtLWdpdCBSRUFETUUudHh0IFJFQURNRS50eHQKbmV3IGZpbGUgbW9kZSAxMDA2NDQKaW5kZXggMDAwMDAwMC4uODg4ZTFlYwotLS0gL2Rldi9udWxsCisrKyBSRUFETUUudHh0CkBAIC0wLDAgKzEgQEAKK3RoaXMgaXMgc3RlcCAxCg==","outputs":{},"previousStepResult":{"Files":null,"Stdout":null,"Stderr":null}}}}
stdout: {"operation":"CACHE_AFTER_STEP_RESULT","timestamp":"2021-11-04T12:43:19.551Z","status":"SUCCESS","metadata":{"key":"0ydsSXJ77syIPdwNrsGlzQ-step-1","value":{"stepIndex":1,"diff":"ZGlmZiAtLWdpdCBSRUFETUUubWQgUkVBRE1FLm1kCmluZGV4IDE5MTQ0OTEuLjVjMmI3MmQgMTAwNjQ0Ci0tLSBSRUFETUUubWQKKysrIFJFQURNRS5tZApAQCAtMyw0ICszLDQgQEAgVGhpcyByZXBvc2l0b3J5IGlzIHVzZWQgdG8gdGVzdCBvcGVuaW5nIGFuZCBjbG9zaW5nIHB1bGwgcmVxdWVzdCB3aXRoIEF1dG9tYXRpb24KIAogKGMpIENvcHlyaWdodCBTb3VyY2VncmFwaCAyMDEzLTIwMjAuCiAoYykgQ29weXJpZ2h0IFNvdXJjZWdyYXBoIDIwMTMtMjAyMC4KLShjKSBDb3B5cmlnaHQgU291cmNlZ3JhcGggMjAxMy0yMDIwLgpcIE5vIG5ld2xpbmUgYXQgZW5kIG9mIGZpbGUKKyhjKSBDb3B5cmlnaHQgU291cmNlZ3JhcGggMjAxMy0yMDIwLnRoaXMgaXMgc3RlcCAyCmRpZmYgLS1naXQgUkVBRE1FLnR4dCBSRUFETUUudHh0Cm5ldyBmaWxlIG1vZGUgMTAwNjQ0CmluZGV4IDAwMDAwMDAuLjg4OGUxZWMKLS0tIC9kZXYvbnVsbAorKysgUkVBRE1FLnR4dApAQCAtMCwwICsxIEBACit0aGlzIGlzIHN0ZXAgMQo=","outputs":{},"previousStepResult":{"Files":{"modified":null,"added":["README.txt"],"deleted":null,"renamed":null},"Stdout":{},"Stderr":{}}}}}
stdout: {"operation":"CACHE_AFTER_STEP_RESULT","timestamp":"2021-11-04T12:43:19.551Z","status":"SUCCESS","metadata":{"key":"utgLpuQ3njDtLe3eztArAQ-step-2","value":{"stepIndex":2,"diff":"ZGlmZiAtLWdpdCBSRUFETUUubWQgUkVBRE1FLm1kCmluZGV4IDE5MTQ0OTEuLmNkMmNjYmYgMTAwNjQ0Ci0tLSBSRUFETUUubWQKKysrIFJFQURNRS5tZApAQCAtMyw0ICszLDUgQEAgVGhpcyByZXBvc2l0b3J5IGlzIHVzZWQgdG8gdGVzdCBvcGVuaW5nIGFuZCBjbG9zaW5nIHB1bGwgcmVxdWVzdCB3aXRoIEF1dG9tYXRpb24KIAogKGMpIENvcHlyaWdodCBTb3VyY2VncmFwaCAyMDEzLTIwMjAuCiAoYykgQ29weXJpZ2h0IFNvdXJjZWdyYXBoIDIwMTMtMjAyMC4KLShjKSBDb3B5cmlnaHQgU291cmNlZ3JhcGggMjAxMy0yMDIwLgpcIE5vIG5ld2xpbmUgYXQgZW5kIG9mIGZpbGUKKyhjKSBDb3B5cmlnaHQgU291cmNlZ3JhcGggMjAxMy0yMDIwLnRoaXMgaXMgc3RlcCAyCit0aGlzIGlzIHN0ZXAgMwpkaWZmIC0tZ2l0IFJFQURNRS50eHQgUkVBRE1FLnR4dApuZXcgZmlsZSBtb2RlIDEwMDY0NAppbmRleCAwMDAwMDAwLi44ODhlMWVjCi0tLSAvZGV2L251bGwKKysrIFJFQURNRS50eHQKQEAgLTAsMCArMSBAQAordGhpcyBpcyBzdGVwIDEK","outputs":{"myOutput":"my-output.txt"},"previousStepResult":{"Files":{"modified":["README.md"],"added":["README.txt"],"deleted":null,"renamed":null},"Stdout":{},"Stderr":{}}}}}
stdout: {"operation":"CACHE_AFTER_STEP_RESULT","timestamp":"2021-11-04T12:43:19.551Z","status":"SUCCESS","metadata":{"key":"RoG8xSgpganc5BJ0_D3XGA-step-3","value":{"stepIndex":3,"diff":"ZGlmZiAtLWdpdCBSRUFETUUubWQgUkVBRE1FLm1kCmluZGV4IDE5MTQ0OTEuLmQ2NzgyZDMgMTAwNjQ0Ci0tLSBSRUFETUUubWQKKysrIFJFQURNRS5tZApAQCAtMyw0ICszLDcgQEAgVGhpcyByZXBvc2l0b3J5IGlzIHVzZWQgdG8gdGVzdCBvcGVuaW5nIGFuZCBjbG9zaW5nIHB1bGwgcmVxdWVzdCB3aXRoIEF1dG9tYXRpb24KIAogKGMpIENvcHlyaWdodCBTb3VyY2VncmFwaCAyMDEzLTIwMjAuCiAoYykgQ29weXJpZ2h0IFNvdXJjZWdyYXBoIDIwMTMtMjAyMC4KLShjKSBDb3B5cmlnaHQgU291cmNlZ3JhcGggMjAxMy0yMDIwLgpcIE5vIG5ld2xpbmUgYXQgZW5kIG9mIGZpbGUKKyhjKSBDb3B5cmlnaHQgU291cmNlZ3JhcGggMjAxMy0yMDIwLnRoaXMgaXMgc3RlcCAyCit0aGlzIGlzIHN0ZXAgMwordGhpcyBpcyBzdGVwIDQKK3ByZXZpb3VzX3N0ZXAubW9kaWZpZWRfZmlsZXM9W1JFQURNRS5tZF0KZGlmZiAtLWdpdCBSRUFETUUudHh0IFJFQURNRS50eHQKbmV3IGZpbGUgbW9kZSAxMDA2NDQKaW5kZXggMDAwMDAwMC4uODg4ZTFlYwotLS0gL2Rldi9udWxsCisrKyBSRUFETUUudHh0CkBAIC0wLDAgKzEgQEAKK3RoaXMgaXMgc3RlcCAxCg==","outputs":{"myOutput":"my-output.txt"},"previousStepResult":{"Files":{"modified":["README.md"],"added":["README.txt"],"deleted":null,"renamed":null},"Stdout":{},"Stderr":{}}}}}
stdout: {"operation":"CACHE_AFTER_STEP_RESULT","timestamp":"2021-11-04T12:43:19.551Z","status":"SUCCESS","metadata":{"key":"Nsw12JxoLSHN4ta6D3G7FQ-step-4","value":{"stepIndex":4,"diff":"ZGlmZiAtLWdpdCBSRUFETUUubWQgUkVBRE1FLm1kCmluZGV4IDE5MTQ0OTEuLmQ2NzgyZDMgMTAwNjQ0Ci0tLSBSRUFETUUubWQKKysrIFJFQURNRS5tZApAQCAtMyw0ICszLDcgQEAgVGhpcyByZXBvc2l0b3J5IGlzIHVzZWQgdG8gdGVzdCBvcGVuaW5nIGFuZCBjbG9zaW5nIHB1bGwgcmVxdWVzdCB3aXRoIEF1dG9tYXRpb24KIAogKGMpIENvcHlyaWdodCBTb3VyY2VncmFwaCAyMDEzLTIwMjAuCiAoYykgQ29weXJpZ2h0IFNvdXJjZWdyYXBoIDIwMTMtMjAyMC4KLShjKSBDb3B5cmlnaHQgU291cmNlZ3JhcGggMjAxMy0yMDIwLgpcIE5vIG5ld2xpbmUgYXQgZW5kIG9mIGZpbGUKKyhjKSBDb3B5cmlnaHQgU291cmNlZ3JhcGggMjAxMy0yMDIwLnRoaXMgaXMgc3RlcCAyCit0aGlzIGlzIHN0ZXAgMwordGhpcyBpcyBzdGVwIDQKK3ByZXZpb3VzX3N0ZXAubW9kaWZpZWRfZmlsZXM9W1JFQURNRS5tZF0KZGlmZiAtLWdpdCBSRUFETUUudHh0IFJFQURNRS50eHQKbmV3IGZpbGUgbW9kZSAxMDA2NDQKaW5kZXggMDAwMDAwMC4uODg4ZTFlYwotLS0gL2Rldi9udWxsCisrKyBSRUFETUUudHh0CkBAIC0wLDAgKzEgQEAKK3RoaXMgaXMgc3RlcCAxCmRpZmYgLS1naXQgbXktb3V0cHV0LnR4dCBteS1vdXRwdXQudHh0Cm5ldyBmaWxlIG1vZGUgMTAwNjQ0CmluZGV4IDAwMDAwMDAuLjI1N2FlOGUKLS0tIC9kZXYvbnVsbAorKysgbXktb3V0cHV0LnR4dApAQCAtMCwwICsxIEBACit0aGlzIGlzIHN0ZXAgNQo=","outputs":{"myOutput":"my-output.txt"},"previousStepResult":{"Files":{"modified":["README.md"],"added":["README.txt"],"deleted":null,"renamed":null},"Stdout":{},"Stderr":{}}}}}`

	entry := workerutil.ExecutionLogEntry{
		Key:        "step.src.0",
		Command:    []string{"src", "batch", "preview", "-f", "spec.yml", "-text-only"},
		StartTime:  time.Now().Add(-5 * time.Second),
		Out:        output,
		DurationMs: intptr(200),
	}

	_, err := workStore.AddExecutionLogEntry(ctx, int(job.ID), entry, dbworkerstore.ExecutionLogEntryOptions{})
	if err != nil {
		t.Fatal(err)
	}

	executionStore := &batchSpecWorkspaceExecutionWorkerStore{Store: workStore, observationContext: &observation.TestContext, accessTokenDeleterForTX: func(tx *Store) accessTokenHardDeleter { return tx.DatabaseDB().AccessTokens().HardDeleteByID }}
	opts := dbworkerstore.MarkFinalOptions{WorkerHostname: "worker-1"}

	setProcessing := func(t *testing.T) {
		t.Helper()
		job.State = btypes.BatchSpecWorkspaceExecutionJobStateProcessing
		job.WorkerHostname = opts.WorkerHostname
		ct.UpdateJobState(t, ctx, s, job)
	}

	attachAccessToken := func(t *testing.T) int64 {
		t.Helper()
		tokenID, _, err := db.AccessTokens().CreateInternal(ctx, user.ID, []string{"user:all"}, "testing", user.ID)
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
		reloadedJob, err := s.GetBatchSpecWorkspaceExecutionJob(ctx, GetBatchSpecWorkspaceExecutionJobOpts{ID: job.ID})
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

		reloadedWorkspace, err := s.GetBatchSpecWorkspace(ctx, GetBatchSpecWorkspaceOpts{ID: workspace.ID})
		if err != nil {
			t.Fatalf("failed to reload workspace: %s", err)
		}

		specs, _, err := s.ListChangesetSpecs(ctx, ListChangesetSpecsOpts{BatchSpecID: batchSpec.ID})
		if err != nil {
			t.Fatalf("failed to load changeset specs: %s", err)
		}
		if have, want := len(specs), 1; have != want {
			t.Fatalf("invalid number of changeset specs created: have=%d want=%d", have, want)
		}
		changesetSpecIDs := make([]int64, 0, len(specs))
		for _, reloadedSpec := range specs {
			changesetSpecIDs = append(changesetSpecIDs, reloadedSpec.ID)
			if reloadedSpec.BatchSpecID != batchSpec.ID {
				t.Fatalf("reloaded changeset spec does not have correct batch spec id: %d", reloadedSpec.BatchSpecID)
			}
		}

		if diff := cmp.Diff(changesetSpecIDs, reloadedWorkspace.ChangesetSpecIDs); diff != "" {
			t.Fatalf("reloaded workspace has wrong changeset spec IDs: %s", diff)
		}

		for _, wantKey := range cacheEntryKeys {
			entries, err := s.ListBatchSpecExecutionCacheEntries(ctx, ListBatchSpecExecutionCacheEntriesOpts{
				UserID: user.ID,
				Keys:   []string{wantKey},
			})
			if err != nil {
				t.Fatal(err)
			}
			if len(entries) != 1 {
				t.Fatal("cache entry not found")
			}
			entry := entries[0]

			var cachedExecutionResult *execution.Result
			if err := json.Unmarshal([]byte(entry.Value), &cachedExecutionResult); err != nil {
				t.Fatal(err)
			}
			if cachedExecutionResult.Diff == "" {
				t.Fatalf("wrong diff extracted")
			}
		}

		_, err = db.AccessTokens().GetByID(ctx, tokenID)
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

		accessTokens := database.NewMockAccessTokenStore()
		accessTokens.HardDeleteByIDFunc.SetDefaultHook(func(ctx context.Context, id int64) error {
			if id != tokenID {
				t.Fatalf("wrong token deleted")
			}
			return errors.New("internal database error")
		})

		prevDeleter := executionStore.accessTokenDeleterForTX
		executionStore.accessTokenDeleterForTX = func(tx *Store) accessTokenHardDeleter {
			return accessTokens.HardDeleteByID
		}
		t.Cleanup(func() {
			executionStore.accessTokenDeleterForTX = prevDeleter
		})

		ok, err := executionStore.MarkComplete(context.Background(), int(job.ID), opts)
		if !ok || err != nil {
			t.Fatalf("MarkComplete failed. ok=%t, err=%s", ok, err)
		}

		assertJobState(t, btypes.BatchSpecWorkspaceExecutionJobStateFailed)
	})
}

func TestBatchSpecWorkspaceExecutionWorkerStore_MarkComplete_EmptyDiff(t *testing.T) {
	ctx := context.Background()
	db := database.NewDB(dbtest.NewDB(t))
	user := ct.CreateTestUser(t, db, true)

	repo, _ := ct.CreateTestRepo(t, ctx, db)

	s := New(db, &observation.TestContext, nil)
	workStore := dbworkerstore.NewWithMetrics(s.Handle(), batchSpecWorkspaceExecutionWorkerStoreOptions, &observation.TestContext)

	// Setup all the associations
	batchSpec := &btypes.BatchSpec{UserID: user.ID, NamespaceUserID: user.ID, RawSpec: "horse", Spec: &batcheslib.BatchSpec{
		ChangesetTemplate: &batcheslib.ChangesetTemplate{},
	}}
	if err := s.CreateBatchSpec(ctx, batchSpec); err != nil {
		t.Fatal(err)
	}

	workspace := &btypes.BatchSpecWorkspace{BatchSpecID: batchSpec.ID, RepoID: repo.ID}
	if err := s.CreateBatchSpecWorkspace(ctx, workspace); err != nil {
		t.Fatal(err)
	}

	job := &btypes.BatchSpecWorkspaceExecutionJob{BatchSpecWorkspaceID: workspace.ID}
	if err := ct.CreateBatchSpecWorkspaceExecutionJob(ctx, s, ScanBatchSpecWorkspaceExecutionJob, job); err != nil {
		t.Fatal(err)
	}

	cacheEntryKeys := []string{"Nsw12JxoLSHN4ta6D3G7FQ", "JkC7Q0OOCZZ3Acv79QfwSA-step-0"}

	// Log entries with cache entries that'll be used to build the changeset specs.
	output := `
stdout: {"operation":"CACHE_RESULT","timestamp":"2021-11-04T12:43:19.551Z","status":"SUCCESS","metadata":{"key":"Nsw12JxoLSHN4ta6D3G7FQ","value":{"diff":"","changedFiles":{"modified":null,"added":null,"deleted":null,"renamed":null},"outputs":{},"Path":""}}}
stdout: {"operation":"CACHE_AFTER_STEP_RESULT","timestamp":"2021-11-04T12:43:19.551Z","status":"SUCCESS","metadata":{"key":"JkC7Q0OOCZZ3Acv79QfwSA-step-0","value":{"stepIndex":0,"diff":"","outputs":{},"previousStepResult":{"Files":null,"Stdout":null,"Stderr":null}}}}`

	entry := workerutil.ExecutionLogEntry{
		Key:        "step.src.0",
		Command:    []string{"src", "batch", "preview", "-f", "spec.yml", "-text-only"},
		StartTime:  time.Now().Add(-5 * time.Second),
		Out:        output,
		DurationMs: intptr(200),
	}

	_, err := workStore.AddExecutionLogEntry(ctx, int(job.ID), entry, dbworkerstore.ExecutionLogEntryOptions{})
	if err != nil {
		t.Fatal(err)
	}

	executionStore := &batchSpecWorkspaceExecutionWorkerStore{Store: workStore, observationContext: &observation.TestContext, accessTokenDeleterForTX: func(tx *Store) accessTokenHardDeleter { return tx.DatabaseDB().AccessTokens().HardDeleteByID }}
	opts := dbworkerstore.MarkFinalOptions{WorkerHostname: "worker-1"}

	attachAccessToken := func(t *testing.T) int64 {
		t.Helper()
		tokenID, _, err := db.AccessTokens().CreateInternal(ctx, user.ID, []string{"user:all"}, "testing", user.ID)
		if err != nil {
			t.Fatal(err)
		}
		if err := s.SetBatchSpecWorkspaceExecutionJobAccessToken(ctx, job.ID, tokenID); err != nil {
			t.Fatal(err)
		}
		return tokenID
	}

	job.State = btypes.BatchSpecWorkspaceExecutionJobStateProcessing
	job.WorkerHostname = opts.WorkerHostname
	ct.UpdateJobState(t, ctx, s, job)
	tokenID := attachAccessToken(t)

	ok, err := executionStore.MarkComplete(context.Background(), int(job.ID), opts)
	if !ok || err != nil {
		t.Fatalf("MarkComplete failed. ok=%t, err=%s", ok, err)
	}

	specs, _, err := s.ListChangesetSpecs(ctx, ListChangesetSpecsOpts{BatchSpecID: batchSpec.ID})
	if err != nil {
		t.Fatalf("failed to load changeset specs: %s", err)
	}
	if have, want := len(specs), 0; have != want {
		t.Fatalf("invalid number of changeset specs created: have=%d want=%d", have, want)
	}

	for _, wantKey := range cacheEntryKeys {
		entries, err := s.ListBatchSpecExecutionCacheEntries(ctx, ListBatchSpecExecutionCacheEntriesOpts{
			UserID: user.ID,
			Keys:   []string{wantKey},
		})
		if err != nil {
			t.Fatal(err)
		}
		if len(entries) != 1 {
			t.Fatal("cache entry not found")
		}
	}

	_, err = db.AccessTokens().GetByID(ctx, tokenID)
	if err != database.ErrAccessTokenNotFound {
		t.Fatalf("access token was not deleted")
	}
}

func intptr(i int) *int { return &i }

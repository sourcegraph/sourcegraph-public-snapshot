package main

import (
	"context"
	"encoding/json"
	"log"
	"sort"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/batches/store"
	"github.com/sourcegraph/sourcegraph/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/gqltestutil"
	"github.com/sourcegraph/sourcegraph/lib/batches"
	"github.com/sourcegraph/sourcegraph/lib/batches/execution"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Test struct {
	PreExistingCacheEntries map[string]execution.AfterStepResult
	BatchSpecInput          string
	ExpectedCacheEntries    map[string]execution.AfterStepResult
	ExpectedChangesetSpecs  []*types.ChangesetSpec
	ExpectedState           gqltestutil.BatchSpecDeep
	CacheDisabled           bool
}

func RunTest(ctx context.Context, client *gqltestutil.Client, bstore *store.Store, test Test) error {
	// Reset DB state.
	if err := bstore.Exec(ctx, sqlf.Sprintf(cleanupBatchChangesDB)); err != nil {
		return err
	}

	_, err := batches.ParseBatchSpec([]byte(test.BatchSpecInput))
	if err != nil {
		return err
	}

	for k, e := range test.PreExistingCacheEntries {
		es, err := json.Marshal(e)
		if err != nil {
			return err
		}

		if err := bstore.CreateBatchSpecExecutionCacheEntry(ctx, &types.BatchSpecExecutionCacheEntry{
			Key:   k,
			Value: string(es),
		}); err != nil {
			return err
		}
	}

	log.Println("fetching user ID")

	id, err := client.CurrentUserID("")
	if err != nil {
		return err
	}

	var userID int32
	if err := relay.UnmarshalSpec(graphql.ID(id), &userID); err != nil {
		return err
	}

	log.Println("Creating empty batch change")

	batchChangeID, err := client.CreateEmptyBatchChange(id, "e2e-test-batch-change")
	if err != nil {
		return err
	}

	log.Println("Creating batch spec")

	batchSpecID, err := client.CreateBatchSpecFromRaw(batchChangeID, id, test.BatchSpecInput)
	if err != nil {
		return err
	}

	log.Println("Waiting for batch spec workspace resolution to finish")

	start := time.Now()
	for {
		if time.Since(start) > 60*time.Second {
			return errors.New("Waiting for batch spec workspace resolution to complete timed out after 60s")
		}
		state, err := client.GetBatchSpecWorkspaceResolutionStatus(batchSpecID)
		if err != nil {
			return err
		}

		// Resolution done, let's go!
		if state == "COMPLETED" {
			break
		}

		if state == "FAILED" || state == "ERRORED" {
			return errors.New("Batch spec workspace resolution failed")
		}
	}

	log.Println("Submitting execution for batch spec")

	// We're off, start the execution.
	if err := client.ExecuteBatchSpec(batchSpecID, test.CacheDisabled); err != nil {
		return err
	}

	log.Println("Waiting for batch spec execution to finish")

	start = time.Now()
	for {
		// Wait for at most 3 minutes to complete.
		if time.Since(start) > 3*60*time.Second {
			return errors.New("Waiting for batch spec execution to complete timed out after 3 min")
		}
		state, failureMessage, err := client.GetBatchSpecState(batchSpecID)
		if err != nil {
			return err
		}
		if state == "FAILED" {
			spec, err := client.GetBatchSpecDeep(batchSpecID)
			if err != nil {
				return err
			}
			d, err := json.MarshalIndent(spec, "", "")
			if err != nil {
				return err
			}
			log.Printf("Batch spec failed:\nFailure message: %s\nSpec: %s\n", failureMessage, string(d))
			return errors.New("Batch spec ended in failed state")
		}
		// Execution is complete, proceed!
		if state == "COMPLETED" {
			break
		}
	}

	log.Println("Loading batch spec to assert")

	gqlResp, err := client.GetBatchSpecDeep(batchSpecID)
	if err != nil {
		return err
	}

	if diff := cmp.Diff(*gqlResp, test.ExpectedState, compareBatchSpecDeepCmpopts()...); diff != "" {
		log.Printf("Batch spec diff detected: %s\n", diff)
		return errors.New("batch spec not in expected state")
	}

	log.Println("Verifying cache entries")

	// Verify the correct cache entries are in the database now.
	haveEntries, err := bstore.ListBatchSpecExecutionCacheEntries(ctx, store.ListBatchSpecExecutionCacheEntriesOpts{
		All: true,
	})
	if err != nil {
		return err
	}
	haveEntriesMap := map[string]execution.AfterStepResult{}
	for _, e := range haveEntries {
		var c execution.AfterStepResult
		if err := json.Unmarshal([]byte(e.Value), &c); err != nil {
			return err
		}
		haveEntriesMap[e.Key] = c
	}

	if diff := cmp.Diff(haveEntriesMap, test.ExpectedCacheEntries); diff != "" {
		log.Printf("Cache entries diff detected: %s\n", diff)
		return errors.New("cache entries not in correct state")
	}

	log.Println("Verifying changeset specs")

	// Verify the correct changeset specs are in the database now.
	haveCSS, _, err := bstore.ListChangesetSpecs(ctx, store.ListChangesetSpecsOpts{})
	if err != nil {
		return err
	}
	// Sort so it's comparable.
	sort.Slice(haveCSS, func(i, j int) bool {
		return haveCSS[i].BaseRepoID < haveCSS[j].BaseRepoID
	})
	sort.Slice(test.ExpectedChangesetSpecs, func(i, j int) bool {
		return test.ExpectedChangesetSpecs[i].BaseRepoID < test.ExpectedChangesetSpecs[j].BaseRepoID
	})

	if diff := cmp.Diff([]*types.ChangesetSpec(haveCSS), test.ExpectedChangesetSpecs, cmpopts.IgnoreFields(types.ChangesetSpec{}, "ID", "RandID", "CreatedAt", "UpdatedAt")); diff != "" {
		log.Printf("Changeset specs diff detected: %s\n", diff)
		return errors.New("changeset specs not in correct state")
	}

	log.Println("Passed!")

	return nil
}

const cleanupBatchChangesDB = `
DELETE FROM batch_changes;
DELETE FROM executor_secrets;
DELETE FROM batch_specs;
DELETE FROM batch_spec_workspace_execution_last_dequeues;
DELETE FROM batch_spec_workspace_files;
DELETE FROM changeset_specs;
`

func compareBatchSpecDeepCmpopts() []cmp.Option {
	// TODO: Reduce the number of ignores in here.
	return []cmp.Option{
		cmpopts.IgnoreFields(gqltestutil.BatchSpecDeep{}, "ID", "CreatedAt", "FinishedAt", "StartedAt", "ExpiresAt"),
		cmpopts.IgnoreFields(gqltestutil.ChangesetSpec{}, "ID"),
		cmpopts.IgnoreFields(gqltestutil.BatchSpecWorkspace{}, "QueuedAt", "StartedAt", "FinishedAt"),
		cmpopts.IgnoreFields(gqltestutil.BatchSpecWorkspaceStep{}, "StartedAt", "FinishedAt", "OutputLines"),
		cmpopts.IgnoreFields(gqltestutil.WorkspaceChangesetSpec{}, "ID"),
		cmpopts.IgnoreFields(gqltestutil.Namespace{}, "ID"),
		cmpopts.IgnoreFields(gqltestutil.Executor{}, "Hostname"),
		cmpopts.IgnoreFields(gqltestutil.ExecutionLogEntry{}, "Command", "StartTime", "Out", "DurationMilliseconds"),
	}
}

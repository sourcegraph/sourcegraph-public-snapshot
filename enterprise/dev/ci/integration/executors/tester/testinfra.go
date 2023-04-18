package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
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
	FileUpload              gqltestutil.BatchSpecFile
}

func RunTest(ctx context.Context, gqlClient *gqltestutil.Client, httpClient *HttpClient, bstore *store.Store, test Test) error {
	// Reset DB state, but only if we're not testing caching.
	if test.CacheDisabled {
		if err := bstore.Exec(ctx, sqlf.Sprintf(cleanupBatchChangesDB)); err != nil {
			return err
		}
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

	id, err := gqlClient.CurrentUserID("")
	if err != nil {
		return err
	}

	var userID int32
	if err := relay.UnmarshalSpec(graphql.ID(id), &userID); err != nil {
		return err
	}

	log.Println("Creating empty batch change")

	batchChangeID, err := gqlClient.CreateEmptyBatchChange(id, fmt.Sprintf("e2e-test-batch-change-cached-%t", !test.CacheDisabled))
	if err != nil {
		return err
	}

	log.Println("Creating batch spec")

	batchSpecID, err := gqlClient.CreateBatchSpecFromRaw(batchChangeID, id, test.BatchSpecInput)
	if err != nil {
		return err
	}

	log.Println("Waiting for batch spec workspace resolution to finish")

	start := time.Now()
	for {
		if time.Since(start) > 60*time.Second {
			return errors.New("Waiting for batch spec workspace resolution to complete timed out after 60s")
		}
		state, err := gqlClient.GetBatchSpecWorkspaceResolutionStatus(batchSpecID)
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

	if test.FileUpload.Name != "" {
		log.Println("Submitting mounted file")
		workdir, err := os.Getwd()
		if err != nil {
			return err
		}
		if err = httpClient.uploadFile(workdir, test.FileUpload.Name, batchSpecID); err != nil {
			return err
		}
	}

	log.Println("Submitting execution for batch spec")

	// We're off, start the execution.
	if err := gqlClient.ExecuteBatchSpec(batchSpecID, test.CacheDisabled); err != nil {
		return err
	}

	log.Println("Waiting for batch spec execution to finish")

	start = time.Now()
	for {
		// Wait for at most 3 minutes to complete.
		if time.Since(start) > 3*60*time.Second {
			return errors.New("Waiting for batch spec execution to complete timed out after 3 min")
		}
		state, failureMessage, err := gqlClient.GetBatchSpecState(batchSpecID)
		if err != nil {
			return err
		}
		if state == "FAILED" {
			spec, err := gqlClient.GetBatchSpecDeep(batchSpecID)
			if err != nil {
				return err
			}
			// TODO: this was not indented before, any reason or just oversight?
			d, err := json.MarshalIndent(spec, "", "  ")
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

	gqlResp, err := gqlClient.GetBatchSpecDeep(batchSpecID)
	if err != nil {
		return err
	}

	if diff := cmp.Diff(*gqlResp, test.ExpectedState, compareBatchSpecDeepCmpopts()...); diff != "" {
		log.Printf("Batch spec diff detected:\n%s", diff)
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

	if diff := cmp.Diff(haveEntriesMap, test.ExpectedCacheEntries, diffTransformer); diff != "" {
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

	if diff := cmp.Diff([]*types.ChangesetSpec(haveCSS), test.ExpectedChangesetSpecs, cmpopts.IgnoreFields(types.ChangesetSpec{}, "ID", "RandID", "CreatedAt", "UpdatedAt"), diffTransformer); diff != "" {
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
ALTER SEQUENCE batch_specs_id_seq RESTART;
UPDATE batch_specs SET id = DEFAULT;
DELETE FROM batch_spec_workspace_execution_last_dequeues;
DELETE FROM batch_spec_workspace_files;
DELETE FROM batch_spec_execution_cache_entries;
ALTER SEQUENCE batch_spec_execution_cache_entries_id_seq RESTART;
UPDATE batch_spec_execution_cache_entries SET id = DEFAULT;
DELETE FROM changeset_specs;
`

// diffTransformer prints only lines that are actually diffing on large multiline strings (such as changeset diffs).
// Without this transformer, the segment containing the diff can get truncated resulting in useless output.
var diffTransformer = cmpopts.AcyclicTransformer("onlyShowDiffs", func(in string) any { return strings.Split(in, "\n") })

func compareBatchSpecDeepCmpopts() []cmp.Option {
	// TODO: Reduce the number of ignores in here.

	return []cmp.Option{
		// The printed diff isn't guaranteed to display the lines that don't match if they are further down in the diff.
		// This transformer breaks up the string in separate lines which will only print the offending lines (and 2 around them).
		// See https://github.com/google/go-cmp/issues/311
		diffTransformer,
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

type HttpClient struct {
	token    string
	endpoint string
	client   *http.Client
}

func (c *HttpClient) uploadFile(workingDir, filePath, batchSpecID string) error {
	// Create a pipe so the requests can be chunked to the server
	pipeReader, pipeWriter := io.Pipe()
	multipartWriter := multipart.NewWriter(pipeWriter)

	// Write in a separate goroutine to properly chunk the file content. Writing to the pipe lets us not have
	// to put the whole file in memory.
	go func() {
		defer pipeWriter.Close()
		defer multipartWriter.Close()

		if err := createFormFile(multipartWriter, workingDir, filePath); err != nil {
			pipeWriter.CloseWithError(err)
		}
	}()

	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(c.endpoint, "/")+"/"+fmt.Sprintf(".api/files/batch-changes/%s", batchSpecID), pipeReader)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "token "+c.token)
	req.Header.Set("Content-Type", multipartWriter.FormDataContentType())

	resp, err := c.client.Do(req)
	if err != nil {
		// Errors passed to pipeWriter.CloseWithError come through here.
		return err
	}
	defer resp.Body.Close()

	// 2xx and 3xx are ok
	p, err := io.ReadAll(resp.Body)
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusBadRequest {
		if err != nil {
			return err
		}
		return errors.New(string(p))
	}
	return nil
}

func createFormFile(w *multipart.Writer, workingDir string, mountPath string) error {
	f, err := os.Open(filepath.Join(workingDir, mountPath))
	if err != nil {
		return err
	}
	defer f.Close()

	filePath, fileName := filepath.Split(mountPath)
	filePath = strings.Trim(strings.TrimSuffix(filePath, string(filepath.Separator)), ".")
	// Ensure Windows separators are changed to Unix.
	filePath = strings.ReplaceAll(filePath, "\\", "/")
	if err = w.WriteField("filepath", filePath); err != nil {
		return err
	}
	fileInfo, err := f.Stat()
	if err != nil {
		return err
	}
	if err = w.WriteField("filemod", fileInfo.ModTime().UTC().String()); err != nil {
		return err
	}

	part, err := w.CreateFormFile("file", fileName)
	if err != nil {
		return err
	}
	if _, err = io.Copy(part, f); err != nil {
		return err
	}
	return nil
}

package lsifstore

import (
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/log/logtest"

	codeintelshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestDeleteLsifDataByUploadIds(t *testing.T) {
	logger := logtest.ScopedWith(t, logtest.LoggerOptions{
		Level: log.LevelError,
	})
	codeIntelDB := codeintelshared.NewCodeIntelDB(logger, dbtest.NewDB(logger, t))
	store := New(&observation.TestContext, codeIntelDB)

	t.Run("scip", func(t *testing.T) {
		for i := 0; i < 5; i++ {
			query := sqlf.Sprintf("INSERT INTO codeintel_scip_metadata (upload_id, text_document_encoding, tooL_name, tool_version, tool_arguments, protocol_version) VALUES (%s, 'utf8', '', '', '{}', 1)", i+1)

			if _, err := codeIntelDB.ExecContext(context.Background(), query.Query(sqlf.PostgresBindVar), query.Args()...); err != nil {
				t.Fatalf("unexpected error inserting repo: %s", err)
			}
		}

		if err := store.DeleteLsifDataByUploadIds(context.Background(), 2, 4); err != nil {
			t.Fatalf("unexpected error clearing bundle data: %s", err)
		}

		dumpIDs, err := basestore.ScanInts(codeIntelDB.QueryContext(context.Background(), "SELECT upload_id FROM codeintel_scip_metadata"))
		if err != nil {
			t.Fatalf("Unexpected error querying dump identifiers: %s", err)
		}

		if diff := cmp.Diff([]int{1, 3, 5}, dumpIDs); diff != "" {
			t.Errorf("unexpected dump identifiers (-want +got):\n%s", diff)
		}
	})
}

func TestDeleteUnreferencedDocuments(t *testing.T) {
	logger := logtest.Scoped(t)
	codeIntelDB := codeintelshared.NewCodeIntelDB(logger, dbtest.NewDB(logger, t))
	internalStore := basestore.NewWithHandle(codeIntelDB.Handle())
	store := New(&observation.TestContext, codeIntelDB)

	for i := 0; i < 200; i++ {
		insertDocumentQuery := sqlf.Sprintf(
			`INSERT INTO codeintel_scip_documents(id, schema_version, payload_hash, raw_scip_payload) VALUES (%s, 1, %s, %s)`,
			i+1,
			fmt.Sprintf("hash-%d", i+1),
			fmt.Sprintf("payload-%d", i+1),
		)
		if err := internalStore.Exec(context.Background(), insertDocumentQuery); err != nil {
			t.Fatalf("unexpected error setting up database: %s", err)
		}
	}

	for i := 0; i < 200; i++ {
		insertDocumentLookupQuery := sqlf.Sprintf(
			`INSERT INTO codeintel_scip_document_lookup(upload_id, document_path, document_id) VALUES (%s, %s, %s)`,
			42,
			fmt.Sprintf("path-%d", i+1),
			i+1,
		)
		if err := internalStore.Exec(context.Background(), insertDocumentLookupQuery); err != nil {
			t.Fatalf("unexpected error setting up database: %s", err)
		}

		if i%3 == 0 {
			insertDocumentLookupQuery := sqlf.Sprintf(
				`INSERT INTO codeintel_scip_document_lookup(upload_id, document_path, document_id) VALUES (%s, %s, %s)`,
				43,
				fmt.Sprintf("path-%d", i+1),
				i+1,
			)
			if err := internalStore.Exec(context.Background(), insertDocumentLookupQuery); err != nil {
				t.Fatalf("unexpected error setting up database: %s", err)
			}
		}
	}

	deleteReferencesQuery := sqlf.Sprintf(`DELETE FROM codeintel_scip_document_lookup WHERE upload_id = 42`)
	if err := internalStore.Exec(context.Background(), deleteReferencesQuery); err != nil {
		t.Fatalf("unexpected error setting up database: %s", err)
	}

	// Check too soon
	_, count, err := store.DeleteUnreferencedDocuments(context.Background(), 20, time.Minute, time.Now())
	if err != nil {
		t.Fatalf("unexpected error deleting unreferenced documents: %s", err)
	}
	if count != 0 {
		t.Fatalf("did not expect any expired records, have %d", count)
	}

	// Consume actual records. We expect 10 batches (200 records deleted / 20 per batch) to be required to
	// process this workload.

	totalCount := 0
	for i := 0; i < 10; i++ {
		_, count, err = store.DeleteUnreferencedDocuments(context.Background(), 20, time.Minute, time.Now().Add(time.Minute*5))
		if err != nil {
			t.Fatalf("unexpected error deleting unreferenced documents: %s", err)
		}
		totalCount += count
	}
	if expected := 2 * 200 / 3; totalCount != expected {
		t.Fatalf("unexpected number of unreferenced documents deleted. want=%d have=%d", expected, totalCount)
	}

	// Assert no more records should be available for processing
	_, count, err = store.DeleteUnreferencedDocuments(context.Background(), 20, time.Minute, time.Now())
	if err != nil {
		t.Fatalf("unexpected error deleting unreferenced documents: %s", err)
	}
	if count != 0 {
		t.Fatalf("did not expect any unprocessed records, have %d", count)
	}

	documentIDsQuery := sqlf.Sprintf(`SELECT id FROM codeintel_scip_documents ORDER BY id`)
	ids, err := basestore.ScanInts(internalStore.Query(context.Background(), documentIDsQuery))
	if err != nil {
		t.Fatalf("unexpected error querying document ids: %s", err)
	}

	var expectedIDs []int
	for i := 0; i < 200; i++ {
		if i%3 == 0 {
			expectedIDs = append(expectedIDs, i+1)
		}
	}
	if diff := cmp.Diff(expectedIDs, ids); diff != "" {
		t.Fatalf("unexpected remaining document identifiers (-want +got):\n%s", diff)
	}
}

func TestIDsWithMeta(t *testing.T) {
	logger := logtest.Scoped(t)
	codeIntelDB := codeintelshared.NewCodeIntelDB(logger, dbtest.NewDB(logger, t))
	store := New(&observation.TestContext, codeIntelDB)
	ctx := context.Background()

	if _, err := codeIntelDB.ExecContext(ctx, `
		INSERT INTO codeintel_scip_metadata (upload_id, text_document_encoding, tool_name, tool_version, tool_arguments, protocol_version) VALUES (200, 'utf8', '', '', '{}', 1);
		INSERT INTO codeintel_scip_metadata (upload_id, text_document_encoding, tool_name, tool_version, tool_arguments, protocol_version) VALUES (202, 'utf8', '', '', '{}', 1);
		INSERT INTO codeintel_scip_metadata (upload_id, text_document_encoding, tool_name, tool_version, tool_arguments, protocol_version) VALUES (204, 'utf8', '', '', '{}', 1);
	`); err != nil {
		t.Fatalf("unexpected error setting up test: %s", err)
	}

	candidates := []int{
		200, // exists
		201,
		203,
		204, // exists
		205,
	}
	ids, err := store.IDsWithMeta(ctx, candidates)
	if err != nil {
		t.Fatalf("failed to find upload IDs with metadata: %s", err)
	}
	expectedIDs := []int{
		200,
		204,
	}
	sort.Ints(ids)
	if diff := cmp.Diff(expectedIDs, ids); diff != "" {
		t.Fatalf("unexpected IDs (-want +got):\n%s", diff)
	}
}

func TestReconcileCandidates(t *testing.T) {
	logger := logtest.Scoped(t)
	codeIntelDB := codeintelshared.NewCodeIntelDB(logger, dbtest.NewDB(logger, t))
	store := New(&observation.TestContext, codeIntelDB)

	ctx := context.Background()
	now := time.Unix(1587396557, 0).UTC()

	if _, err := codeIntelDB.ExecContext(ctx, `
		INSERT INTO codeintel_scip_metadata (upload_id, text_document_encoding, tool_name, tool_version, tool_arguments, protocol_version) VALUES (200, 'utf8', '', '', '{}', 1);
		INSERT INTO codeintel_scip_metadata (upload_id, text_document_encoding, tool_name, tool_version, tool_arguments, protocol_version) VALUES (201, 'utf8', '', '', '{}', 1);
		INSERT INTO codeintel_scip_metadata (upload_id, text_document_encoding, tool_name, tool_version, tool_arguments, protocol_version) VALUES (202, 'utf8', '', '', '{}', 1);
		INSERT INTO codeintel_scip_metadata (upload_id, text_document_encoding, tool_name, tool_version, tool_arguments, protocol_version) VALUES (203, 'utf8', '', '', '{}', 1);
		INSERT INTO codeintel_scip_metadata (upload_id, text_document_encoding, tool_name, tool_version, tool_arguments, protocol_version) VALUES (204, 'utf8', '', '', '{}', 1);
		INSERT INTO codeintel_scip_metadata (upload_id, text_document_encoding, tool_name, tool_version, tool_arguments, protocol_version) VALUES (205, 'utf8', '', '', '{}', 1);
	`); err != nil {
		t.Fatalf("unexpected error setting up test: %s", err)
	}

	// Initial batch of records
	ids, err := store.ReconcileCandidatesWithTime(ctx, 4, now)
	if err != nil {
		t.Fatalf("failed to get candidate IDs for reconciliation: %s", err)
	}
	expectedIDs := []int{
		200,
		201,
		202,
		203,
	}
	sort.Ints(ids)
	if diff := cmp.Diff(expectedIDs, ids); diff != "" {
		t.Fatalf("unexpected IDs (-want +got):\n%s", diff)
	}

	// Wraps around after exhausting first records
	ids, err = store.ReconcileCandidatesWithTime(ctx, 4, now.Add(time.Minute*1))
	if err != nil {
		t.Fatalf("failed to get candidate IDs for reconciliation: %s", err)
	}
	expectedIDs = []int{
		200,
		201,
		204,
		205,
	}
	sort.Ints(ids)
	if diff := cmp.Diff(expectedIDs, ids); diff != "" {
		t.Fatalf("unexpected IDs (-want +got):\n%s", diff)
	}

	// Continues to wrap around
	ids, err = store.ReconcileCandidatesWithTime(ctx, 2, now.Add(time.Minute*2))
	if err != nil {
		t.Fatalf("failed to get candidate IDs for reconciliation: %s", err)
	}
	expectedIDs = []int{
		202,
		203,
	}
	sort.Ints(ids)
	if diff := cmp.Diff(expectedIDs, ids); diff != "" {
		t.Fatalf("unexpected IDs (-want +got):\n%s", diff)
	}
}

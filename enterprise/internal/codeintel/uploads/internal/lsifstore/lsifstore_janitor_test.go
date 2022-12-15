package lsifstore

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/log/logtest"

	codeintelshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestDeleteUnreferencedDocuments(t *testing.T) {
	logger := logtest.Scoped(t)
	codeIntelDB := codeintelshared.NewCodeIntelDB(logger, dbtest.NewDB(logger, t))
	store := newStore(&observation.TestContext, codeIntelDB)

	for i := 0; i < 200; i++ {
		insertDocumentQuery := sqlf.Sprintf(
			`INSERT INTO codeintel_scip_documents(id, schema_version, payload_hash, raw_scip_payload) VALUES (%s, 1, %s, %s)`,
			i+1,
			fmt.Sprintf("hash-%d", i+1),
			fmt.Sprintf("payload-%d", i+1),
		)
		if err := store.db.Exec(context.Background(), insertDocumentQuery); err != nil {
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
		if err := store.db.Exec(context.Background(), insertDocumentLookupQuery); err != nil {
			t.Fatalf("unexpected error setting up database: %s", err)
		}

		if i%3 == 0 {
			insertDocumentLookupQuery := sqlf.Sprintf(
				`INSERT INTO codeintel_scip_document_lookup(upload_id, document_path, document_id) VALUES (%s, %s, %s)`,
				43,
				fmt.Sprintf("path-%d", i+1),
				i+1,
			)
			if err := store.db.Exec(context.Background(), insertDocumentLookupQuery); err != nil {
				t.Fatalf("unexpected error setting up database: %s", err)
			}
		}
	}

	deleteReferencesQuery := sqlf.Sprintf(`DELETE FROM codeintel_scip_document_lookup WHERE upload_id = 42`)
	if err := store.db.Exec(context.Background(), deleteReferencesQuery); err != nil {
		t.Fatalf("unexpected error setting up database: %s", err)
	}

	// Check too soon
	count, err := store.DeleteUnreferencedDocuments(context.Background(), 20, time.Minute, time.Now())
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
		count, err = store.DeleteUnreferencedDocuments(context.Background(), 20, time.Minute, time.Now().Add(time.Minute*5))
		if err != nil {
			t.Fatalf("unexpected error deleting unreferenced documents: %s", err)
		}
		totalCount += count
	}
	if expected := 2 * 200 / 3; totalCount != expected {
		t.Fatalf("unexpected number of unreferenced documents deleted. want=%d have=%d", expected, totalCount)
	}

	// Assert no more records should be available for processing
	count, err = store.DeleteUnreferencedDocuments(context.Background(), 20, time.Minute, time.Now())
	if err != nil {
		t.Fatalf("unexpected error deleting unreferenced documents: %s", err)
	}
	if count != 0 {
		t.Fatalf("did not expect any unprocessed records, have %d", count)
	}

	documentIDsQuery := sqlf.Sprintf(`SELECT id FROM codeintel_scip_documents ORDER BY id`)
	ids, err := basestore.ScanInts(store.db.Query(context.Background(), documentIDsQuery))
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

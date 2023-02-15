package lsifstore

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log/logtest"

	codeintelshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

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
	store := newStore(&observation.TestContext, codeIntelDB)

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
	ids, err := store.reconcileCandidates(ctx, 4, now)
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
	ids, err = store.reconcileCandidates(ctx, 4, now.Add(time.Minute*1))
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
	ids, err = store.reconcileCandidates(ctx, 2, now.Add(time.Minute*2))
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

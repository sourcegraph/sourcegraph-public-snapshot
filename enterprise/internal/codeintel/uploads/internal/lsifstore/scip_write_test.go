package lsifstore

import (
	"context"
	"testing"

	"github.com/sourcegraph/log/logtest"

	codeintelshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestInsertMetadata(t *testing.T) {
	logger := logtest.Scoped(t)
	codeIntelDB := codeintelshared.NewCodeIntelDB(dbtest.NewDB(logger, t), logger)
	store := New(codeIntelDB, &observation.TestContext)
	ctx := context.Background()

	if err := store.InsertMetadata(ctx, 42, ProcessedMetadata{
		TextDocumentEncoding: "UTF8",
		ToolName:             "scip-test",
		ToolVersion:          "0.1.0",
		ToolArguments:        []string{"-p", "src"},
		ProtocolVersion:      1,
	}); err != nil {
		t.Fatalf("failed to insert metadata: %s", err)
	}
}

func TestInsertSCIPDocument(t *testing.T) {
	logger := logtest.Scoped(t)
	codeIntelDB := codeintelshared.NewCodeIntelDB(dbtest.NewDB(logger, t), logger)
	store := New(codeIntelDB, &observation.TestContext)
	ctx := context.Background()

	if _, err := store.InsertSCIPDocument(
		ctx,
		24,
		"internal/util.go",
		[]byte("deadbeef"),
		[]byte("lorem ipsum dolor sit amet"),
	); err != nil {
		t.Fatalf("failed to write SCIP document: %s", err)
	}

	if _, err := store.InsertSCIPDocument(
		ctx,
		25,
		"internal/util.go",
		[]byte("deadbeef"),
		[]byte("lorem ipsum dolor sit amet"),
	); err != nil {
		t.Fatalf("failed to write SCIP document: %s", err)
	}

	if _, err := store.InsertSCIPDocument(
		ctx,
		25,
		"internal/util_test.go",
		[]byte("cafebabe"),
		[]byte("consectetur adipiscing elit, sed do eiusmod"),
	); err != nil {
		t.Fatalf("failed to write SCIP document: %s", err)
	}

	count, _, err := basestore.ScanFirstInt(codeIntelDB.Handle().QueryContext(ctx, `SELECT COUNT(*) FROM codeintel_scip_documents`))
	if err != nil {
		t.Fatalf("failed to query number of SCIP documents: %s", err)
	} else if expected := 2; count != expected {
		t.Fatalf("unexpected number of documents. want=%d have=%d", expected, count)
	}
}

func TestWriteSCIPSymbols(t *testing.T) {
	logger := logtest.Scoped(t)
	codeIntelDB := codeintelshared.NewCodeIntelDB(dbtest.NewDB(logger, t), logger)
	store := New(codeIntelDB, &observation.TestContext)
	ctx := context.Background()

	uploadID := 24

	documentLookupID, err := store.InsertSCIPDocument(
		ctx,
		uploadID,
		"internal/util.go",
		[]byte("deadbeef"),
		[]byte("lorem ipsum dolor sit amet"),
	)
	if err != nil {
		t.Fatalf("failed to write SCIP document: %s", err)
	}

	symbols := []types.InvertedRangeIndex{
		{
			SymbolName: "foo.bar.ident",
			DefinitionRanges: []int32{
				3, 25, 3, 30,
			},
			ReferenceRanges: []int32{
				4, 25, 4, 30,
				5, 10, 5, 15,
				5, 25, 5, 30,
				6, 16, 6, 21,
			},
		},
		{
			SymbolName: "bar.baz.longerName",
			ReferenceRanges: []int32{
				100, 10, 100, 20,
				101, 15, 101, 25,
				103, 16, 103, 26,
				103, 31, 103, 41,
				103, 55, 103, 65,
				151, 10, 151, 20,
				152, 15, 152, 25,
				154, 25, 154, 35,
				154, 50, 154, 60,
			},
			ImplementationRanges: []int32{
				342, 5, 342, 15,
				364, 5, 364, 15,
			},
		},
		{
			SymbolName: "baz.bonk.quux",
			DefinitionRanges: []int32{
				251, 24, 251, 30,
			},
			TypeDefinitionRanges: []int32{
				151, 14, 151, 20,
			},
		},
	}

	tx, err := store.Transact(ctx)
	if err != nil {
		t.Fatalf("failed to start transaction: %s", err)
	}
	defer func() { _ = tx.Done(nil) }()

	symbolWriter, err := tx.NewSymbolWriter(ctx, uploadID)
	if err != nil {
		t.Fatalf("failed to write SCIP symbols: %s", err)
	}
	if err := symbolWriter.WriteSCIPSymbols(ctx, documentLookupID, symbols); err != nil {
		t.Fatalf("failed to write SCIP symbols: %s", err)
	}

	if n, err := symbolWriter.Flush(ctx); err != nil {
		t.Fatalf("failed to write SCIP symbols: %s", err)
	} else if expected := uint32(3); n != expected {
		t.Fatalf("unexpected number of symbols inserted. want=%d have=%d", expected, n)
	}
}

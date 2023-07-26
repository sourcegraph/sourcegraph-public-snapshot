package lsif

import (
	"context"
	"os"
	"testing"

	"github.com/sourcegraph/log/logtest"

	stores "github.com/sourcegraph/sourcegraph/internal/codeintel/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
)

func init() {
	scipMigratorUploadReaderBatchSize = 1
	scipMigratorDocumentReaderBatchSize = 4
	scipMigratorResultChunkReaderCacheSize = 16
}

func TestSCIPMigrator(t *testing.T) {
	logger := logtest.Scoped(t)
	rawDB := lastDBWithLSIF(logger, t)
	db := database.NewDB(logger, rawDB)
	codeIntelDB := stores.NewCodeIntelDB(logger, rawDB)
	store := basestore.NewWithHandle(db.Handle())
	codeIntelStore := basestore.NewWithHandle(codeIntelDB.Handle())
	migrator := NewSCIPMigrator(store, codeIntelStore)
	ctx := context.Background()

	contents, err := os.ReadFile("./testdata/lsif.sql")
	if err != nil {
		t.Fatalf("unexpected error reading file: %s", err)
	}
	if _, err := codeIntelDB.ExecContext(ctx, string(contents)); err != nil {
		t.Fatalf("unexpected error executing test file: %s", err)
	}

	assertProgress := func(expectedProgress float64, applyReverse bool) {
		if progress, err := migrator.Progress(context.Background(), applyReverse); err != nil {
			t.Fatalf("unexpected error querying progress: %s", err)
		} else if progress != expectedProgress {
			t.Errorf("unexpected progress. want=%.2f have=%.2f", expectedProgress, progress)
		}
	}

	// Initial state
	assertProgress(0, false)

	// Migrate first upload record
	if err := migrator.Up(context.Background()); err != nil {
		t.Fatalf("unexpected error performing up migration: %s", err)
	}
	assertProgress(0.5, false)

	// Migrate second upload record
	if err := migrator.Up(context.Background()); err != nil {
		t.Fatalf("unexpected error performing up migration: %s", err)
	}
	assertProgress(1, false)

	// Assert no-op downwards progress
	assertProgress(0, true)

	// Assert migrated state
	documentsCount, _, err := basestore.ScanFirstInt(codeIntelDB.QueryContext(ctx, `SELECT COUNT(*) FROM codeintel_scip_documents`))
	if err != nil {
		t.Fatalf("unexpected error counting documents: %s", err)
	}
	if expected := 59; documentsCount != expected {
		t.Fatalf("unexpected number of documents. want=%d have=%d", expected, documentsCount)
	}
	symbolsCount, _, err := basestore.ScanFirstInt(codeIntelDB.QueryContext(ctx, `SELECT COUNT(*) FROM codeintel_scip_symbols`))
	if err != nil {
		t.Fatalf("unexpected error counting symbols: %s", err)
	}
	if expected := 4221; symbolsCount != expected {
		t.Fatalf("unexpected number of documents. want=%d have=%d", expected, symbolsCount)
	}
}

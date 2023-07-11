package scip

import (
	"context"
	"os"
	"testing"

	"github.com/sourcegraph/log/logtest"

	stores "github.com/sourcegraph/sourcegraph/internal/codeintel/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

func TestSymbolsMigratorUp(t *testing.T) {
	logger := logtest.Scoped(t)
	rawDB := dbtest.NewDB(logger, t)
	codeIntelDB := stores.NewCodeIntelDB(logger, rawDB)
	codeIntelStore := basestore.NewWithHandle(codeIntelDB.Handle())
	migrator := NewSCIPSymbolsMigrator(codeIntelStore)
	ctx := context.Background()

	contents, err := os.ReadFile("./testdata/trie.sql")
	if err != nil {
		t.Fatalf("unexpected error reading file: %s", err)
	}
	if _, err := codeIntelDB.ExecContext(ctx, string(contents)); err != nil {
		t.Fatalf("unexpected error executing test file: %s", err)
	}

	assertProgress := func(expectedProgress float64, applyReverse bool) {
		if progress, err := migrator.Progress(ctx, applyReverse); err != nil {
			t.Fatalf("unexpected error querying progress: %s", err)
		} else if progress != expectedProgress {
			t.Errorf("unexpected progress. want=%.2f have=%.2f", expectedProgress, progress)
		}
	}

	// Initial state
	assertProgress(0, false)

	// Migrate first upload record
	if err := migrator.Up(ctx); err != nil {
		t.Fatalf("unexpected error performing up migration: %s", err)
	}
	assertProgress(0.5, false)

	// Migrate second upload record
	if err := migrator.Up(ctx); err != nil {
		t.Fatalf("unexpected error performing up migration: %s", err)
	}
	assertProgress(1, false)

	// Assert migrated state
	// TODO
	t.Fatalf("unimplemented")
}

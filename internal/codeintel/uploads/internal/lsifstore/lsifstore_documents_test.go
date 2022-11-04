package lsifstore

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestGetUploadDocumentsForPath(t *testing.T) {
	store := populateTestStore(t)

	if paths, count, err := store.GetUploadDocumentsForPath(context.Background(), testBundleID, "%%"); err != nil {
		t.Fatalf("unexpected error %s", err)
	} else if count != 7 || len(paths) != 7 {
		t.Errorf("expected %d document paths but got none: count=%d len=%d", 7, count, len(paths))
	} else {
		expected := []string{
			"cmd/lsif-go/main.go",
			"internal/gomod/module.go",
			"internal/index/helper.go",
			"internal/index/indexer.go",
			"internal/index/types.go",
			"protocol/protocol.go",
			"protocol/writer.go",
		}

		if diff := cmp.Diff(expected, paths); diff != "" {
			t.Errorf("unexpected document paths (-want +got):\n%s", diff)
		}
	}
}

const testBundleID = 1

func populateTestStore(t testing.TB) LsifStore {
	logger := logtest.Scoped(t)
	sqlDB := dbtest.NewDB(logger, t)
	db := database.NewDB(logger, sqlDB)
	store := New(db, &observation.TestContext)

	contents, err := os.ReadFile("./testdata/lsif-go@ad3507cb.sql")
	if err != nil {
		t.Fatalf("unexpected error reading testdata: %s", err)
	}

	tx, err := db.Transact(context.Background())
	if err != nil {
		t.Fatalf("unexpected error starting transaction: %s", err)
	}
	defer func() {
		if err := tx.Done(nil); err != nil {
			t.Fatalf("unexpected error finishing transaction: %s", err)
		}
	}()

	// Remove comments from the lines.
	var withoutComments []byte
	for _, line := range bytes.Split(contents, []byte{'\n'}) {
		if string(line) == "" || bytes.HasPrefix(line, []byte("--")) {
			continue
		}
		withoutComments = append(withoutComments, line...)
		withoutComments = append(withoutComments, '\n')
	}

	// Execute each statement. Split on ";\n" because statements may have e.g. string literals that
	// span multiple lines.
	for _, statement := range strings.Split(string(withoutComments), ";\n") {
		if strings.Contains(statement, "_schema_versions") {
			// Statements which insert into lsif_data_*_schema_versions should not be executed, as
			// these are already inserted during regular DB up migrations.
			continue
		}
		if _, err := tx.ExecContext(context.Background(), statement); err != nil {
			t.Fatalf("unexpected error loading database data: %s", err)
		}
	}

	return store
}

package lsifstore

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log/logtest"

	codeintelshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestGetUploadDocumentsForPath(t *testing.T) {
	store := populateTestStore(t)

	t.Run("scip", func(t *testing.T) {
		if paths, count, err := store.GetUploadDocumentsForPath(context.Background(), scipTestBundleID, "template/src/util/%%"); err != nil {
			t.Fatalf("unexpected error %s", err)
		} else if expected := 8; count != expected || len(paths) != expected {
			t.Errorf("unexpected number of paths: want=%d have=%d (%d total)", expected, len(paths), count)
		} else {
			expected := []string{
				"template/src/util/api.ts",
				"template/src/util/graphql.ts",
				"template/src/util/helpers.ts",
				"template/src/util/ix.test.ts",
				"template/src/util/ix.ts",
				"template/src/util/promise.ts",
				"template/src/util/uri.test.ts",
				"template/src/util/uri.ts",
			}

			if diff := cmp.Diff(expected, paths); diff != "" {
				t.Errorf("unexpected document paths (-want +got):\n%s", diff)
			}
		}
	})
}

const (
	scipTestBundleID = 2408562
)

func populateTestStore(t testing.TB) LsifStore {
	logger := logtest.Scoped(t)
	codeIntelDB := codeintelshared.NewCodeIntelDB(logger, dbtest.NewDB(logger, t))
	store := New(&observation.TestContext, codeIntelDB)

	loadTestFile(t, codeIntelDB, "./testdata/code-intel-extensions@7802976b.sql")
	return store
}

func loadTestFile(t testing.TB, codeIntelDB codeintelshared.CodeIntelDB, filepath string) {
	contents, err := os.ReadFile(filepath)
	if err != nil {
		t.Fatalf("unexpected error reading testdata from %q: %s", filepath, err)
	}

	tx, err := codeIntelDB.Transact(context.Background())
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
}

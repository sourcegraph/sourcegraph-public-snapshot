package definition

import (
	"io/fs"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/definition/testdata"
)

func TestReadDefinitions(t *testing.T) {
	t.Run("well-formed", func(t *testing.T) {
		fs, err := fs.Sub(testdata.Content, "well-formed")
		if err != nil {
			t.Fatalf("unexpected error fetching schema %q: %s", "well-formed", err)
		}

		definitions, err := ReadDefinitions(fs)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		expectedDefinitions := []Definition{
			{ID: 10001, UpQuery: sqlf.Sprintf("10001 UP"), DownQuery: sqlf.Sprintf("10001 DOWN")},
			{ID: 10002, UpQuery: sqlf.Sprintf("10002 UP"), DownQuery: sqlf.Sprintf("10002 DOWN"), Parents: []int{10001}},
			{ID: 10003, UpQuery: sqlf.Sprintf("10003 UP"), DownQuery: sqlf.Sprintf("10003 DOWN"), Parents: []int{10002}},
			{ID: 10004, UpQuery: sqlf.Sprintf("10004 UP"), DownQuery: sqlf.Sprintf("10004 DOWN"), Parents: []int{10003}},
			{ID: 10005, UpQuery: sqlf.Sprintf("10005 UP"), DownQuery: sqlf.Sprintf("10005 DOWN"), Parents: []int{10004}},
		}
		if diff := cmp.Diff(expectedDefinitions, definitions.definitions, queryComparer); diff != "" {
			t.Fatalf("unexpected definitions (-want +got):\n%s", diff)
		}
	})

	t.Run("concurrent", func(t *testing.T) {
		fs, err := fs.Sub(testdata.Content, "concurrent")
		if err != nil {
			t.Fatalf("unexpected error fetching schema %q: %s", "concurrent", err)
		}

		definitions, err := ReadDefinitions(fs)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		expectedDefinitions := []Definition{
			{ID: 10001, UpQuery: sqlf.Sprintf("10001 UP"), DownQuery: sqlf.Sprintf("10001 DOWN"), Parents: nil}, // first
			{
				ID:                        10002,
				UpQuery:                   sqlf.Sprintf("-- Some docs here\nCREATE INDEX CONCURRENTLY IF NOT EXISTS idx ON tbl(col1, col2, col3);"),
				DownQuery:                 sqlf.Sprintf("DROP INDEX IF EXISTS idx;"),
				IsCreateIndexConcurrently: true,
				IndexMetadata: &IndexMetadata{
					TableName: "tbl",
					IndexName: "idx",
				},
				Parents: []int{10001},
			}, // second
		}
		if diff := cmp.Diff(expectedDefinitions, definitions.definitions, queryComparer); diff != "" {
			t.Fatalf("unexpected definitions (-want +got):\n%s", diff)
		}
	})

	t.Run("missing metadata", func(t *testing.T) { testReadDefinitionsError(t, "missing-metadata", "malformed") })
	t.Run("missing upgrade query", func(t *testing.T) { testReadDefinitionsError(t, "missing-upgrade-query", "malformed") })
	t.Run("missing downgrade query", func(t *testing.T) { testReadDefinitionsError(t, "missing-downgrade-query", "malformed") })
	t.Run("no roots", func(t *testing.T) { testReadDefinitionsError(t, "no-roots", "no roots") })
	t.Run("multiple roots", func(t *testing.T) { testReadDefinitionsError(t, "multiple-roots", "multiple roots") })
	t.Run("unknown parent", func(t *testing.T) { testReadDefinitionsError(t, "unknown-parent", "unknown migration") })

	errConcurrentUnexpected := "did not expect up migration to contain concurrent creation of an index"
	errConcurrentExpected := "expected up migration to contain concurrent creation of an index"
	errConcurrentDown := "did not expect down migration to contain concurrent creation of an index"

	t.Run("unexpected concurrent index creation", func(t *testing.T) { testReadDefinitionsError(t, "concurrent-unexpected", errConcurrentUnexpected) })
	t.Run("missing concurrent index creation", func(t *testing.T) { testReadDefinitionsError(t, "concurrent-expected", errConcurrentExpected) })
	t.Run("concurrent index creation down", func(t *testing.T) { testReadDefinitionsError(t, "concurrent-down", errConcurrentDown) })
}

func testReadDefinitionsError(t *testing.T, name, expectedError string) {
	t.Helper()

	fs, err := fs.Sub(testdata.Content, name)
	if err != nil {
		t.Fatalf("unexpected error fetching schema %q: %s", name, err)
	}

	if _, err := ReadDefinitions(fs); err == nil || !strings.Contains(err.Error(), expectedError) {
		t.Fatalf("unexpected error. want=%q got=%q", expectedError, err)
	}
}

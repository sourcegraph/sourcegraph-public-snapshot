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
	queryComparer := cmp.Comparer(func(a, b *sqlf.Query) bool {
		if a == nil {
			return b == nil
		}
		if b == nil {
			return false
		}
		return strings.TrimSpace(a.Query(sqlf.PostgresBindVar)) == strings.TrimSpace(b.Query(sqlf.PostgresBindVar))
	})

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
			{ID: 10001, UpFilename: "10001_first.up.sql", DownFilename: "10001_first.down.sql", UpQuery: sqlf.Sprintf("10001 UP"), DownQuery: sqlf.Sprintf("10001 DOWN")},
			{ID: 10002, UpFilename: "10002_second.up.sql", DownFilename: "10002_second.down.sql", UpQuery: sqlf.Sprintf("10002 UP"), DownQuery: sqlf.Sprintf("10002 DOWN"), Metadata: Metadata{Parent: 10001}},
			{ID: 10003, UpFilename: "10003_third.up.sql", DownFilename: "10003_third.down.sql", UpQuery: sqlf.Sprintf("10003 UP"), DownQuery: sqlf.Sprintf("10003 DOWN"), Metadata: Metadata{Parent: 10002}},
			{ID: 10004, UpFilename: "10004_fourth.up.sql", DownFilename: "10004_fourth.down.sql", UpQuery: sqlf.Sprintf("10004 UP"), DownQuery: sqlf.Sprintf("10004 DOWN"), Metadata: Metadata{Parent: 10003}},
			{ID: 10005, UpFilename: "10005_fifth.up.sql", DownFilename: "10005_fifth.down.sql", UpQuery: sqlf.Sprintf("10005 UP"), DownQuery: sqlf.Sprintf("10005 DOWN"), Metadata: Metadata{Parent: 10004}},
		}
		if diff := cmp.Diff(expectedDefinitions, definitions.definitions, queryComparer); diff != "" {
			t.Fatalf("unexpected definitions (-want +got):\n%s", diff)
		}
	})

	t.Run("missing upgrade query", func(t *testing.T) {
		testReadDefinitionsError(t, "missing-upgrade-query", "not found")
	})

	t.Run("missing downgrade query", func(t *testing.T) {
		testReadDefinitionsError(t, "missing-downgrade-query", "not found")
	})

	t.Run("duplicate upgrade query", func(t *testing.T) {
		testReadDefinitionsError(t, "duplicate-upgrade-query", "duplicate upgrade query")
	})

	t.Run("duplicate downgrade query", func(t *testing.T) {
		testReadDefinitionsError(t, "duplicate-downgrade-query", "duplicate downgrade query")
	})

	t.Run("gap in sequence", func(t *testing.T) {
		testReadDefinitionsError(t, "gap-in-sequence", "migration identifiers jump")
	})

	t.Run("root-with-parent", func(t *testing.T) {
		testReadDefinitionsError(t, "root-with-parent", "no roots")
	})

	t.Run("unexpected-parent", func(t *testing.T) {
		testReadDefinitionsError(t, "unexpected-parent", "cycle")
	})
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

package definition

import (
	"fmt"
	"io/fs"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/definition/testdata"
)

const relativeWorkingDirectory = "internal/database/migration/definition"

func TestReadDefinitions(t *testing.T) {
	t.Run("well-formed", func(t *testing.T) {
		fsys, err := fs.Sub(testdata.Content, "well-formed")
		if err != nil {
			t.Fatalf("unexpected error fetching schema %q: %s", "well-formed", err)
		}

		definitions, err := ReadDefinitions(fsys, relativeWorkingDirectory)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		expectedDefinitions := []Definition{
			{ID: 10001, Name: "first", UpQuery: sqlf.Sprintf("10001 UP"), DownQuery: sqlf.Sprintf("10001 DOWN"), Parents: nil},
			{ID: 10002, Name: "second", UpQuery: sqlf.Sprintf("10002 UP"), DownQuery: sqlf.Sprintf("10002 DOWN"), Parents: []int{10001}},
			{ID: 10003, Name: "third or fourth (1)", UpQuery: sqlf.Sprintf("10003 UP"), DownQuery: sqlf.Sprintf("10003 DOWN"), Parents: []int{10002}},
			{ID: 10004, Name: "third or fourth (2)", UpQuery: sqlf.Sprintf("10004 UP"), DownQuery: sqlf.Sprintf("10004 DOWN"), Parents: []int{10002}},
			{ID: 10005, Name: "fifth", UpQuery: sqlf.Sprintf("10005 UP"), DownQuery: sqlf.Sprintf("10005 DOWN"), Parents: []int{10003, 10004}},
			{ID: 10006, Name: "do the thing", UpQuery: sqlf.Sprintf("10006 UP"), DownQuery: sqlf.Sprintf("10006 DOWN"), Parents: []int{10005}},
		}
		if diff := cmp.Diff(expectedDefinitions, definitions.definitions, queryComparer); diff != "" {
			t.Fatalf("unexpected definitions (-want +got):\n%s", diff)
		}
	})

	t.Run("concurrent", func(t *testing.T) {
		fsys, err := fs.Sub(testdata.Content, "concurrent")
		if err != nil {
			t.Fatalf("unexpected error fetching schema %q: %s", "concurrent", err)
		}

		definitions, err := ReadDefinitions(fsys, relativeWorkingDirectory)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		expectedDefinitions := []Definition{
			{
				ID:        10001,
				Name:      "first",
				UpQuery:   sqlf.Sprintf("10001 UP"),
				DownQuery: sqlf.Sprintf("10001 DOWN"),
			},
			{
				ID:                        10002,
				Name:                      "second",
				UpQuery:                   sqlf.Sprintf("-- Some docs here\nCREATE INDEX CONCURRENTLY IF NOT EXISTS idx ON tbl(col1, col2, col3);"),
				DownQuery:                 sqlf.Sprintf("DROP INDEX IF EXISTS idx;"),
				IsCreateIndexConcurrently: true,
				IndexMetadata: &IndexMetadata{
					TableName: "tbl",
					IndexName: "idx",
				},
				Parents: []int{10001},
			},
		}
		if diff := cmp.Diff(expectedDefinitions, definitions.definitions, queryComparer); diff != "" {
			t.Fatalf("unexpected definitions (-want +got):\n%s", diff)
		}
	})

	t.Run("concurrent unique", func(t *testing.T) {
		fsys, err := fs.Sub(testdata.Content, "concurrent-unique")
		if err != nil {
			t.Fatalf("unexpected error fetching schema %q: %s", "concurrent", err)
		}

		definitions, err := ReadDefinitions(fsys, relativeWorkingDirectory)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		expectedDefinitions := []Definition{
			{
				ID:        10001,
				Name:      "first",
				UpQuery:   sqlf.Sprintf("10001 UP"),
				DownQuery: sqlf.Sprintf("10001 DOWN"),
			},
			{
				ID:                        10002,
				Name:                      "second",
				UpQuery:                   sqlf.Sprintf("-- Some docs here\nCREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS idx ON tbl(col1, col2, col3);"),
				DownQuery:                 sqlf.Sprintf("DROP INDEX IF EXISTS idx;"),
				IsCreateIndexConcurrently: true,
				IndexMetadata: &IndexMetadata{
					TableName: "tbl",
					IndexName: "idx",
				},
				Parents: []int{10001},
			},
		}
		if diff := cmp.Diff(expectedDefinitions, definitions.definitions, queryComparer); diff != "" {
			t.Fatalf("unexpected definitions (-want +got):\n%s", diff)
		}
	})

	t.Run("privileged", func(t *testing.T) {
		fsys, err := fs.Sub(testdata.Content, "privileged")
		if err != nil {
			t.Fatalf("unexpected error fetching schema %q: %s", "privileged", err)
		}

		definitions, err := ReadDefinitions(fsys, relativeWorkingDirectory)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		expectedDefinitions := []Definition{
			{
				ID:         10001,
				Name:       "first",
				UpQuery:    sqlf.Sprintf("CREATE EXTENSION IF NOT EXISTS citext;"),
				DownQuery:  sqlf.Sprintf("DROP EXTENSION IF EXISTS citext;"),
				Privileged: true,
			},
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
	t.Run("cycle (connected to root)", func(t *testing.T) { testReadDefinitionsError(t, "cycle-traversal", "cycle") })
	t.Run("cycle (disconnected from root)", func(t *testing.T) { testReadDefinitionsError(t, "cycle-size", "cycle") })
	t.Run("unknown parent", func(t *testing.T) { testReadDefinitionsError(t, "unknown-parent", "unknown migration") })

	errConcurrentUnexpected := fmt.Sprintf("did not expect up query of migration at '%s/10002' to contain concurrent creation of an index", relativeWorkingDirectory)
	errConcurrentExpected := fmt.Sprintf("expected up query of migration at '%s/10002' to contain concurrent creation of an index", relativeWorkingDirectory)
	errConcurrentExtra := fmt.Sprintf(" did not expect up query of migration at '%s/10002' to contain additional statements", relativeWorkingDirectory)
	errConcurrentDown := fmt.Sprintf("did not expect down query of migration at '%s/10002' to contain concurrent creation of an index", relativeWorkingDirectory)
	errUnmarkedPrivilege := fmt.Sprintf("did not expect queries of migration at '%s/10001' to require elevated permissions", relativeWorkingDirectory)

	t.Run("unexpected concurrent index creation", func(t *testing.T) { testReadDefinitionsError(t, "concurrent-unexpected", errConcurrentUnexpected) })
	t.Run("missing concurrent index creation", func(t *testing.T) { testReadDefinitionsError(t, "concurrent-expected", errConcurrentExpected) })
	t.Run("non-isolated concurrent index creation", func(t *testing.T) { testReadDefinitionsError(t, "concurrent-extra", errConcurrentExtra) })
	t.Run("concurrent index creation down", func(t *testing.T) { testReadDefinitionsError(t, "concurrent-down", errConcurrentDown) })

	t.Run("unmarked privilege", func(t *testing.T) { testReadDefinitionsError(t, "unmarked-privilege", errUnmarkedPrivilege) })
}

func testReadDefinitionsError(t *testing.T, name, expectedError string) {
	t.Helper()

	fsys, err := fs.Sub(testdata.Content, name)
	if err != nil {
		t.Fatalf("unexpected error fetching schema %q: %s", name, err)
	}

	if _, err := ReadDefinitions(fsys, relativeWorkingDirectory); err == nil || !strings.Contains(err.Error(), expectedError) {
		t.Fatalf("unexpected error. want=%q got=%q", expectedError, err)
	}
}

var testFrontmatter = `
-- +++
parent: 12345
-- +++
`

func TestCanonicalizeQuery(t *testing.T) {
	for _, testCase := range []struct {
		name     string
		input    string
		expected string
	}{
		{"noop", "MY QUERY;", "MY QUERY;"},
		{"whitespace", "  MY QUERY;  ", "MY QUERY;"},
		{"yaml frontmatter", testFrontmatter + "\n\nMY QUERY;\n", "MY QUERY;"},
		{"kitchen sink", "BEGIN;\n\nMY QUERY;\n\nCOMMIT;\n", "MY QUERY;"},
		{"transactions", testFrontmatter + "\n\nMY QUERY;\n", "MY QUERY;"},
		{"kitchen sink", testFrontmatter + "\n\nBEGIN;\n\nMY QUERY;\n\nCOMMIT;\n", "MY QUERY;"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			if query := CanonicalizeQuery(testCase.input); query != testCase.expected {
				t.Errorf("unexpected canonical query. want=%q have=%q", testCase.expected, query)
			}
		})
	}
}

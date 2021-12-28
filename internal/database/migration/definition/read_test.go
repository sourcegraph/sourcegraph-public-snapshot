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

		type comparableDefinition struct {
			ID           int
			UpFilename   string
			UpQuery      string
			DownFilename string
			DownQuery    string
		}

		comparableDefinitions := make([]comparableDefinition, 0, len(definitions.definitions))
		for _, definition := range definitions.definitions {
			comparableDefinitions = append(comparableDefinitions, comparableDefinition{
				ID:           definition.ID,
				UpFilename:   definition.UpFilename,
				UpQuery:      strings.TrimSpace(definition.UpQuery.Query(sqlf.PostgresBindVar)),
				DownFilename: definition.DownFilename,
				DownQuery:    strings.TrimSpace(definition.DownQuery.Query(sqlf.PostgresBindVar)),
			})
		}

		expectedDefinitions := []comparableDefinition{
			{ID: 10001, UpFilename: "10001_first.up.sql", DownFilename: "10001_first.down.sql", UpQuery: "10001 UP", DownQuery: "10001 DOWN"},
			{ID: 10002, UpFilename: "10002_second.up.sql", DownFilename: "10002_second.down.sql", UpQuery: "10002 UP", DownQuery: "10002 DOWN"},
			{ID: 10003, UpFilename: "10003_third.up.sql", DownFilename: "10003_third.down.sql", UpQuery: "10003 UP", DownQuery: "10003 DOWN"},
			{ID: 10004, UpFilename: "10004_fourth.up.sql", DownFilename: "10004_fourth.down.sql", UpQuery: "10004 UP", DownQuery: "10004 DOWN"},
			{ID: 10005, UpFilename: "10005_fifth.up.sql", DownFilename: "10005_fifth.down.sql", UpQuery: "10005 UP", DownQuery: "10005 DOWN"},
		}
		if diff := cmp.Diff(expectedDefinitions, comparableDefinitions); diff != "" {
			t.Fatalf("unexpected definitions (-want +got):\n%s", diff)
		}
	})

	t.Run("missing upgrade query", func(t *testing.T) { testReadDefinitionsError(t, "missing-upgrade-query") })
	t.Run("missing downgrade query", func(t *testing.T) { testReadDefinitionsError(t, "missing-downgrade-query") })
	t.Run("duplicate upgrade query", func(t *testing.T) { testReadDefinitionsError(t, "duplicate-upgrade-query") })
	t.Run("duplicate downgrade query", func(t *testing.T) { testReadDefinitionsError(t, "duplicate-downgrade-query") })
	t.Run("gap in sequence", func(t *testing.T) { testReadDefinitionsError(t, "gap-in-sequence") })
}

func testReadDefinitionsError(t *testing.T, name string) {
	t.Helper()

	fs, err := fs.Sub(testdata.Content, name)
	if err != nil {
		t.Fatalf("unexpected error fetching schema %q: %s", name, err)
	}

	if _, err := ReadDefinitions(fs); err == nil {
		t.Fatalf("expected error")
	}
}

package store

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestRepositoryExceptions(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))
	store := New(&observation.TestContext, db)

	query := sqlf.Sprintf(
		`INSERT INTO repo (id, name) VALUES (%s, %s)`,
		42,
		"github.com/baz/honk",
	)
	if _, err := db.ExecContext(context.Background(), query.Query(sqlf.PostgresBindVar), query.Args()...); err != nil {
		t.Fatalf("unexpected error inserting repo: %s", err)
	}

	for _, testCase := range []struct {
		canSchedule bool
		canInfer    bool
	}{
		{true, false},
		{false, true},
		{false, false},
		{true, true},
	} {
		if err := store.SetRepositoryExceptions(context.Background(), 42, testCase.canSchedule, testCase.canInfer); err != nil {
			t.Fatalf("failed to update repository exception: %s", err)
		}

		canSchedule, canInfer, err := store.RepositoryExceptions(context.Background(), 42)
		if err != nil {
			t.Fatalf("unexpected error getting repository exceptions: %s", err)
		}
		if canSchedule != testCase.canSchedule {
			t.Errorf("unexpected exception for can_schedule. want=%v have=%v", testCase.canSchedule, canSchedule)
		}
		if canInfer != testCase.canInfer {
			t.Errorf("unexpected exception for can_infer. want=%v have=%v", testCase.canInfer, canInfer)
		}
	}
}

func TestGetIndexConfigurationByRepositoryID(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))
	store := New(&observation.TestContext, db)

	expectedConfigurationData := []byte(`{
		"foo": "bar",
		"baz": "bonk",
	}`)

	query := sqlf.Sprintf(
		`INSERT INTO repo (id, name) VALUES (%s, %s)`,
		42,
		"github.com/baz/honk",
	)
	if _, err := db.ExecContext(context.Background(), query.Query(sqlf.PostgresBindVar), query.Args()...); err != nil {
		t.Fatalf("unexpected error inserting repo: %s", err)
	}

	query = sqlf.Sprintf(
		`INSERT INTO lsif_index_configuration (id, repository_id, data) VALUES (%s, %s, %s)`,
		1,
		42,
		expectedConfigurationData,
	)
	if _, err := db.ExecContext(context.Background(), query.Query(sqlf.PostgresBindVar), query.Args()...); err != nil {
		t.Fatalf("unexpected error inserting repo: %s", err)
	}

	indexConfiguration, ok, err := store.GetIndexConfigurationByRepositoryID(context.Background(), 42)
	if err != nil {
		t.Fatalf("unexpected error while fetching index configuration: %s", err)
	}
	if !ok {
		t.Fatalf("expected a configuration record")
	}

	if diff := cmp.Diff(expectedConfigurationData, indexConfiguration.Data); diff != "" {
		t.Errorf("unexpected configuration payload (-want +got):\n%s", diff)
	}
}

func TestUpdateIndexConfigurationByRepositoryID(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))
	store := New(&observation.TestContext, db)

	query := sqlf.Sprintf(
		`INSERT INTO repo (id, name) VALUES (%s, %s)`,
		42,
		"github.com/baz/honk",
	)
	if _, err := db.ExecContext(context.Background(), query.Query(sqlf.PostgresBindVar), query.Args()...); err != nil {
		t.Fatalf("unexpected error inserting repo: %s", err)
	}

	expectedConfigurationDataInsert := []byte(`{
		"foo": "bar",
		"baz": "bonk",
	}`)
	if err := store.UpdateIndexConfigurationByRepositoryID(context.Background(), 42, expectedConfigurationDataInsert); err != nil {
		t.Fatalf("unexpected error while fetching index configuration: %s", err)
	}
	if indexConfiguration, ok, err := store.GetIndexConfigurationByRepositoryID(context.Background(), 42); err != nil {
		t.Fatalf("unexpected error while fetching index configuration: %s", err)
	} else if !ok {
		t.Fatalf("expected a configuration record")
	} else if diff := cmp.Diff(expectedConfigurationDataInsert, indexConfiguration.Data); diff != "" {
		t.Errorf("unexpected configuration payload (-want +got):\n%s", diff)
	}

	expectedConfigurationDataUpdate := []byte(`{
		"foo": "baz",
		"baz": "bonk",
	}`)
	if err := store.UpdateIndexConfigurationByRepositoryID(context.Background(), 42, expectedConfigurationDataUpdate); err != nil {
		t.Fatalf("unexpected error while fetching index configuration: %s", err)
	}
	if indexConfiguration, ok, err := store.GetIndexConfigurationByRepositoryID(context.Background(), 42); err != nil {
		t.Fatalf("unexpected error while fetching index configuration: %s", err)
	} else if !ok {
		t.Fatalf("expected a configuration record")
	} else if diff := cmp.Diff(expectedConfigurationDataUpdate, indexConfiguration.Data); diff != "" {
		t.Errorf("unexpected configuration payload (-want +got):\n%s", diff)
	}
}

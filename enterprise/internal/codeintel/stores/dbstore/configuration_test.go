package dbstore

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtesting"
)

func TestGetIndexConfigurationByRepositoryID(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	db := dbtesting.GetDB(t)
	store := testStore(db)

	expectedConfigurationData := []byte(`{
		"foo": "bar",
		"baz": "bonk",
	}`)

	query := sqlf.Sprintf(
		`INSERT INTO repo (id, name) VALUES (%s, %s)`,
		42,
		"github.com/baz/honk",
	)
	if _, err := db.Exec(query.Query(sqlf.PostgresBindVar), query.Args()...); err != nil {
		t.Fatalf("unexpected error inserting repo: %s", err)
	}

	query = sqlf.Sprintf(
		`INSERT INTO lsif_index_configuration (id, repository_id, data) VALUES (%s, %s, %s)`,
		1,
		42,
		expectedConfigurationData,
	)
	if _, err := db.Exec(query.Query(sqlf.PostgresBindVar), query.Args()...); err != nil {
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
	if testing.Short() {
		t.Skip()
	}
	db := dbtesting.GetDB(t)
	store := testStore(db)

	query := sqlf.Sprintf(
		`INSERT INTO repo (id, name) VALUES (%s, %s)`,
		42,
		"github.com/baz/honk",
	)
	if _, err := db.Exec(query.Query(sqlf.PostgresBindVar), query.Args()...); err != nil {
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

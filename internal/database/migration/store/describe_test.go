package store

import (
	"context"
	"encoding/json"
	"io"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/store/testdata"
)

func TestDescribe(t *testing.T) {
	db := dbtest.NewRawDB(t)
	store := testStore(db)
	ctx := context.Background()

	if _, err := db.Exec(readQuery(t)); err != nil {
		t.Fatalf("failed to create database objects: %s", err)
	}

	schemas, err := store.Describe(ctx)
	if err != nil {
		t.Fatalf("unexpected error describing schema: %s", err)
	}

	expectedSchemas := readGolden(t)

	mx := map[string]json.RawMessage{}
	for schemaName, schema := range schemas {
		p, _ := json.Marshal(schema)
		mx[schemaName] = p

		if diff := cmp.Diff(expectedSchemas[schemaName], schema); diff != "" {
			t.Errorf("unexpected payload for schema %s (-want +got):\n%s", schemaName, diff)
		}
	}
}

func readQuery(t *testing.T) string {
	return readTestdataFile(t, "schema.sql")
}

func readGolden(t *testing.T) (m map[string]Schema) {
	filename := "schema.golden.json"

	if err := json.Unmarshal([]byte(readTestdataFile(t, filename)), &m); err != nil {
		t.Fatalf("failed to deserialize testdata/%s: %s", filename, err)
	}

	return m
}

func readTestdataFile(t *testing.T, name string) string {
	f, err := testdata.Data.Open(name)
	if err != nil {
		t.Fatalf("failed to open testdata/%s: %s", name, err)
	}
	defer f.Close()

	query, err := io.ReadAll(f)
	if err != nil {
		t.Fatalf("failed to read testdata/%s: %s", name, err)
	}

	return string(query)
}

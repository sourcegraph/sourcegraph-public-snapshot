package lsifstore

import (
	"database/sql"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

const testBundleID = 39162
const testIndexPath = "./testdata/lsif-go@ad3507cb.sql"

func testStore(t testing.TB) (*sql.DB, *Store) {
	if testing.Short() {
		t.Skip()
	}

	db := dbtesting.GetDB2(t, "lsifstore", []string{
		"lsif_data_definitions",
		"lsif_data_documents",
		"lsif_data_metadata",
		"lsif_data_references",
		"lsif_data_result_chunks",
	})

	store := NewStore(db, &observation.TestContext)
	return db, store
}

func populateTestIndex(t testing.TB, db *sql.DB) {
	contents, err := ioutil.ReadFile(testIndexPath)
	if err != nil {
		t.Fatalf("unexpected error reading testdata: %s", err)
	}

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("unexpected error starting transaction: %s", err)
	}
	defer func() {
		if err := tx.Commit(); err != nil {
			t.Fatalf("unexpected error finishing transaction: %s", err)
		}
	}()

	for _, line := range strings.Split(string(contents), "\n") {
		if line == "" || strings.HasPrefix(line, "---") {
			continue
		}

		if _, err := tx.Exec(line); err != nil {
			t.Fatalf("unexpected error loading database data: %s", err)
		}
	}
}

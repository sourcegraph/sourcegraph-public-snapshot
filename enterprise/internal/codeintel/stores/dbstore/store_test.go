package dbstore

import (
	"database/sql"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func testStore(t testing.TB) (*sql.DB, *Store) {
	if testing.Short() {
		t.Skip()
	}

	db := dbtesting.GetDB2(t, "codeintel", []string{
		"lsif_dirty_repositories",
		"lsif_index_configuration",
		"lsif_indexable_repositories",
		"lsif_indexes",
		"lsif_nearest_uploads_links",
		"lsif_nearest_uploads",
		"lsif_packages",
		"lsif_references",
		"lsif_uploads_visible_at_tip",
		"lsif_uploads",
		"repo",
	})

	store := NewWithDB(db, &observation.TestContext)
	return db, store
}

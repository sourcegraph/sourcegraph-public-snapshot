package dbstore_test

import (
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func testStore(db dbutil.DB) *dbstore.Store {
	return dbstore.NewWithDB(db, &observation.TestContext, nil)
}

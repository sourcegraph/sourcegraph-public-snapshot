package dbstore

import (
	"github.com/sourcegraph/sourcegraph/internal/database/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func init() {
	dbtesting.DBNameSuffix = "codeintel"
}

func testStore() *Store {
	return NewWithDB(dbconn.Global, &observation.TestContext)
}

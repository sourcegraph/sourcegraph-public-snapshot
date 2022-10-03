package dbstore

import (
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func testStore(db database.DB) *Store {
	return NewWithDB(db, &observation.TestContext)
}

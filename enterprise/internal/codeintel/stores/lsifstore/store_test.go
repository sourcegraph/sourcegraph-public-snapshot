package lsifstore

import (
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func init() {
	dbtesting.DBNameSuffix = "lsifstore"
}

func testStore() Store {
	// Wrap in observed, as that's how it's used in production
	return NewObserved(NewStore(dbconn.Global), &observation.TestContext)
}

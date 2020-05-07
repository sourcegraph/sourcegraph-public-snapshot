package db

import (
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func init() {
	dbtesting.DBNameSuffix = "precise-code-intel-api-server"
}

func rawTestDB() *dbImpl {
	return &dbImpl{db: dbconn.Global}
}

func testDB() DB {
	// Wrap in observed, as that's how it's used in production
	return NewObserved(rawTestDB(), &observation.Context{}, "test")

}

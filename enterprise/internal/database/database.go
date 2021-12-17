package database

import (
	"github.com/sourcegraph/sourcegraph/internal/database"
)

type EnterpriseDB interface {
	database.DB
}

func NewEnterpriseDB(db database.DB) EnterpriseDB {
	// If the underlying type already implements EnterpriseDB,
	// return that rather than wrapping it. This enables us to
	// pass a mock EnterpriseDB through as a database.DB, and
	// avoid overwriting its mocked methods by wrapping it.
	if edb, ok := db.(EnterpriseDB); ok {
		return edb
	}
	return &enterpriseDB{db}
}

type enterpriseDB struct {
	database.DB
}

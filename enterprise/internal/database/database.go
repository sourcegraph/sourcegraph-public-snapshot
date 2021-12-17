package database

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
)

type EnterpriseDB interface {
	database.DB
	CodeMonitors() CodeMonitorStore
	Perms() PermsStore
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

func (edb *enterpriseDB) CodeMonitors() CodeMonitorStore {
	return &codeMonitorStore{Store: basestore.NewWithHandle(edb.Handle())}
}

func (edb *enterpriseDB) Perms() PermsStore {
	return &permsStore{Store: basestore.NewWithHandle(edb.Handle()), clock: time.Now}
}

package database

import (
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

type DB interface {
	dbutil.DB
	Repos() RepoStore
}

type db struct {
	dbutil.DB
}

func (db *db) Repos() RepoStore {
	return Repos(db.DB)
}

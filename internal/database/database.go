package database

import (
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

type DB interface {
	dbutil.DB
	Repos() RepoStore
}

func NewDB(inner dbutil.DB) DB {
	return &db{inner}
}

type db struct {
	dbutil.DB
}

func (d *db) Repos() RepoStore {
	return Repos(d.DB)
}

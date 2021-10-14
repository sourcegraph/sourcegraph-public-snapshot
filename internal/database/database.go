package database

import (
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

type DB interface {
	dbutil.DB
	Repos() RepoStore
}

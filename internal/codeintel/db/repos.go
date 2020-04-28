package db

import (
	"context"

	"github.com/keegancsmith/sqlf"
)

// RepoName returns the name for the repo with the given identfier. This is the only method
// in this package that touches any table that does not start with `lsif_`.
func (db *dbImpl) RepoName(ctx context.Context, repositoryID int) (string, error) {
	return scanString(db.queryRow(ctx, sqlf.Sprintf(`SELECT name FROM repo WHERE id = %s`, repositoryID)))
}

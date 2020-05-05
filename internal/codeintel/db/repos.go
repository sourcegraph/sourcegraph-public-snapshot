package db

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/pkg/errors"
)

// ErrUnknownRepository occurs when a repository does not exist.
var ErrUnknownRepository = errors.New("unknown repository")

// RepoName returns the name for the repo with the given identifier. This is the only method
// in this package that touches any table that does not start with `lsif_`.
func (db *dbImpl) RepoName(ctx context.Context, repositoryID int) (string, error) {
	name, exists, err := scanFirstString(db.query(
		ctx,
		sqlf.Sprintf(`SELECT name FROM repo WHERE id = %s`, repositoryID),
	))
	if err != nil {
		return "", err
	}
	if !exists {
		return "", ErrUnknownRepository
	}
	return name, nil
}

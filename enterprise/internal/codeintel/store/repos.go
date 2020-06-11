package store

import (
	"context"

	"github.com/keegancsmith/sqlf"
)

// RepoName returns the name for the repo with the given identifier.
func (s *store) RepoName(ctx context.Context, repositoryID int) (string, error) {
	name, exists, err := scanFirstString(s.query(
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

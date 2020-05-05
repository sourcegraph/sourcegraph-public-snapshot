package db

import (
	"context"

	"github.com/keegancsmith/sqlf"
)

// UpdateCommits upserts commits/parent-commit relations for the given repository ID.
func (db *dbImpl) UpdateCommits(ctx context.Context, repositoryID int, commits map[string][]string) (err error) {
	var rows []*sqlf.Query
	for commit, parents := range commits {
		for _, parent := range parents {
			rows = append(rows, sqlf.Sprintf("(%d, %s, %s)", repositoryID, commit, parent))
		}

		if len(parents) == 0 {
			rows = append(rows, sqlf.Sprintf("(%d, %s, NULL)", repositoryID, commit))
		}
	}

	return db.exec(ctx, sqlf.Sprintf(`
		INSERT INTO lsif_commits (repository_id, "commit", parent_commit)
		VALUES %s
		ON CONFLICT DO NOTHING
	`, sqlf.Join(rows, ",")))
}

package db

import (
	"context"
	"database/sql"

	"github.com/keegancsmith/sqlf"
)

// UpdateCommits upserts commits/parent-commit relations for the given repository ID.
func (db *dbImpl) UpdateCommits(ctx context.Context, tx *sql.Tx, repositoryID int, commits map[string][]string) (err error) {
	if tx == nil {
		tx, err = db.db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer func() {
			err = closeTx(tx, err)
		}()
	}
	tw := &transactionWrapper{tx}

	var rows []*sqlf.Query
	for commit, parents := range commits {
		for _, parent := range parents {
			rows = append(rows, sqlf.Sprintf("(%d, %s, %s)", repositoryID, commit, parent))
		}

		if len(parents) == 0 {
			rows = append(rows, sqlf.Sprintf("(%d, %s, NULL)", repositoryID, commit))
		}
	}

	// TODO(efritz) - add a test for conflicting commits
	query := `INSERT INTO lsif_commits (repository_id, "commit", parent_commit) VALUES %s ON CONFLICT DO NOTHING`
	_, err = tw.exec(ctx, sqlf.Sprintf(query, sqlf.Join(rows, ",")))
	return err
}

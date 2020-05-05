package db

import (
	"context"
	"database/sql"

	"github.com/keegancsmith/sqlf"
)

// HasCommit determines if the given commit is known for the given repository.
func (db *dbImpl) HasCommit(ctx context.Context, repositoryID int, commit string) (bool, error) {
	count, err := scanInt(db.queryRow(ctx, sqlf.Sprintf(`SELECT COUNT(*) FROM lsif_commits WHERE repository_id = %s and commit = %s LIMIT 1`, repositoryID, commit)))
	return count > 0, err
}

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

	query := `
		INSERT INTO lsif_commits (repository_id, "commit", parent_commit)
		VALUES %s
		ON CONFLICT DO NOTHING
	`

	_, err = tw.exec(ctx, sqlf.Sprintf(query, sqlf.Join(rows, ",")))
	return err
}

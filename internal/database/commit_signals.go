package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
)

type OwnSignalStore interface {
	AddCommit(ctx context.Context, commit Commit) error
}

type Commit struct {
	RepoID       api.RepoID
	AuthorName   string
	AuthorEmail  string
	Timestamp    time.Time
	CommitSHA    string
	FilesChanged []string
}

type ownSignalStore struct {
	logger log.Logger
	*basestore.Store
}

func (s *ownSignalStore) AddCommit(ctx context.Context, commit Commit) error {
	return s.Store.WithTransact(context.Background(), func(tx *basestore.Store) error {
		// Get or create commit author
		var authorID int
		err := tx.QueryRow(
			ctx,
			sqlf.Sprintf(`SELECT id FROM commit_authors WHERE email = %s AND name = %s`, commit.AuthorEmail, commit.AuthorName),
		).Scan(&authorID)
		if err == sql.ErrNoRows {
			err = tx.QueryRow(
				ctx,
				sqlf.Sprintf(`INSERT INTO commit_authors (email, name) VALUES (%s, %s) RETURNING id`, commit.AuthorEmail, commit.AuthorName),
			).Scan(&authorID)
		}
		if err != nil {
			return err
		}

		// Get or create repo paths
		pathIDs := make([]int, len(commit.FilesChanged))
		for i, path := range commit.FilesChanged {
			// Get or create repo path
			var pathID int
			err = tx.QueryRow(
				ctx,
				sqlf.Sprintf(`SELECT id FROM repo_paths WHERE repo_id = %s AND path = %s`, commit.RepoID, path),
			).Scan(&pathID)
			if err == sql.ErrNoRows {
				err = tx.QueryRow(
					ctx,
					sqlf.Sprintf(`INSERT INTO repo_paths (repo_id, path) VALUES (%s, %s) RETURNING id`, commit.RepoID, path),
				).Scan(&pathID)
			}
			if err != nil {
				return err
			}
			pathIDs[i] = pathID
		}

		// Insert into own_signal_recent_contribution
		for _, pathID := range pathIDs {
			q := sqlf.Sprintf(`INSERT INTO own_signal_recent_contribution (commit_author_id, changed_file_path_id,
				commit_timestamp, commit_id_hash) VALUES (%s, %s, %s, %s)`,
				authorID,
				pathID,
				commit.Timestamp,
				commit.CommitSHA,
			)
			err = tx.Exec(ctx, q)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func OwnSignalsStoreWith(other basestore.ShareableStore) OwnSignalStore {
	return &ownSignalStore{Store: basestore.NewWithHandle(other.Handle())}
}

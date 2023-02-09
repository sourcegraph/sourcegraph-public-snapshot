package database

import (
	"context"
	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"time"
)

// RepoHistory represents the contents of the single row in the
// repo_history table.
type RepoHistory struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	RepoID    int       `json:"repoID"`
	Timestamp time.Time `json:"timestamp"`
	EventType string    `json:"eventType"`
	Message   string    `json:"message"`
	Metadata  string    `json:"metadata"`
}

type RepoHistoryStore interface {
	basestore.ShareableStore
	WithTransact(context.Context, func(RepoHistoryStore) error) error
	With(basestore.ShareableStore) RepoHistoryStore

	GetRepoHistory(ctx context.Context) (RepoHistory, error)
	Create(ctx context.Context, rh RepoHistory) error
}

// repoHistoryStore is responsible for data stored in the repo_statistics
// and the gitserver_repos_statistics tables.
type repoHistoryStore struct {
	*basestore.Store
}

// RepoHistoryWith instantiates and returns a new RepoHistoryStore using
// the other store handle.
func RepoHistoryWith(other basestore.ShareableStore) RepoHistoryStore {
	return &repoHistoryStore{Store: basestore.NewWithHandle(other.Handle())}
}

func (s *repoHistoryStore) With(other basestore.ShareableStore) RepoHistoryStore {
	return &repoHistoryStore{Store: s.Store.With(other)}
}

func (s *repoHistoryStore) WithTransact(ctx context.Context, f func(RepoHistoryStore) error) error {
	return s.Store.WithTransact(ctx, func(tx *basestore.Store) error {
		return f(&repoHistoryStore{Store: tx})
	})
}

func (s *repoHistoryStore) GetRepoHistory(ctx context.Context) (RepoHistory, error) {
	var rs RepoHistory
	row := s.QueryRow(ctx, sqlf.Sprintf(getRepoHistoryQueryFmtstr))
	err := row.Scan(&rs.ID, &rs.Name, &rs.Timestamp, &rs.EventType, &rs.Message, &rs.Metadata)
	if err != nil {
		return rs, err
	}
	return rs, nil
}

const getRepoHistoryQueryFmtstr = `
SELECT
	id,
	name,
	event_type,
	message,
	metadata
FROM repo_history rh, repo r
WHERE rh.repo_id = r.id
`

func (e *repoHistoryStore) Create(ctx context.Context, rh RepoHistory) error {
	return e.QueryRow(
		ctx,
		sqlf.Sprintf(
			createRepoHistoryQueryFmtstr,
			rh.RepoID,
			rh.Timestamp,
			rh.EventType,
			rh.Message,
			rh.Metadata,
		),
	).Scan()
}

const createRepoHistoryQueryFmtstr = `
INSERT INTO repo_history
	(repo_id, timestamp, event_type, message, metadata)
	VALUES(%s, %s, %s, %s, %s)
RETURNING id

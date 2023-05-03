package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type RepoLifeCycleEvent string

const (
	RepoLifeCycleEventAddedFromCodeHostSync RepoLifeCycleEvent = "added from code host sync"
	RepoLifeCycleEventCloneStarted          RepoLifeCycleEvent = "clone started"
	RepoLifeCycleEventCloneCompleted        RepoLifeCycleEvent = "clone completed"
	RepoLifeCycleEventMetadataUpdated       RepoLifeCycleEvent = "metadata updated"

	RepoLifeCycleEventUpdateStarted RepoLifeCycleEvent = "update started"

	RepoLifeCycleEventDeletedFromCodeHostSync RepoLifeCycleEvent = "deleted from code host sync"
	RepoLifeCycleEventPurgedFromDisk          RepoLifeCycleEvent = "purged from disk"
	RepoLifeCycleEventPurgedDiskPressure      RepoLifeCycleEvent = "purged due to disk pressure"
	RepoLifeCycleEventGcStarted               RepoLifeCycleEvent = "git gc started"
	RepoLifeCycleEventForegroundSyncStarted   RepoLifeCycleEvent = "foregound sync started"
	RepoLifeCycleEventBackgroundSyncStarted   RepoLifeCycleEvent = "background sync started"
)

type RepoLifeCycleStore interface {
	Transact(context.Context) (RepoLifeCycleStore, error)
	With(basestore.ShareableStore) RepoLifeCycleStore

	Upsert(ctx context.Context, repoID api.RepoID, event RepoLifeCycleEvent) error
	Get(ctx context.Context, repoID int) (*RepoLifeCycle, error)
}

type repoLifeCycleStore struct {
	*basestore.Store
	logger log.Logger
}

var _ RepoLifeCycleStore = (*repoLifeCycleStore)(nil)

// RepoLifeCycleWith instantiates and returns a new RepoLifeCycleStore using the other store handle.
func RepoLifeCycleWith(logger log.Logger, other basestore.ShareableStore) RepoLifeCycleStore {
	return &repoLifeCycleStore{
		logger: logger,
		Store:  basestore.NewWithHandle(other.Handle()),
	}
}

// Wraps the basestore.With method to return the correct type.
func (s *repoLifeCycleStore) With(other basestore.ShareableStore) RepoLifeCycleStore {
	return &repoLifeCycleStore{Store: s.Store.With(other) /*, ... */}
}

// Wraps the basestore.Transact method to return the correct type.
func (s *repoLifeCycleStore) Transact(ctx context.Context) (RepoLifeCycleStore, error) {
	txBase, err := s.Store.Transact(ctx)
	if err != nil {
		return nil, err
	}

	return &repoLifeCycleStore{Store: txBase /*, ... */}, nil
}

var upsertRepoLifeCycleFmtStr = `
INSERT INTO repos_life_cycle (repo_id, logs)
VALUES (%d, %s)
ON CONFLICT (repo_id) DO UPDATE 
SET logs = repos_life_cycle.logs || %s`

func (s *repoLifeCycleStore) Upsert(ctx context.Context, repoID api.RepoID, event RepoLifeCycleEvent) error {
	log, err := json.Marshal([]repoLifeCycleLogItem{{
		Event:     event,
		Timestamp: time.Now(),
	}})
	if err != nil {
		return errors.Wrapf(err, "reposLifeCycleStore: failed to generate repoLifeCycleLogItem for event %q", event)
	}

	q := sqlf.Sprintf(upsertRepoLifeCycleFmtStr, repoID, string(log), string(log))
	return errors.Wrap(s.Exec(ctx, q), "reposLifeCycleStore: failed to upsert")
}

var getRepoLifeCycleFmtStr = "SELECT repo_id, logs FROM repos_life_cycle WHERE repo_id = %d"

func (s *repoLifeCycleStore) Get(ctx context.Context, repoID int) (*RepoLifeCycle, error) {
	row := s.QueryRow(ctx, sqlf.Sprintf(getRepoLifeCycleFmtStr, repoID))
	item, err := scanRepoLifeCycleRow(row)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return item, err
}

type repoLifeCycleLogItem struct {
	Event     RepoLifeCycleEvent `json:"event"`
	Timestamp time.Time          `json:"timestamp"`
}

type RepoLifeCycle struct {
	repoID int32
	log    string
}

var scanRepoLifeCycleRow = func(scanner dbutil.Scanner) (*RepoLifeCycle, error) {
	var s RepoLifeCycle
	err := scanner.Scan(&s.repoID, &s.log)
	return &s, err
}

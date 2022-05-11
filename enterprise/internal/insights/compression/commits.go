package compression

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type DBCommitStore struct {
	*basestore.Store
}

type CommitStore interface {
	Save(ctx context.Context, id api.RepoID, commit *gitdomain.Commit, debugInfo string) error
	Get(ctx context.Context, id api.RepoID, start time.Time, end time.Time) ([]CommitStamp, error)
	GetMetadata(ctx context.Context, id api.RepoID) (CommitIndexMetadata, error)
	UpsertMetadataStamp(ctx context.Context, id api.RepoID, indexedThrough time.Time) (CommitIndexMetadata, error)
	InsertCommits(ctx context.Context, id api.RepoID, commits []*gitdomain.Commit, indexedThrough time.Time, debugInfo string) error
}

func NewCommitStore(db dbutil.DB) *DBCommitStore {
	return &DBCommitStore{
		Store: basestore.NewWithDB(db, sql.TxOptions{}),
	}
}

func (c *DBCommitStore) With(other basestore.ShareableStore) *DBCommitStore {
	return &DBCommitStore{Store: c.Store.With(other)}
}

func (c *DBCommitStore) Transact(ctx context.Context) (*DBCommitStore, error) {
	txBase, err := c.Store.Transact(ctx)
	return &DBCommitStore{Store: txBase}, err
}

func (c *DBCommitStore) Save(ctx context.Context, id api.RepoID, commit *gitdomain.Commit, debugInfo string) error {
	commitID := commit.ID
	debugMsg := fmt.Sprintf("author:%s|msgSub:%s|msgBody:%s|commitTime:%s|authorTime:%s|debugInfo:%s", commit.Author.Name, commit.Message.Subject(), commit.Message.Body(), commit.Committer.Date, commit.Author.Date, debugInfo)
	if err := c.Exec(ctx, sqlf.Sprintf(insertCommitIndexStr, id, dbutil.CommitBytea(commitID), commit.Committer.Date, debugMsg)); err != nil {
		return errors.Errorf("error saving commit for repo_id: %v commit_id %v: %w", id, commitID, err)
	}

	return nil
}

func (c *DBCommitStore) InsertCommits(ctx context.Context, id api.RepoID, commits []*gitdomain.Commit, indexedThrough time.Time, debugInfo string) (err error) {
	tx, err := c.Transact(ctx)
	if err != nil {
		return err
	}

	defer func() { err = tx.Store.Done(err) }()

	for _, commit := range commits {
		if err = tx.Save(ctx, id, commit, debugInfo); err != nil {
			return err
		}
	}

	if _, err = tx.UpsertMetadataStamp(ctx, id, indexedThrough); err != nil {
		return err
	}

	return nil
}

// Get Fetch all commits that occur for a specific repository and fall in a specific time range. The time range
// is start inclusive and end exclusive [start, end)
func (c *DBCommitStore) Get(ctx context.Context, id api.RepoID, start time.Time, end time.Time) (_ []CommitStamp, err error) {
	rows, err := c.Query(ctx, sqlf.Sprintf(getCommitsInRangeStr, id, start, end))
	if err != nil {
		return []CommitStamp{}, err
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	results := make([]CommitStamp, 0)
	for rows.Next() {
		var stamp CommitStamp
		if err := rows.Scan(&stamp.RepoID, &stamp.Commit, &stamp.CommittedAt); err != nil {
			return []CommitStamp{}, err
		}

		results = append(results, stamp)
	}

	return results, nil
}

// GetMetadata Returns commit index metadata for a given repository
func (c *DBCommitStore) GetMetadata(ctx context.Context, id api.RepoID) (CommitIndexMetadata, error) {
	row := c.QueryRow(ctx, sqlf.Sprintf(getCommitIndexMetadataStr, id, id))

	var metadata CommitIndexMetadata
	if err := row.Scan(&metadata.RepoId, &metadata.Enabled, &metadata.LastIndexedAt, &metadata.OldestIndexedAt); err != nil {
		return CommitIndexMetadata{}, err
	}

	return metadata, nil
}

// UpsertMetadataStamp inserts (or updates, if the row already exists) the index metadata timestamp for a given repository
func (c *DBCommitStore) UpsertMetadataStamp(ctx context.Context, id api.RepoID, indexedThrough time.Time) (CommitIndexMetadata, error) {
	row := c.QueryRow(ctx, sqlf.Sprintf(upsertCommitIndexMetadataStampStr, id, indexedThrough))

	var metadata CommitIndexMetadata
	if err := row.Scan(&metadata.RepoId, &metadata.Enabled, &metadata.LastIndexedAt); err != nil {
		return CommitIndexMetadata{}, err
	}

	return metadata, nil
}

type CommitStamp struct {
	RepoID      int
	Commit      dbutil.CommitBytea
	CommittedAt time.Time
}

type CommitIndexMetadata struct {
	RepoId          int
	Enabled         bool
	LastIndexedAt   time.Time
	OldestIndexedAt *time.Time
}

const getCommitsInRangeStr = `
-- source: enterprise/internal/insights/compression/commits.go:Get
SELECT repo_id, commit_bytea, committed_at FROM commit_index WHERE repo_id = %s AND committed_at >= %s AND committed_at < %s ORDER BY committed_at desc;
`

const insertCommitIndexStr = `
-- source: enterprise/internal/insights/compression/commits.go:Save
INSERT INTO commit_index(repo_id, commit_bytea, committed_at, debug_field)
VALUES (%s, %s, %s, %s)
ON CONFLICT ON CONSTRAINT commit_index_pkey DO NOTHING;
`

const getCommitIndexMetadataStr = `
-- source: enterprise/internal/insights/compression/commits.go:GetMetadata
SELECT repo_id, enabled, last_indexed_at, (select min(committed_at) from commit_index where repo_id = %s) oldest_indexed_at
FROM commit_index_metadata
WHERE repo_id = %s;
`

const upsertCommitIndexMetadataStampStr = `
-- source: enterprise/internal/insights/compression/commits.go:UpsertMetadataStamp
INSERT INTO commit_index_metadata(repo_id)
VALUES (%v)
ON CONFLICT (repo_id) DO UPDATE
SET last_indexed_at = %v
RETURNING repo_id, enabled, last_indexed_at;
`

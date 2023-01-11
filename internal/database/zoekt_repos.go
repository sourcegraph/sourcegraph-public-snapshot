package database

import (
	"context"
	"encoding/json"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/zoekt"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/batch"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type ZoektReposStore interface {
	basestore.ShareableStore

	With(other basestore.ShareableStore) ZoektReposStore

	// UpdateIndexStatuses updates the index status of the rows in zoekt_repos
	// whose repo_id matches an entry in the `indexed` map.
	UpdateIndexStatuses(ctx context.Context, indexed map[uint32]*zoekt.MinimalRepoListEntry) error

	// GetStatistics returns a summary of the zoekt_repos table.
	GetStatistics(ctx context.Context) (ZoektRepoStatistics, error)

	// GetZoektRepo returns the ZoektRepo for the given repository ID.
	GetZoektRepo(ctx context.Context, repo api.RepoID) (*ZoektRepo, error)
}

var _ ZoektReposStore = (*zoektReposStore)(nil)

// zoektReposStore is responsible for data stored in the zoekt_repos table.
type zoektReposStore struct {
	*basestore.Store
}

// ZoektReposWith instantiates and returns a new zoektReposStore using
// the other store handle.
func ZoektReposWith(other basestore.ShareableStore) ZoektReposStore {
	return &zoektReposStore{Store: basestore.NewWithHandle(other.Handle())}
}

func (s *zoektReposStore) With(other basestore.ShareableStore) ZoektReposStore {
	return &zoektReposStore{Store: s.Store.With(other)}
}

func (s *zoektReposStore) Transact(ctx context.Context) (ZoektReposStore, error) {
	txBase, err := s.Store.Transact(ctx)
	return &zoektReposStore{Store: txBase}, err
}

type ZoektRepo struct {
	RepoID      api.RepoID
	IndexStatus string
	Branches    []zoekt.RepositoryBranch

	UpdatedAt time.Time
	CreatedAt time.Time
}

func (s *zoektReposStore) GetZoektRepo(ctx context.Context, repo api.RepoID) (*ZoektRepo, error) {
	return scanZoektRepo(s.QueryRow(ctx, sqlf.Sprintf(getZoektRepoQueryFmtstr, repo)))
}

func scanZoektRepo(sc dbutil.Scanner) (*ZoektRepo, error) {
	var zr ZoektRepo
	var branches json.RawMessage

	err := sc.Scan(
		&zr.RepoID,
		&branches,
		&zr.IndexStatus,
		&zr.UpdatedAt,
		&zr.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(branches, &zr.Branches); err != nil {
		return nil, errors.Wrapf(err, "scanZoektRepo: failed to unmarshal branches")
	}

	return &zr, nil
}

const getZoektRepoQueryFmtstr = `
SELECT
	zr.repo_id,
	zr.branches,
	zr.index_status,
	zr.updated_at,
	zr.created_at
FROM zoekt_repos zr
JOIN repo ON repo.id = zr.repo_id
WHERE
	repo.deleted_at is NULL
AND
	repo.blocked IS NULL
AND
	zr.repo_id = %s
;
`

func (s *zoektReposStore) UpdateIndexStatuses(ctx context.Context, indexed map[uint32]*zoekt.MinimalRepoListEntry) (err error) {
	tx, err := s.Store.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	if err := tx.Exec(ctx, sqlf.Sprintf(updateIndexStatusesCreateTempTableQuery)); err != nil {
		return err
	}

	inserter := batch.NewInserter(ctx, tx.Handle(), "temp_table", batch.MaxNumPostgresParameters, tempTableColumns...)

	for repoID, entry := range indexed {
		branches, err := branchesColumn(entry.Branches)
		if err != nil {
			return err
		}

		if err := inserter.Insert(ctx, repoID, "indexed", branches); err != nil {
			return err
		}
	}

	if err := inserter.Flush(ctx); err != nil {
		return err
	}

	if err := tx.Exec(ctx, sqlf.Sprintf(updateIndexStatusesUpdateQuery)); err != nil {
		return errors.Wrap(err, "updating zoekt repos failed")
	}

	return nil
}

func branchesColumn(branches []zoekt.RepositoryBranch) (msg json.RawMessage, err error) {
	if len(branches) == 0 {
		msg = json.RawMessage("[]")
	} else {
		msg, err = json.Marshal(branches)
	}
	return
}

var tempTableColumns = []string{
	"repo_id",
	"index_status",
	"branches",
}

const updateIndexStatusesCreateTempTableQuery = `
CREATE TEMPORARY TABLE temp_table (
	repo_id      integer NOT NULL,
	index_status text NOT NULL,
	branches     jsonb
) ON COMMIT DROP
`

const updateIndexStatusesUpdateQuery = `
UPDATE zoekt_repos zr
SET
	index_status = source.index_status,
	branches     = source.branches,
	updated_at   = now()
FROM temp_table source
WHERE
	zr.repo_id = source.repo_id
AND
	(zr.index_status != source.index_status OR zr.branches != source.branches)
;
`

type ZoektRepoStatistics struct {
	Total      int
	Indexed    int
	NotIndexed int
}

func (s *zoektReposStore) GetStatistics(ctx context.Context) (ZoektRepoStatistics, error) {
	var zrs ZoektRepoStatistics
	row := s.QueryRow(ctx, sqlf.Sprintf(getZoektRepoStatisticsQueryFmtstr))
	err := row.Scan(&zrs.Total, &zrs.Indexed, &zrs.NotIndexed)
	if err != nil {
		return zrs, err
	}
	return zrs, nil
}

const getZoektRepoStatisticsQueryFmtstr = `
-- source: internal/database/zoekt_repos.go:zoektReposStore.GetStatistics
SELECT
	COUNT(*) AS total,
	COUNT(*) FILTER(WHERE index_status = 'indexed') AS indexed,
	COUNT(*) FILTER(WHERE index_status = 'not_indexed') AS not_indexed
FROM zoekt_repos zr
JOIN repo ON repo.id = zr.repo_id
WHERE
	repo.deleted_at is NULL
AND
	repo.blocked IS NULL
;
`

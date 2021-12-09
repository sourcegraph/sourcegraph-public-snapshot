package database

import (
	"context"
	"database/sql"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type OrgStatsStore interface {
	Upsert(ctx context.Context, orgID int32, codeHostRepoCount int32) (*types.OrgStats, error)
	With(basestore.ShareableStore) OrgStatsStore
	basestore.ShareableStore
}

type orgStatsStore struct {
	*basestore.Store
}

// OrgStats instantiates and returns a new OrgStatsStore with prepared statements.
func OrgStats(db dbutil.DB) OrgStatsStore {
	return &orgStatsStore{Store: basestore.NewWithDB(db, sql.TxOptions{})}
}

// OrgStatsWith instantiates and returns a new OrgStatsStore using the other store handle.
func OrgStatsWith(other basestore.ShareableStore) OrgStatsStore {
	return &orgStatsStore{Store: basestore.NewWithHandle(other.Handle())}
}

func (o *orgStatsStore) With(other basestore.ShareableStore) OrgStatsStore {
	return &orgStatsStore{Store: o.Store.With(other)}
}

func (o *orgStatsStore) Upsert(ctx context.Context, orgID int32, codeHostRepoCount int32) (*types.OrgStats, error) {
	newStatistic := &types.OrgStats{
		OrgID:             orgID,
		CodeHostRepoCount: codeHostRepoCount,
	}
	err := o.Handle().DB().QueryRowContext(
		ctx,
		"INSERT INTO org_stats(org_id, code_host_repo_count) VALUES($1, $2) ON CONFLICT (org_id) DO UPDATE SET code_host_repo_count = $2 RETURNING code_host_repo_count;",
		newStatistic.OrgID, newStatistic.CodeHostRepoCount).Scan(&newStatistic.CodeHostRepoCount)
	if err != nil {
		return nil, err
	}

	return newStatistic, nil
}

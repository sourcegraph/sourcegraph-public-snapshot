package store

import (
	"context"
	"database/sql"

	"github.com/cockroachdb/errors"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

type InsightPermStore struct {
	*basestore.Store
}

// GetUnauthorizedRepoIDs returns a list of repo IDs that the current user does *not* have access to. The primary
// purpose of this is to quickly resolve permissions at query time from the primary postgres database and filter
// code insights in the timeseries database. This approach makes the assumption that most users have access to most
// repos - which is highly likely given the public / private model that repos use today.
func (i *InsightPermStore) GetUnauthorizedRepoIDs(ctx context.Context) (results []api.RepoID, err error) {
	db := i.Store.Handle().DB()
	db, done, err := database.WithEnforcedAuthz(ctx, db)
	if err != nil {
		return []api.RepoID{}, errors.Wrap(err, "enforcing authz")
	}
	defer func() { err = done(err) }()
	store := database.Repos(db)

	rows, err := store.Query(ctx, sqlf.Sprintf(fetchUnauthorizedReposSql))
	if err != nil {
		return []api.RepoID{}, err
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	for rows.Next() {
		var temp int
		if err := rows.Scan(&temp); err != nil {
			return []api.RepoID{}, err
		}
		results = append(results, api.RepoID(temp))
	}

	return results, nil
}

// All non deleted repos will have at least one entry in external_service_repos.
// We join against the repo table so that our permissions checking on that table
// kicks in and any joins that don't match indicate a repo we don't have access
// to.
const fetchUnauthorizedReposSql = `
-- source: enterprise/internal/insights/resolver/permissions.go:FetchUnauthorizedRepos
	SELECT distinct(repo_id) FROM external_service_repos esr
    LEFT JOIN repo r on r.id = esr.repo_id
    WHERE r.id IS NULL
`

func NewInsightPermissionStore(db dbutil.DB) *InsightPermStore {
	return &InsightPermStore{
		Store: basestore.NewWithDB(db, sql.TxOptions{}),
	}
}

type InsightPermissionStore interface {
	GetUnauthorizedRepoIDs(ctx context.Context) (results []api.RepoID, err error)
}

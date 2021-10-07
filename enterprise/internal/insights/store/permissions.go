package store

import (
	"context"
	"database/sql"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
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
	store := database.Repos(db)
	conds, err := database.AuthzQueryConds(ctx, db)
	if err != nil {
		return []api.RepoID{}, err
	}

	q := sqlf.Join([]*sqlf.Query{sqlf.Sprintf(fetchUnauthorizedReposSql), conds}, " ")

	rows, err := store.Query(ctx, q)
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

const fetchUnauthorizedReposSql = `
-- source: enterprise/internal/insights/resolver/permissions.go:FetchUnauthorizedRepos
	SELECT id FROM repo WHERE NOT`

func NewInsightPermissionStore(db dbutil.DB) *InsightPermStore {
	return &InsightPermStore{
		Store: basestore.NewWithDB(db, sql.TxOptions{}),
	}
}

type InsightPermissionStore interface {
	GetUnauthorizedRepoIDs(ctx context.Context) (results []api.RepoID, err error)
}

type InsightViewGrant struct {
	UserID *int
	OrgID  *int
	Global *bool
}

func (i InsightViewGrant) toQuery(insightViewID int) *sqlf.Query {
	// insight_view_id, org_id, user_id, global
	valuesFmt := "(%s, %s, %s, %s)"
	return sqlf.Sprintf(valuesFmt, insightViewID, i.OrgID, i.UserID, i.Global)
}

func UserGrant(userID int) InsightViewGrant {
	return InsightViewGrant{UserID: &userID}
}

func OrgGrant(orgID int) InsightViewGrant {
	return InsightViewGrant{OrgID: &orgID}
}

func GlobalGrant() InsightViewGrant {
	b := true
	return InsightViewGrant{Global: &b}
}

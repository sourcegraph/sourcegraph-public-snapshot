package store

import (
	"context"
	"database/sql"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type InsightPermStore struct {
	logger log.Logger
	*basestore.Store
}

func NewInsightPermissionStore(db database.DB) *InsightPermStore {
	return &InsightPermStore{
		logger: log.Scoped("InsightPermStore"),
		Store:  basestore.NewWithHandle(db.Handle()),
	}
}

type InsightPermissionStore interface {
	GetUnauthorizedRepoIDs(ctx context.Context) (results []api.RepoID, err error)
	GetUserPermissions(ctx context.Context) (userIDs []int, orgIDs []int, err error)
}

// GetUnauthorizedRepoIDs returns a list of repo IDs that the current user does *not* have access to. The primary
// purpose of this is to quickly resolve permissions at query time from the primary postgres database and filter
// code insights in the timeseries database. This approach makes the assumption that most users have access to most
// repos - which is highly likely given the public / private model that repos use today.
func (i *InsightPermStore) GetUnauthorizedRepoIDs(ctx context.Context) (results []api.RepoID, err error) {
	db := database.NewDBWith(i.logger, i.Store)
	store := db.Repos()
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
SELECT id FROM repo WHERE NOT
`

func (i *InsightPermStore) GetUserPermissions(ctx context.Context) ([]int, []int, error) {
	db := database.NewDBWith(i.logger, i.Store)
	orgStore := db.Orgs()

	currentActor := actor.FromContext(ctx)
	var userIDs, orgIds []int
	if currentActor.IsAuthenticated() {
		userId := currentActor.UID // UID is only equal to 0 if the actor is unauthenticated.
		orgs, err := orgStore.GetByUserID(ctx, userId)
		if err != nil {
			return nil, nil, errors.Wrap(err, "GetByUserID")
		}
		for _, org := range orgs {
			orgIds = append(orgIds, int(org.ID))
		}
		userIDs = append(userIDs, int(userId))
	}
	return userIDs, orgIds, nil
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

type DashboardGrant struct {
	UserID *int
	OrgID  *int
	Global *bool
}

func scanDashboardGrants(rows *sql.Rows, queryErr error) (_ []*DashboardGrant, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	var results []*DashboardGrant
	var placeholder int
	for rows.Next() {
		var temp DashboardGrant
		if err := rows.Scan(
			&placeholder,
			&placeholder,
			&temp.UserID,
			&temp.OrgID,
			&temp.Global,
		); err != nil {
			return []*DashboardGrant{}, err
		}
		results = append(results, &temp)
	}

	return results, nil
}

func (i DashboardGrant) IsValid() bool {
	if i.OrgID != nil || i.UserID != nil || i.Global != nil {
		return true
	}
	return false
}

func (i DashboardGrant) toQuery(dashboardID int) (*sqlf.Query, error) {
	if !i.IsValid() {
		return nil, errors.New("invalid dashboard grant, no principal assigned")
	}
	// dashboard_id, user_id, org_id, global
	valuesFmt := "(%s, %s, %s, %s)"
	return sqlf.Sprintf(valuesFmt, dashboardID, i.UserID, i.OrgID, i.Global), nil
}

func UserDashboardGrant(userID int) DashboardGrant {
	return DashboardGrant{UserID: &userID}
}

func OrgDashboardGrant(orgID int) DashboardGrant {
	return DashboardGrant{OrgID: &orgID}
}

func GlobalDashboardGrant() DashboardGrant {
	b := true
	return DashboardGrant{Global: &b}
}

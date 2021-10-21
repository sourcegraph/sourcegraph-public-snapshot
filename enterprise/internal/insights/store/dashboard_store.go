package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/lib/pq"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

type DBDashboardStore struct {
	*basestore.Store
	Now func() time.Time
}

// NewDashboardStore returns a new DBDashboardStore backed by the given Timescale db.
func NewDashboardStore(db dbutil.DB) *DBDashboardStore {
	return &DBDashboardStore{Store: basestore.NewWithDB(db, sql.TxOptions{}), Now: time.Now}
}

// Handle returns the underlying transactable database handle.
// Needed to implement the ShareableStore interface.
func (s *DBDashboardStore) Handle() *basestore.TransactableHandle { return s.Store.Handle() }

// With creates a new DBDashboardStore with the given basestore. Shareable store as the underlying basestore.Store.
// Needed to implement the basestore.Store interface
func (s *DBDashboardStore) With(other *DBDashboardStore) *DBDashboardStore {
	return &DBDashboardStore{Store: s.Store.With(other.Store), Now: other.Now}
}

func (s *DBDashboardStore) Transact(ctx context.Context) (*DBDashboardStore, error) {
	txBase, err := s.Store.Transact(ctx)
	return &DBDashboardStore{Store: txBase, Now: s.Now}, err
}

type DashboardQueryArgs struct {
	UserID  []int
	OrgID   []int
	ID      int
	Deleted bool
	Limit   int
	After   int
}

func (s *DBDashboardStore) GetDashboards(ctx context.Context, args DashboardQueryArgs) ([]*types.Dashboard, error) {
	preds := make([]*sqlf.Query, 0, 1)
	if args.ID > 0 {
		preds = append(preds, sqlf.Sprintf("db.id = %s", args.ID))
	}
	if args.Deleted {
		preds = append(preds, sqlf.Sprintf("db.deleted_at is not null"))
	} else {
		preds = append(preds, sqlf.Sprintf("db.deleted_at is null"))
	}
	if args.After > 0 {
		preds = append(preds, sqlf.Sprintf("db.id > %s", args.After))
	}

	preds = append(preds, dashboardPermissionsQuery(args))
	if len(preds) == 0 {
		preds = append(preds, sqlf.Sprintf("%s", "TRUE"))
	}
	var limitClause *sqlf.Query
	if args.Limit > 0 {
		limitClause = sqlf.Sprintf("LIMIT %s", args.Limit)
	} else {
		limitClause = sqlf.Sprintf("")
	}

	q := sqlf.Sprintf(getDashboardsSql, sqlf.Join(preds, "\n AND"), limitClause)
	return scanDashboard(s.Query(ctx, q))
}

func (s *DBDashboardStore) DeleteDashboard(ctx context.Context, id int64) error {
	err := s.Exec(ctx, sqlf.Sprintf(deleteDashboardSql, id))
	if err != nil {
		return errors.Wrapf(err, "failed to delete dashboard with id: %s", id)
	}
	return nil
}

func dashboardPermissionsQuery(args DashboardQueryArgs) *sqlf.Query {
	permsPreds := make([]*sqlf.Query, 0, 2)
	if len(args.OrgID) > 0 {
		elems := make([]*sqlf.Query, 0, len(args.OrgID))
		for _, id := range args.OrgID {
			elems = append(elems, sqlf.Sprintf("%s", id))
		}
		permsPreds = append(permsPreds, sqlf.Sprintf("dg.org_id IN (%s)", sqlf.Join(elems, ",")))
	}
	if len(args.UserID) > 0 {
		elems := make([]*sqlf.Query, 0, len(args.UserID))
		for _, id := range args.UserID {
			elems = append(elems, sqlf.Sprintf("%s", id))
		}
		permsPreds = append(permsPreds, sqlf.Sprintf("dg.user_id IN (%s)", sqlf.Join(elems, ",")))
	}
	permsPreds = append(permsPreds, sqlf.Sprintf("dg.global is true"))
	return sqlf.Sprintf("(%s)", sqlf.Join(permsPreds, "OR"))
}

func scanDashboard(rows *sql.Rows, queryErr error) (_ []*types.Dashboard, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	var results []*types.Dashboard
	for rows.Next() {
		var temp types.Dashboard
		if err := rows.Scan(
			&temp.ID,
			&temp.Title,
			pq.Array(&temp.InsightIDs),
			pq.Array(&temp.UserIdGrants),
			pq.Array(&temp.OrgIdGrants),
			&temp.GlobalGrant,
		); err != nil {
			return []*types.Dashboard{}, err
		}
		results = append(results, &temp)
	}
	return results, nil
}

const getDashboardsSql = `
-- source: enterprise/internal/insights/store/dashboard_store.go:GetDashboards
SELECT db.id, db.title, t.uuid_array as insight_view_unique_ids,
	ARRAY_REMOVE(ARRAY_AGG(dg.user_id), NULL) AS granted_users,
	ARRAY_REMOVE(ARRAY_AGG(dg.org_id), NULL)  AS granted_orgs,
	BOOL_OR(dg.global IS TRUE)                AS granted_global
FROM dashboard db
         JOIN dashboard_grants dg ON db.id = dg.dashboard_id
         LEFT JOIN (SELECT ARRAY_AGG(iv.unique_id) AS uuid_array, div.dashboard_id
               FROM insight_view iv
                        JOIN dashboard_insight_view div ON iv.id = div.insight_view_id
               GROUP BY div.dashboard_id) t on t.dashboard_id = db.id
WHERE %S
GROUP BY db.id, t.uuid_array
ORDER BY db.id
%S;
`

const deleteDashboardSql = `
-- source: enterprise/internal/insights/store/dashboard_store.go:DeleteDashboard
update dashboard set deleted_at = NOW() where id = %s;
`

type CreateDashboardArgs struct {
	Dashboard types.Dashboard
	Grants    []DashboardGrant
	UserID    []int // For dashboard permissions
	OrgID     []int // For dashboard permissions
}

func (s *DBDashboardStore) CreateDashboard(ctx context.Context, args CreateDashboardArgs) (_ *types.Dashboard, err error) {
	tx, err := s.Transact(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.Done(err) }()

	row := tx.QueryRow(ctx, sqlf.Sprintf(insertDashboardSql,
		args.Dashboard.Title,
		args.Dashboard.Save,
	))
	if row.Err() != nil {
		return nil, row.Err()
	}
	var dashboardId int
	err = row.Scan(&dashboardId)
	if err != nil {
		return nil, errors.Wrap(err, "CreateDashboard")
	}
	err = tx.AddViewsToDashboard(ctx, dashboardId, args.Dashboard.InsightIDs)
	if err != nil {
		return nil, errors.Wrap(err, "AddViewsToDashboard")
	}
	err = tx.AddDashboardGrants(ctx, dashboardId, args.Grants)
	if err != nil {
		return nil, errors.Wrap(err, "AddDashboardGrants")
	}

	dashboards, err := tx.GetDashboards(ctx, DashboardQueryArgs{ID: dashboardId, UserID: args.UserID, OrgID: args.OrgID})
	if err != nil {
		return nil, errors.Wrap(err, "GetDashboards")
	}
	if len(dashboards) > 0 {
		return dashboards[0], nil
	}
	return nil, nil
}

type UpdateDashboardArgs struct {
	ID     int
	Title  *string
	Grants []DashboardGrant
	UserID []int // For dashboard permissions
	OrgID  []int // For dashboard permissions
}

func (s *DBDashboardStore) UpdateDashboard(ctx context.Context, args UpdateDashboardArgs) (_ *types.Dashboard, err error) {
	tx, err := s.Transact(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.Done(err) }()

	if args.Title != nil {
		err := tx.Exec(ctx, sqlf.Sprintf(updateDashboardSql,
			*args.Title,
			args.ID,
		))
		if err != nil {
			return nil, errors.Wrap(err, "updating title")
		}
	}
	if args.Grants != nil {
		err := tx.Exec(ctx, sqlf.Sprintf(removeDashboardGrants,
			args.ID,
		))
		if err != nil {
			return nil, errors.Wrap(err, "removing existing dashboard grants")
		}
		err = tx.AddDashboardGrants(ctx, args.ID, args.Grants)
		if err != nil {
			return nil, errors.Wrap(err, "AddDashboardGrants")
		}
	}
	dashboards, err := tx.GetDashboards(ctx, DashboardQueryArgs{ID: args.ID, UserID: args.UserID, OrgID: args.OrgID})
	if err != nil {
		return nil, errors.Wrap(err, "GetDashboards")
	}
	if len(dashboards) > 0 {
		return dashboards[0], nil
	}
	return nil, nil
}

func (s *DBDashboardStore) AddViewsToDashboard(ctx context.Context, dashboardId int, viewIds []string) error {
	if dashboardId == 0 {
		return errors.New("unable to associate views to dashboard invalid dashboard ID")
	} else if len(viewIds) == 0 {
		return nil
	}
	q := sqlf.Sprintf(insertDashboardInsightViewConnectionsByViewIds, dashboardId, pq.Array(viewIds))
	err := s.Exec(ctx, q)
	if err != nil {
		return err
	}
	return nil
}

func (s *DBDashboardStore) RemoveViewsFromDashboard(ctx context.Context, dashboardId int, viewIds []string) error {
	if dashboardId == 0 {
		return errors.New("unable to remove views from dashboard invalid dashboard ID")
	} else if len(viewIds) == 0 {
		return nil
	}
	q := sqlf.Sprintf(removeDashboardInsightViewConnectionsByViewIds, dashboardId, pq.Array(viewIds))
	err := s.Exec(ctx, q)
	if err != nil {
		return err
	}
	return nil
}

func (s *DBDashboardStore) IsViewOnDashboard(ctx context.Context, dashboardId int, viewId string) (bool, error) {
	count, _, err := basestore.ScanFirstInt(s.Query(ctx, sqlf.Sprintf(getViewFromDashboardByViewId, dashboardId, viewId)))
	return count != 0, err
}

func (s *DBDashboardStore) GetDashboardGrants(ctx context.Context, dashboardId int) ([]*DashboardGrant, error) {
	return scanDashboardGrants(s.Query(ctx, sqlf.Sprintf(getDashboardGrantsSql, dashboardId)))
}

func (s *DBDashboardStore) HasDashboardPermission(ctx context.Context, dashboardId int, userIds []int, orgIds []int) (bool, error) {
	query := sqlf.Sprintf(getDashboardGrantsByPermissionsSql, dashboardId, dashboardPermissionsQuery(DashboardQueryArgs{UserID: userIds, OrgID: orgIds}))
	count, _, err := basestore.ScanFirstInt(s.Query(ctx, query))
	return count != 0, err
}

func (s *DBDashboardStore) AddDashboardGrants(ctx context.Context, dashboardId int, grants []DashboardGrant) error {
	if dashboardId == 0 {
		return errors.New("unable to grant dashboard permissions invalid dashboard id")
	} else if len(grants) == 0 {
		return nil
	}

	values := make([]*sqlf.Query, 0, len(grants))
	for _, grant := range grants {
		grantQuery, err := grant.toQuery(dashboardId)
		if err != nil {
			return err
		}
		values = append(values, grantQuery)
	}
	q := sqlf.Sprintf(addDashboardGrantsSql, sqlf.Join(values, ",\n"))
	err := s.Exec(ctx, q)
	if err != nil {
		return err
	}
	return nil
}

const insertDashboardSql = `
-- source: enterprise/internal/insights/store/dashboard_store.go:CreateDashboard
INSERT INTO dashboard (title, save) VALUES (%s, %s) RETURNING id;
`

const insertDashboardInsightViewConnectionsByViewIds = `
-- source: enterprise/internal/insights/store/dashboard_store.go:AddViewsToDashboard
INSERT INTO dashboard_insight_view (dashboard_id, insight_view_id) (
    SELECT %s AS dashboard_id, insight_view.id AS insight_view_id
    FROM insight_view
    WHERE unique_id = ANY(%s)
)
ON CONFLICT DO NOTHING;
`

const updateDashboardSql = `
-- source: enterprise/internal/insights/store/dashboard_store.go:UpdateDashboard
UPDATE dashboard SET title = %s WHERE id = %s;
`

const removeDashboardGrants = `
-- source: enterprise/internal/insights/store/dashboard_store.go:removeDashboardGrants
delete from dashboard_grants where dashboard_id = %s;
`

const removeDashboardInsightViewConnectionsByViewIds = `
-- source: enterprise/internal/insights/store/dashboard_store.go:RemoveViewsFromDashboard
DELETE
FROM dashboard_insight_view
WHERE dashboard_id = %s
  AND insight_view_id IN (SELECT id FROM insight_view WHERE unique_id = ANY(%s));
`

const getViewFromDashboardByViewId = `
-- source: enterprise/internal/insights/store/insight_store.go:GetViewFromDashboardByViewId
SELECT COUNT(*)
FROM dashboard_insight_view div
	INNER JOIN insight_view iv ON div.insight_view_id = iv.id
WHERE div.dashboard_id = %s AND iv.unique_id = %s
`

const getDashboardGrantsSql = `
-- source: enterprise/internal/insights/store/insight_store.go:GetDashboardGrants
SELECT * FROM dashboard_grants where dashboard_id = %s
`

const getDashboardGrantsByPermissionsSql = `
-- source: enterprise/internal/insights/store/insight_store.go:GetDashboardGrants
SELECT COUNT(*) FROM dashboard_grants as dg
WHERE dg.dashboard_id = %s AND %s
`

const addDashboardGrantsSql = `
-- source: enterprise/internal/insights/store/insight_store.go:AddDashboardGrants
INSERT INTO dashboard_grants (dashboard_id, user_id, org_id, global)
VALUES %s;
`

type DashboardStore interface {
	GetDashboards(ctx context.Context, args DashboardQueryArgs) ([]*types.Dashboard, error)
	CreateDashboard(ctx context.Context, args CreateDashboardArgs) (_ *types.Dashboard, err error)
	UpdateDashboard(ctx context.Context, args UpdateDashboardArgs) (_ *types.Dashboard, err error)
	DeleteDashboard(ctx context.Context, id int64) error
	HasDashboardPermission(ctx context.Context, dashboardId int, userIds []int, orgIds []int) (bool, error)
}

package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"

	edb "github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type DBDashboardStore struct {
	*basestore.Store
	Now func() time.Time
}

// NewDashboardStore returns a new DBDashboardStore backed by the given Postgres db.
func NewDashboardStore(db edb.InsightsDB) *DBDashboardStore {
	return &DBDashboardStore{Store: basestore.NewWithHandle(db.Handle()), Now: time.Now}
}

// With creates a new DBDashboardStore with the given basestore. Shareable store as the underlying basestore.Store.
// Needed to implement the basestore.Store interface
func (s *DBDashboardStore) With(other basestore.ShareableStore) *DBDashboardStore {
	return &DBDashboardStore{Store: s.Store.With(other), Now: s.Now}
}

func (s *DBDashboardStore) Transact(ctx context.Context) (*DBDashboardStore, error) {
	txBase, err := s.Store.Transact(ctx)
	return &DBDashboardStore{Store: txBase, Now: s.Now}, err
}

type DashboardType string

const (
	Standard DashboardType = "standard"
	// This is a singleton dashboard that facilitates users having global access to their insights in Limited Access Mode.
	LimitedAccessMode DashboardType = "limited_access_mode"
)

type DashboardQueryArgs struct {
	UserIDs          []int
	OrgIDs           []int
	IDs              []int
	WithViewUniqueID *string
	Deleted          bool
	Limit            int
	After            int

	// This field will disable user level authorization checks on the dashboards. This should only be used interally,
	// and not to return dashboards to users.
	WithoutAuthorization bool
}

func (s *DBDashboardStore) GetDashboards(ctx context.Context, args DashboardQueryArgs) ([]*types.Dashboard, error) {
	preds := make([]*sqlf.Query, 0, 1)
	if len(args.IDs) > 0 {
		elems := make([]*sqlf.Query, 0, len(args.IDs))
		for _, id := range args.IDs {
			elems = append(elems, sqlf.Sprintf("%s", id))
		}
		preds = append(preds, sqlf.Sprintf("db.id in (%s)", sqlf.Join(elems, ",")))
	}
	if args.Deleted {
		preds = append(preds, sqlf.Sprintf("db.deleted_at is not null"))
	} else {
		preds = append(preds, sqlf.Sprintf("db.deleted_at is null"))
	}
	if args.After > 0 {
		preds = append(preds, sqlf.Sprintf("db.id > %s", args.After))
	}
	if args.WithViewUniqueID != nil {
		preds = append(preds, sqlf.Sprintf("%s = ANY(t.uuid_array)", *args.WithViewUniqueID))
	}

	if !args.WithoutAuthorization {
		preds = append(preds, sqlf.Sprintf("db.id in (%s)", visibleDashboardsQuery(args.UserIDs, args.OrgIDs)))
	}
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

func (s *DBDashboardStore) DeleteDashboard(ctx context.Context, id int) error {
	err := s.Exec(ctx, sqlf.Sprintf(deleteDashboardSql, id))
	if err != nil {
		return errors.Wrapf(err, "failed to delete dashboard with id: %s", id)
	}
	return nil
}

func (s *DBDashboardStore) RestoreDashboard(ctx context.Context, id int) error {
	err := s.Exec(ctx, sqlf.Sprintf(restoreDashboardSql, id))
	if err != nil {
		return errors.Wrapf(err, "failed to restore dashboard with id: %s", id)
	}
	return nil
}

// visibleDashboardsQuery generates the SQL query for filtering dashboards based on granted permissions.
// This returns a query that will generate a set of dashboard.id that the provided context can see.
func visibleDashboardsQuery(userIDs, orgIDs []int) *sqlf.Query {
	permsPreds := make([]*sqlf.Query, 0, 2)
	if len(orgIDs) > 0 {
		elems := make([]*sqlf.Query, 0, len(orgIDs))
		for _, id := range orgIDs {
			elems = append(elems, sqlf.Sprintf("%s", id))
		}
		permsPreds = append(permsPreds, sqlf.Sprintf("org_id IN (%s)", sqlf.Join(elems, ",")))
	}
	if len(userIDs) > 0 {
		elems := make([]*sqlf.Query, 0, len(userIDs))
		for _, id := range userIDs {
			elems = append(elems, sqlf.Sprintf("%s", id))
		}
		permsPreds = append(permsPreds, sqlf.Sprintf("user_id IN (%s)", sqlf.Join(elems, ",")))
	}
	permsPreds = append(permsPreds, sqlf.Sprintf("global is true"))
	return sqlf.Sprintf("SELECT dashboard_id FROM dashboard_grants WHERE %s", sqlf.Join(permsPreds, "OR"))
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
update dashboard set deleted_at = NOW() where id = %s;
`

const restoreDashboardSql = `
update dashboard set deleted_at = NULL where id = %s;
`

type CreateDashboardArgs struct {
	Dashboard types.Dashboard
	Grants    []DashboardGrant
	UserIDs   []int // For dashboard permissions
	OrgIDs    []int // For dashboard permissions
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
		Standard,
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

	dashboards, err := tx.GetDashboards(ctx, DashboardQueryArgs{IDs: []int{dashboardId}, UserIDs: args.UserIDs, OrgIDs: args.OrgIDs})
	if err != nil {
		return nil, errors.Wrap(err, "GetDashboards")
	}
	if len(dashboards) > 0 {
		return dashboards[0], nil
	}
	return nil, nil
}

type UpdateDashboardArgs struct {
	ID      int
	Title   *string
	Grants  []DashboardGrant
	UserIDs []int // For dashboard permissions
	OrgIDs  []int // For dashboard permissions
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
	dashboards, err := tx.GetDashboards(ctx, DashboardQueryArgs{IDs: []int{args.ID}, UserIDs: args.UserIDs, OrgIDs: args.OrgIDs})
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

	// Create rows for an inline table which is used to preserve the ordering of the viewIds.
	orderings := make([]*sqlf.Query, 0, 1)
	for i, viewId := range viewIds {
		orderings = append(orderings, sqlf.Sprintf("(%s, %s)", viewId, fmt.Sprintf("%d", i)))
	}

	q := sqlf.Sprintf(insertDashboardInsightViewConnectionsByViewIds, dashboardId, sqlf.Join(orderings, ","), pq.Array(viewIds))
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

func (s *DBDashboardStore) HasDashboardPermission(ctx context.Context, dashboardIds []int, userIds []int, orgIds []int) (bool, error) {
	query := sqlf.Sprintf(getDashboardGrantsByPermissionsSql, pq.Array(dashboardIds), visibleDashboardsQuery(userIds, orgIds))
	count, _, err := basestore.ScanFirstInt(s.Query(ctx, query))
	return count == len(dashboardIds), err
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

func (s *DBDashboardStore) EnsureLimitedAccessModeDashboard(ctx context.Context) (_ int, err error) {
	tx, err := s.Transact(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { err = tx.Done(err) }()

	id, _, err := basestore.ScanFirstInt(tx.Query(ctx, sqlf.Sprintf("SELECT id FROM dashboard WHERE type = %s", LimitedAccessMode)))
	if err != nil {
		return 0, err
	}
	if id == 0 {
		query := sqlf.Sprintf(insertDashboardSql, "Limited Access Mode Dashboard", true, LimitedAccessMode)
		id, _, err = basestore.ScanFirstInt(tx.Query(ctx, query))
		if err != nil {
			return 0, err
		}
		global := true
		err = tx.AddDashboardGrants(ctx, id, []DashboardGrant{{Global: &global}})
		if err != nil {
			return 0, err
		}
	}
	// This dashboard may have been previously deleted.
	tx.RestoreDashboard(ctx, id)
	if err != nil {
		return 0, errors.Wrap(err, "RestoreDashboard")
	}
	return id, nil
}

const insertDashboardSql = `
INSERT INTO dashboard (title, save, type) VALUES (%s, %s, %s) RETURNING id;
`

const insertDashboardInsightViewConnectionsByViewIds = `
INSERT INTO dashboard_insight_view (dashboard_id, insight_view_id) (
    SELECT %s AS dashboard_id, insight_view.id AS insight_view_id
    FROM insight_view
		JOIN
			( VALUES %s) as ids (id, ordering)
		ON ids.id = insight_view.unique_id
    WHERE unique_id = ANY(%s)
	ORDER BY ids.ordering
) ON CONFLICT DO NOTHING;
`
const updateDashboardSql = `
UPDATE dashboard SET title = %s WHERE id = %s;
`

const removeDashboardGrants = `
delete from dashboard_grants where dashboard_id = %s;
`

const removeDashboardInsightViewConnectionsByViewIds = `
DELETE
FROM dashboard_insight_view
WHERE dashboard_id = %s
  AND insight_view_id IN (SELECT id FROM insight_view WHERE unique_id = ANY(%s));
`

const getViewFromDashboardByViewId = `
SELECT COUNT(*)
FROM dashboard_insight_view div
	INNER JOIN insight_view iv ON div.insight_view_id = iv.id
WHERE div.dashboard_id = %s AND iv.unique_id = %s
`

const getDashboardGrantsSql = `
SELECT id, dashboard_id, user_id, org_id, global FROM dashboard_grants where dashboard_id = %s
`

const getDashboardGrantsByPermissionsSql = `
SELECT count(*)
FROM dashboard
WHERE id = ANY (%s)
AND id IN (%s);
`

const addDashboardGrantsSql = `
INSERT INTO dashboard_grants (dashboard_id, user_id, org_id, global)
VALUES %s;
`

type DashboardStore interface {
	GetDashboards(ctx context.Context, args DashboardQueryArgs) ([]*types.Dashboard, error)
	CreateDashboard(ctx context.Context, args CreateDashboardArgs) (_ *types.Dashboard, err error)
	UpdateDashboard(ctx context.Context, args UpdateDashboardArgs) (_ *types.Dashboard, err error)
	DeleteDashboard(ctx context.Context, id int) error
	RestoreDashboard(ctx context.Context, id int) error
	HasDashboardPermission(ctx context.Context, dashboardId []int, userIds []int, orgIds []int) (bool, error)
}

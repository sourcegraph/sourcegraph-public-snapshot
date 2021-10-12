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

	q := sqlf.Sprintf(getDashboardSql, sqlf.Join(preds, "\n AND"), limitClause)
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
		); err != nil {
			return []*types.Dashboard{}, err
		}
		results = append(results, &temp)
	}
	return results, nil
}

const getDashboardSql = `
-- source: enterprise/internal/insights/store/dashboard_store.go:GetDashboards
SELECT db.id, db.title, t.uuid_array as insight_view_unique_ids
FROM dashboard db
         JOIN dashboard_grants dg ON db.id = dg.dashboard_id
         LEFT JOIN (SELECT ARRAY_AGG(iv.unique_id) AS uuid_array, div.dashboard_id
               FROM insight_view iv
                        JOIN dashboard_insight_view div ON iv.id = div.insight_view_id
               GROUP BY div.dashboard_id) t on t.dashboard_id = db.id
WHERE %S
ORDER BY db.id
%S;
`

const deleteDashboardSql = `
-- source: enterprise/internal/insights/store/dashboard_store.go:DeleteDashboard
update dashboard set deleted_at = NOW() where id = %s;
`

func (s *DBDashboardStore) CreateDashboard(ctx context.Context, dashboard types.Dashboard, grants []DashboardGrant) (_ types.Dashboard, err error) {
	tx, err := s.Transact(ctx)
	if err != nil {
		return types.Dashboard{}, err
	}
	defer func() { err = tx.Done(err) }()

	row := tx.QueryRow(ctx, sqlf.Sprintf(insertDashboardSql,
		dashboard.Title,
		dashboard.Save,
	))
	if row.Err() != nil {
		return types.Dashboard{}, row.Err()
	}
	var id int
	err = row.Scan(&id)
	if err != nil {
		return types.Dashboard{}, errors.Wrap(err, "CreateDashboard")
	}
	dashboard.ID = id
	err = tx.AddViewsToDashboard(ctx, dashboard.ID, dashboard.InsightIDs)
	if err != nil {
		return types.Dashboard{}, errors.Wrap(err, "AddViewsToDashboard")
	}
	err = tx.AddDashboardGrants(ctx, dashboard, grants)
	if err != nil {
		return types.Dashboard{}, errors.Wrap(err, "AddDashboardGrants")
	}

	return dashboard, nil
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

func (s *DBDashboardStore) AddDashboardGrants(ctx context.Context, dashboard types.Dashboard, grants []DashboardGrant) error {
	if dashboard.ID == 0 {
		return errors.New("unable to grant dashboard permissions invalid dashboard id")
	} else if len(grants) == 0 {
		return nil
	}

	values := make([]*sqlf.Query, 0, len(grants))
	for _, grant := range grants {
		grantQuery, err := grant.toQuery(dashboard.ID)
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

const addDashboardGrantsSql = `
-- source: enterprise/internal/insights/store/insight_store.go:AddDashboardGrants
INSERT INTO dashboard_grants (dashboard_id, org_id, user_id, global)
VALUES %s;
`

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
);`

const removeDashboardInsightViewConnectionsByViewIds = `
-- source: enterprise/internal/insights/store/dashboard_store.go:RemoveViewsFromDashboard
DELETE
FROM dashboard_insight_view
WHERE dashboard_id = %s
  AND insight_view_id IN (SELECT id FROM insight_view WHERE unique_id = ANY(%s));
`

type DashboardStore interface {
	GetDashboards(ctx context.Context, args DashboardQueryArgs) ([]*types.Dashboard, error)
	CreateDashboard(ctx context.Context, dashboard types.Dashboard, grants []DashboardGrant) (_ types.Dashboard, err error)
	DeleteDashboard(ctx context.Context, id int64) error
}

package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

type DashboardStore struct {
	*basestore.Store
	Now func() time.Time
}

// NewDashboardStore returns a new DashboardStore backed by the given Timescale db.
func NewDashboardStore(db dbutil.DB) *DashboardStore {
	return &DashboardStore{Store: basestore.NewWithDB(db, sql.TxOptions{}), Now: time.Now}
}

// Handle returns the underlying transactable database handle.
// Needed to implement the ShareableStore interface.
func (s *DashboardStore) Handle() *basestore.TransactableHandle { return s.Store.Handle() }

// With creates a new DashboardStore with the given basestore. Shareable store as the underlying basestore.Store.
// Needed to implement the basestore.Store interface
func (s *DashboardStore) With(other *DashboardStore) *DashboardStore {
	return &DashboardStore{Store: s.Store.With(other.Store), Now: other.Now}
}

func (s *DashboardStore) Transact(ctx context.Context) (*DashboardStore, error) {
	txBase, err := s.Store.Transact(ctx)
	return &DashboardStore{Store: txBase, Now: s.Now}, err
}

type DashboardQueryArgs struct {
	UserID  []int
	OrgID   []int
	ID      int
	Deleted bool
}

func (s *DashboardStore) GetDashboards(ctx context.Context, args DashboardQueryArgs) ([]*types.Dashboard, error) {
	preds := make([]*sqlf.Query, 0, 1)
	if args.ID > 0 {
		preds = append(preds, sqlf.Sprintf("id = %s", args.ID))
	}
	if args.Deleted {
		preds = append(preds, sqlf.Sprintf("deleted_at is not null"))
	} else {
		preds = append(preds, sqlf.Sprintf("deleted_at is null"))
	}
	preds = append(preds, dashboardPermissionsQuery(args))
	if len(preds) == 0 {
		preds = append(preds, sqlf.Sprintf("%s", "TRUE"))
	}

	q := sqlf.Sprintf(getDashboardSql, sqlf.Join(preds, "\n AND"))
	return scanDashboard(s.Query(ctx, q))
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

	results := make([]*types.Dashboard, 0)
	for rows.Next() {
		var temp types.Dashboard
		if err := rows.Scan(
			&temp.ID,
			&temp.Title,
		); err != nil {
			return []*types.Dashboard{}, err
		}
		results = append(results, &temp)
	}
	return results, nil
}

const getDashboardSql = `
-- source: enterprise/internal/insights/store/dashboard_store.go:Get
SELECT db.id, db.title
FROM dashboard db
JOIN dashboard_grants dg ON db.id = dg.dashboard_id
WHERE %S;
`

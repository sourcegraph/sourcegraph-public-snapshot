package store

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"

	"github.com/gosimple/slug"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/tenant"
)

type Tenant struct {
	ID        int64
	Name      string
	Slug      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type TenantNotFoundError struct {
	error
}

func (e TenantNotFoundError) NotFound() bool {
	return true
}

type ListTenantsOptions struct {
	Cursor string
	Limit  int
}

type Store interface {
	CreateTenant(ctx context.Context, name string) (*Tenant, error)
	// If the tenant doesn't exist, a TenantNotFoundError is returned.
	GetByName(ctx context.Context, name string) (*Tenant, error)
	// If the tenant doesn't exist, a TenantNotFoundError is returned.
	DeleteTenant(ctx context.Context, id int64) error
	ListTenants(ctx context.Context, opts ListTenantsOptions) (_ []*Tenant, nextCursor string, _ error)
}

func New(db dbutil.DB) Store {
	return &store{db: db}
}

type store struct {
	db dbutil.DB
}

var tenantScanRows = []*sqlf.Query{
	sqlf.Sprintf("id"),
	sqlf.Sprintf("name"),
	sqlf.Sprintf("slug"),
	sqlf.Sprintf("created_at"),
	sqlf.Sprintf("updated_at"),
}

const createTenantQueryFmtstr = `
INSERT INTO
	tenants (name, slug)
VALUES
	(%s, %s)
RETURNING
	%s
`

func (s *store) CreateTenant(ctx context.Context, name string) (*Tenant, error) {
	slg := slug.Make(name)
	if slg == "" {
		return nil, errors.New("slug cannot be empty after slugifying")
	}

	q := sqlf.Sprintf(createTenantQueryFmtstr, name, slg, sqlf.Join(tenantScanRows, ", "))
	row := s.db.QueryRowContext(tenant.InsecureGlobalContext(ctx), q.Query(sqlf.PostgresBindVar), q.Args()...)

	var t Tenant

	// TODO:
	// -- INSERT INTO "public"."roles"("created_at", "system", "name", "tenant_id") VALUES('2023-01-04 17:29:41.195966+01', 'TRUE', 'USER', 2) RETURNING "id", "created_at", "system", "name", "tenant_id";
	// -- INSERT INTO "public"."roles"("created_at", "system", "name", "tenant_id") VALUES('2023-01-04 17:29:41.195966+01', 'TRUE', 'SITE_ADMINISTRATOR', 2) RETURNING "id", "created_at", "system", "name", "tenant_id";
	// 	INSERT INTO "public"."own_signal_configurations"("name", "description", "enabled", "tenant_id") VALUES('recent-contributors', 'Indexes contributors in each file using repository history.', 'FALSE', 2) RETURNING "id", "name", "description", "excluded_repo_patterns", "enabled", "tenant_id";
	// INSERT INTO "public"."own_signal_configurations"("name", "description", "enabled", "tenant_id") VALUES('recent-views', 'Indexes users that recently viewed files in Sourcegraph.', 'FALSE', 2) RETURNING "id", "name", "description", "excluded_repo_patterns", "enabled", "tenant_id";
	// INSERT INTO "public"."own_signal_configurations"("name", "description", "enabled", "tenant_id") VALUES('analytics', 'Indexes ownership data to present in aggregated views like Admin > Analytics > Own and Repo > Ownership', 'FALSE', 2) RETURNING "id", "name", "description", "excluded_repo_patterns", "enabled", "tenant_id";

	return &t, row.Scan(&t.ID, &t.Name, &t.Slug, &t.CreatedAt, &t.UpdatedAt)
}

const getTenantByNameQueryFmtstr = `
SELECT
	%s
FROM
	tenants
WHERE
	name = %s
`

func (s *store) GetByName(ctx context.Context, name string) (*Tenant, error) {
	q := sqlf.Sprintf(getTenantByNameQueryFmtstr, sqlf.Join(tenantScanRows, ", "), name)
	row := s.db.QueryRowContext(tenant.InsecureGlobalContext(ctx), q.Query(sqlf.PostgresBindVar), q.Args()...)

	var t Tenant

	err := row.Scan(&t.ID, &t.Name, &t.Slug, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &TenantNotFoundError{errors.New("not found")}
		}
		return nil, err
	}

	return &t, nil
}

// TODO: ERROR: insert or update on table "repo_statistics" violates foreign key constraint "repo_statistics_tenant_id_fkey" (SQLSTATE 23503)
const deleteTenantQueryFmtstr = `
DELETE FROM
	tenants
CASCADE
WHERE
	id = %s
`

func (s *store) DeleteTenant(ctx context.Context, id int64) error {
	q := sqlf.Sprintf(deleteTenantQueryFmtstr, id)
	res, err := s.db.ExecContext(tenant.InsecureGlobalContext(ctx), q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return err
	}

	aff, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if aff == 0 {
		return &TenantNotFoundError{errors.New("not found")}
	}

	return nil
}

const listTenantsQueryFmtstr = `
SELECT
	%s
FROM
	tenants
WHERE
	%s
ORDER BY
	id ASC
LIMIT %s
`

func (s *store) ListTenants(ctx context.Context, opts ListTenantsOptions) (_ []*Tenant, nextCursor string, err error) {
	conds := []*sqlf.Query{sqlf.Sprintf("TRUE")}

	if opts.Cursor != "" {
		cur, err := strconv.Atoi(opts.Cursor)
		if err != nil {
			return nil, "", err
		}
		conds = append(conds, sqlf.Sprintf("id >= %s", cur))
	}

	q := sqlf.Sprintf(
		listTenantsQueryFmtstr,
		sqlf.Join(tenantScanRows, ", "),
		sqlf.Join(conds, " AND "),
		opts.Limit+1,
	)

	rows, err := s.db.QueryContext(tenant.InsecureGlobalContext(ctx), q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return nil, "", err
	}
	defer func() {
		err = basestore.CloseRows(rows, err)
	}()

	var tenants []*Tenant
	for rows.Next() {
		var t Tenant
		if err := rows.Scan(&t.ID, &t.Name, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, "", err
		}
		tenants = append(tenants, &t)
	}

	if len(tenants) == opts.Limit+1 {
		nextCursor = strconv.Itoa(int(tenants[len(tenants)-1].ID))
		tenants = tenants[:len(tenants)-1]
	}

	return tenants, nextCursor, nil
}

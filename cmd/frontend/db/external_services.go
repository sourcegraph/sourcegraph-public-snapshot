package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/dbconn"
)

type externalServices struct{}

// ExternalServicesListOptions contains options for listing external services.
type ExternalServicesListOptions struct {
	Kind string
	*LimitOffset
}

func (o ExternalServicesListOptions) sqlConditions() []*sqlf.Query {
	conds := []*sqlf.Query{sqlf.Sprintf("deleted_at IS NULL")}
	if o.Kind != "" {
		conds = append(conds, sqlf.Sprintf("kind=%s", o.Kind))
	}
	return conds
}

// Create creates a external service.
//
// 🚨 SECURITY: The caller must ensure that the actor is a site admin.
func (c *externalServices) Create(ctx context.Context, externalService *types.ExternalService) error {
	externalService.CreatedAt = time.Now()
	externalService.UpdatedAt = externalService.CreatedAt
	return dbconn.Global.QueryRowContext(
		ctx,
		"INSERT INTO external_services(kind, display_name, config, created_at, updated_at) VALUES($1, $2, $3, $4, $5) RETURNING id",
		externalService.Kind, externalService.DisplayName, externalService.Config, externalService.CreatedAt, externalService.UpdatedAt,
	).Scan(&externalService.ID)
}

// ExternalServiceUpdate contains optional fields to update.
type ExternalServiceUpdate struct {
	DisplayName *string
	Config      *string
}

// Update updates a external service.
//
// 🚨 SECURITY: The caller must ensure that the actor is a site admin.
func (c *externalServices) Update(ctx context.Context, id int64, update *ExternalServiceUpdate) error {
	execUpdate := func(ctx context.Context, tx *sql.Tx, update *sqlf.Query) error {
		q := sqlf.Sprintf("UPDATE external_services SET %s, updated_at=now() WHERE id=%d AND deleted_at IS NULL", update, id)
		res, err := tx.ExecContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
		if err != nil {
			return err
		}
		affected, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if affected == 0 {
			return externalServiceNotFoundError{id: id}
		}
		return nil
	}
	return Transaction(ctx, dbconn.Global, func(tx *sql.Tx) error {
		if update.DisplayName != nil {
			if err := execUpdate(ctx, tx, sqlf.Sprintf("display_name=%s", update.DisplayName)); err != nil {
				return err
			}
		}
		if update.Config != nil {
			if err := execUpdate(ctx, tx, sqlf.Sprintf("config=%s", update.Config)); err != nil {
				return err
			}
		}
		return nil
	})
}

type externalServiceNotFoundError struct {
	id int64
}

func (e externalServiceNotFoundError) Error() string {
	return fmt.Sprintf("external service not found: %v", e.id)
}

func (e externalServiceNotFoundError) NotFound() bool {
	return true
}

// Delete deletes an external service.
//
// 🚨 SECURITY: The caller must ensure that the actor is a site admin.
func (*externalServices) Delete(ctx context.Context, id int64) error {
	res, err := dbconn.Global.ExecContext(ctx, "UPDATE external_services SET deleted_at=now() WHERE id=$1 AND deleted_at IS NULL", id)
	if err != nil {
		return err
	}
	nrows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if nrows == 0 {
		return externalServiceNotFoundError{id: id}
	}
	return nil
}

// GetByID returns the external service for id.
//
// 🚨 SECURITY: The caller must ensure that the actor is a site admin.
func (c *externalServices) GetByID(ctx context.Context, id int64) (*types.ExternalService, error) {
	conds := []*sqlf.Query{sqlf.Sprintf("id=%d", id)}
	externalServices, err := c.list(ctx, conds, nil)
	if err != nil {
		return nil, err
	}
	if len(externalServices) == 0 {
		return nil, fmt.Errorf("external service not found: id=%d", id)
	}
	return externalServices[0], nil
}

// List returns all external services.
//
// 🚨 SECURITY: The caller must ensure that the actor is a site admin.
func (c *externalServices) List(ctx context.Context, opt ExternalServicesListOptions) ([]*types.ExternalService, error) {
	return c.list(ctx, opt.sqlConditions(), opt.LimitOffset)
}

func (c *externalServices) list(ctx context.Context, conds []*sqlf.Query, limitOffset *LimitOffset) ([]*types.ExternalService, error) {
	q := sqlf.Sprintf(`
		SELECT id, kind, display_name, config, created_at, updated_at
		FROM external_services
		WHERE (%s)
		ORDER BY id DESC
		%s`,
		sqlf.Join(conds, ") AND ("),
		limitOffset.SQL(),
	)

	rows, err := dbconn.Global.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*types.ExternalService
	for rows.Next() {
		var h types.ExternalService
		if err := rows.Scan(&h.ID, &h.Kind, &h.DisplayName, &h.Config, &h.CreatedAt, &h.UpdatedAt); err != nil {
			return nil, err
		}
		results = append(results, &h)
	}
	return results, nil
}

// Count counts all access tokens that satisfy the options (ignoring limit and offset).
//
// 🚨 SECURITY: The caller must ensure that the actor is a site admin.
func (c *externalServices) Count(ctx context.Context, opt ExternalServicesListOptions) (int, error) {
	q := sqlf.Sprintf("SELECT COUNT(*) FROM external_services WHERE (%s)", sqlf.Join(opt.sqlConditions(), ") AND ("))
	var count int
	if err := dbconn.Global.QueryRowContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

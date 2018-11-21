package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db/dbconn"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
)

type codehosts struct{}

// CodehostsListOptions contains options for listing code hosts.
type CodehostsListOptions struct {
	*LimitOffset
}

func (o CodehostsListOptions) sqlConditions() []*sqlf.Query {
	return []*sqlf.Query{sqlf.Sprintf("deleted_at IS NULL")}
}

// Create creates a codehost.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to create codehosts.
func (c *codehosts) Create(ctx context.Context, codehost *types.Codehost) error {
	codehost.CreatedAt = time.Now()
	codehost.UpdatedAt = codehost.CreatedAt
	return dbconn.Global.QueryRowContext(
		ctx,
		"INSERT INTO codehosts(kind, display_name, config, created_at, updated_at) VALUES($1, $2, $3, $4, $5) RETURNING id",
		codehost.Kind, codehost.DisplayName, codehost.Config, codehost.CreatedAt, codehost.UpdatedAt).Scan(&codehost.ID)
}

// CodehostUpdate contains optional fields to update.
type CodehostUpdate struct {
	DisplayName *string
	Config      *string
}

// Update updates a codehost.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to update codehosts.
func (c *codehosts) Update(ctx context.Context, id int64, update *CodehostUpdate) error {
	execUpdate := func(ctx context.Context, tx *sql.Tx, update *sqlf.Query) error {
		q := sqlf.Sprintf("UPDATE codehosts SET %s, updated_at=now() WHERE id=%d AND deleted_at IS NULL", update, id)
		res, err := tx.ExecContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
		if err != nil {
			return err
		}
		affected, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if affected == 0 {
			return errors.New("no rows updated")
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

// GetByID returns the codehost for id.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to read codehosts.
func (c *codehosts) GetByID(ctx context.Context, id int64) (*types.Codehost, error) {
	conds := []*sqlf.Query{sqlf.Sprintf("id=%d", id)}
	codehosts, err := c.list(ctx, conds, nil)
	if err != nil {
		return nil, err
	}
	if len(codehosts) == 0 {
		return nil, fmt.Errorf("codehost not found: id=%d", id)
	}
	return codehosts[0], nil
}

// List returns all codehost connections.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to list codehosts.
func (c *codehosts) List(ctx context.Context, opt CodehostsListOptions) ([]*types.Codehost, error) {
	return c.list(ctx, opt.sqlConditions(), opt.LimitOffset)
}

func (c *codehosts) list(ctx context.Context, conds []*sqlf.Query, limitOffset *LimitOffset) ([]*types.Codehost, error) {
	q := sqlf.Sprintf(`
SELECT id, kind, display_name, config, created_at, updated_at 
FROM codehosts
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

	var results []*types.Codehost
	for rows.Next() {
		var h types.Codehost
		if err := rows.Scan(&h.ID, &h.Kind, &h.DisplayName, &h.Config, &h.CreatedAt, &h.UpdatedAt); err != nil {
			return nil, err
		}
		results = append(results, &h)
	}
	return results, nil
}

// Count counts all access tokens that satisfy the options (ignoring limit and offset).
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to count codehosts.
func (c *codehosts) Count(ctx context.Context, opt CodehostsListOptions) (int, error) {
	q := sqlf.Sprintf("SELECT COUNT(*) FROM codehosts WHERE (%s)", sqlf.Join(opt.sqlConditions(), ") AND ("))
	var count int
	if err := dbconn.Global.QueryRowContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

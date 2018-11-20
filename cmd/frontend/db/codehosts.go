package db

import (
	"context"
	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db/dbconn"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"time"
)

type codehosts struct{}

// CodehostsListOptions contains options for listing code hosts.
type CodehostsListOptions struct {
	*LimitOffset
}

func (o CodehostsListOptions) sqlConditions() []*sqlf.Query {
	return []*sqlf.Query{sqlf.Sprintf("deleted_at IS NULL")}
}

func (c *codehosts) Create(ctx context.Context, codehost *types.Codehost) error {
	codehost.CreatedAt = time.Now()
	codehost.UpdatedAt = codehost.CreatedAt
	return dbconn.Global.QueryRowContext(
		ctx,
		"INSERT INTO codehosts(kind, display_name, config, created_at, updated_at) VALUES($1, $2, $3, $4, $5) RETURNING id",
		codehost.Kind, codehost.DisplayName, codehost.Config, codehost.CreatedAt, codehost.UpdatedAt).Scan(&codehost.ID)
}

// List returns all codehost connections.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to list the codehost connections.
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
// 🚨 SECURITY: The caller must ensure that the actor is permitted to count the codehost connections.
func (c *codehosts) Count(ctx context.Context, opt CodehostsListOptions) (int, error) {
	q := sqlf.Sprintf("SELECT COUNT(*) FROM codehosts WHERE (%s)", sqlf.Join(opt.sqlConditions(), ") AND ("))
	var count int
	if err := dbconn.Global.QueryRowContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

package accessrequests

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type List struct {
	FArgs *FilterArgs
	PArgs *database.PaginationArgs

	Response []*types.AccessRequest
}

func (c *List) Execute(ctx context.Context, store *basestore.Store) error {
	if c.FArgs == nil {
		c.FArgs = &FilterArgs{}
	}
	where := c.FArgs.SQL()
	if c.PArgs == nil {
		c.PArgs = &database.PaginationArgs{}
	}
	p := c.PArgs.SQL()

	if p.Where != nil {
		where = append(where, p.Where)
	}

	query := sqlf.Sprintf(listQuery, sqlf.Join(columns, ","), sqlf.Join(where, ") AND ("))
	query = p.AppendOrderToQuery(query)
	query = p.AppendLimitToQuery(query)

	nodes, err := scanAccessRequests(store.Query(ctx, query))
	if err != nil {
		return err
	}

	c.Response = nodes
	return nil
}

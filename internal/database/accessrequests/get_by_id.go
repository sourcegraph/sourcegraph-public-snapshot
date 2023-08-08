package accessrequests

import (
	"context"
	"database/sql"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type GetByID struct {
	ID int32

	Response *types.AccessRequest
}

func (c *GetByID) Execute(ctx context.Context, store *basestore.Store) error {
	row := store.QueryRow(ctx, sqlf.Sprintf("SELECT %s FROM access_requests WHERE id = %s", sqlf.Join(columns, ","), c.ID))
	node, err := scanAccessRequest(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return &ErrNotFound{ID: c.ID}
		}
		return err
	}

	c.Response = node
	return nil
}

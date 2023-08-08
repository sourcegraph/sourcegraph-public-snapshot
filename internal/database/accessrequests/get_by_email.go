package accessrequests

import (
	"context"
	"database/sql"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type GetByEmail struct {
	Email string

	Response *types.AccessRequest
}

func (c *GetByEmail) Execute(ctx context.Context, store *basestore.Store) error {
	row := store.QueryRow(ctx, sqlf.Sprintf("SELECT %s FROM access_requests WHERE email = %s", sqlf.Join(columns, ","), c.Email))
	node, err := scanAccessRequest(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return &ErrNotFound{Email: c.Email}
		}
		return err
	}

	c.Response = node
	return nil
}

func (c *Store) GetByEmail(ctx context.Context, email string) (*types.AccessRequest, error) {
	command := &GetByEmail{
		Email: email,
	}

	if err := c.dbStore.ExecuteCommand(ctx, command); err != nil {
		return nil, err
	}

	return command.Response, nil
}

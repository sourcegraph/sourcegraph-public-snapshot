package postgresql

import (
	"context"
	"database/sql"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/runtime"
)

type Client struct {
	db *sql.DB
}

func NewClient(ctx context.Context, contract runtime.Contract) (*Client, error) {
	sqlDB, err := contract.PostgreSQL.OpenDatabase(ctx, "primary")
	if err != nil {
		return nil, errors.Wrap(err, "contract.GetPostgreSQLDB")
	}
	return &Client{sqlDB}, nil
}

func (c *Client) Ping(ctx context.Context) error {
	if err := c.db.PingContext(ctx); err != nil {
		return errors.Wrap(err, "sqlDB.PingContext")
	}

	if _, err := c.db.ExecContext(ctx, "SELECT current_user;"); err != nil {
		return errors.Wrap(err, "sqlDB.ExecContext SELECT current_user")
	}

	return nil
}

func (c *Client) Close() error { return c.db.Close() }

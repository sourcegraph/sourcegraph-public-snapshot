package dotcomdb

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Database struct {
	conn *pgx.Conn
}

// NewDatabase wraps a direct connection to the Sourcegraph.com database. It
// ONLY executes read queries, so the connection can (and should) be
// authenticated by a read-only user.
func NewDatabase(conn *pgx.Conn) *Database {
	return &Database{conn: conn}
}

func (d *Database) Ping(ctx context.Context) error {
	if err := d.conn.Ping(ctx); err != nil {
		return errors.Wrap(err, "sqlDB.PingContext")
	}
	if _, err := d.conn.Exec(ctx, "SELECT current_user;"); err != nil {
		return errors.Wrap(err, "sqlDB.Exec SELECT current_user")
	}
	return nil
}

func (d *Database) Close(ctx context.Context) error { return d.conn.Close(ctx) }

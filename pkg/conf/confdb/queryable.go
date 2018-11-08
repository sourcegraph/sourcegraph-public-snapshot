package confdb

import (
	"context"
	"database/sql"
)

// queryable allows us to reuse the same logic for certain operations both
// inside and outside an explicit transaction.
// TODO@ggilmore this is copied from cmd/frontend/db/settings.go - figure out if / how to
// share this.
type queryable interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

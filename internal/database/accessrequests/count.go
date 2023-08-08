package accessrequests

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
)

type Count struct {
	FArgs *FilterArgs

	Response int
}

func (c *Count) Execute(ctx context.Context, store *basestore.Store) error {
	query := sqlf.Sprintf("SELECT COUNT(*) FROM access_requests WHERE (%s)", sqlf.Join(c.FArgs.SQL(), ") AND ("))
	count, err := basestore.ScanInt(store.QueryRow(ctx, query))
	if err != nil {
		return err
	}

	c.Response = count
	return nil
}

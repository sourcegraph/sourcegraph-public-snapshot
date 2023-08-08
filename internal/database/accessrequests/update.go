package accessrequests

import (
	"context"
	"database/sql"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Update struct {
	AccessRequest *types.AccessRequest

	Response *types.AccessRequest
}

const updateQuery = `
UPDATE access_requests
SET status = %s, updated_at = NOW(), decision_by_user_id = %s
WHERE id = %s
RETURNING %s`

func (c *Update) Execute(ctx context.Context, store *basestore.Store) error {
	query := sqlf.Sprintf(updateQuery, c.AccessRequest.Status, *c.AccessRequest.DecisionByUserID, c.AccessRequest.ID, sqlf.Join(columns, ","))
	updated, err := scanAccessRequest(store.QueryRow(ctx, query))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &ErrNotFound{ID: c.AccessRequest.ID}
		}
		return errors.Wrap(err, "scanning access_request")
	}

	c.Response = updated
	return nil
}

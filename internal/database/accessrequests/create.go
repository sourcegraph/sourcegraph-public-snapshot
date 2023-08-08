package accessrequests

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Create struct {
	AccessRequest *types.AccessRequest

	Response *types.AccessRequest
}

const insertQuery = `
INSERT INTO access_requests (%s)
VALUES ( %s, %s, %s, %s )
RETURNING %s`

var insertColumns = []*sqlf.Query{
	sqlf.Sprintf("name"),
	sqlf.Sprintf("email"),
	sqlf.Sprintf("additional_info"),
	sqlf.Sprintf("status"),
}

func (c *Create) Execute(ctx context.Context, store *basestore.Store) error {
	err := store.WithTransact(ctx, func(tx *basestore.Store) error {
		// We don't allow adding a new request_access with an email address that has already been
		// verified by another user.
		userExistsQuery := sqlf.Sprintf("SELECT TRUE FROM user_emails WHERE email = %s AND verified_at IS NOT NULL", c.AccessRequest.Email)
		exists, _, err := basestore.ScanFirstBool(tx.Query(ctx, userExistsQuery))
		if err != nil {
			return err
		}
		if exists {
			return ErrCannotCreate{errorCodeUserWithEmailExists}
		}

		// We don't allow adding a new request_access with an email address that has already been used
		accessRequestsExistsQuery := sqlf.Sprintf("SELECT TRUE FROM access_requests WHERE email = %s", c.AccessRequest.Email)
		exists, _, err = basestore.ScanFirstBool(tx.Query(ctx, accessRequestsExistsQuery))
		if err != nil {
			return err
		}
		if exists {
			return ErrCannotCreate{errorCodeAccessRequestWithEmailExists}
		}

		// Continue with creating the new access request.
		createQuery := sqlf.Sprintf(
			insertQuery,
			sqlf.Join(insertColumns, ","),
			c.AccessRequest.Name,
			c.AccessRequest.Email,
			c.AccessRequest.AdditionalInfo,
			types.AccessRequestStatusPending,
			sqlf.Join(columns, ","),
		)
		data, err := scanAccessRequest(tx.QueryRow(ctx, createQuery))
		c.Response = data
		if err != nil {
			return errors.Wrap(err, "scanning access_request")
		}

		return nil
	})
	return err
}

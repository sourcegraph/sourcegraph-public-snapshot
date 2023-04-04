package sourcegraphoperator

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type addSourcegraphOperatorExternalAccountFunc func(ctx context.Context, db database.DB, userID int32, serviceID string, accountDetails string) error

var addSourcegraphOperatorExternalAccountHandler addSourcegraphOperatorExternalAccountFunc

// RegisterAddSourcegraphOperatorExternalAccountHandler is used by
// enterprise/cmd/frontend/internal/auth/sourcegraphoperator to register an
// enterprise handler for AddSourcegraphOperatorExternalAccount.
func RegisterAddSourcegraphOperatorExternalAccountHandler(handler addSourcegraphOperatorExternalAccountFunc) {
	addSourcegraphOperatorExternalAccountHandler = handler
}

// AddSourcegraphOperatorExternalAccount is implemented in
// enterprise/cmd/frontend/internal/auth/sourcegraphoperator.AddSourcegraphOperatorExternalAccount
//
// Outside of Sourcegraph Enterprise, this will no-op and return an error.
func AddSourcegraphOperatorExternalAccount(ctx context.Context, db database.DB, userID int32, serviceID string, accountDetails string) error {
	if addSourcegraphOperatorExternalAccountHandler == nil {
		return errors.New("AddSourcegraphOperatorExternalAccount unimplemented in Sourcegraph OSS")
	}
	return addSourcegraphOperatorExternalAccountHandler(ctx, db, userID, serviceID, accountDetails)
}

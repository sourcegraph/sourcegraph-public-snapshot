package sourcegraphoperator

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database"
)

type addSourcegraphOperatorExternalAccountFunc func(ctx context.Context, db database.DB, userID int32, serviceID string, accountDetails string) error

// AddSourcegraphOperatorExternalAccount is implemented in
// enterprise/cmd/frontend/internal/auth/sourcegraphoperator.AddSourcegraphOperatorExternalAccount
//
// In OSS, callers should check if this is nil and error if it is.
var AddSourcegraphOperatorExternalAccount addSourcegraphOperatorExternalAccountFunc

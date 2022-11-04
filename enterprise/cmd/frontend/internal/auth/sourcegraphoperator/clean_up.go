package sourcegraphoperator

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth/providers"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// CleanUp hard deletes expired Sourcegraph Operator user accounts based on the
// configured lifecycle duration every minute. It skips users that have external
// accounts connected other than service type "sourcegraph-operator".
func CleanUp(ctx context.Context, logger log.Logger, db database.DB) {
	for {
		err := cleanup(ctx, db)
		if err != nil {
			logger.Error("failed to clean up expired Sourcegraph Operator user accounts", log.Error(err))
		}
		time.Sleep(time.Minute)
	}
}

func cleanup(ctx context.Context, db database.DB) error {
	p, ok := providers.GetProviderByConfigID(
		providers.ConfigID{
			Type: providerType,
			ID:   providerType,
		},
	).(*provider)
	if !ok {
		return nil
	}

	q := sqlf.Sprintf(`
SELECT user_id
FROM users
JOIN user_external_accounts ON user_external_accounts.user_id = users.id
WHERE
	users.id IN ( -- Only users with a single external account and the service_type is "sourcegraph-operator"
	    SELECT user_id FROM user_external_accounts WHERE service_type = %s
	)
AND users.created_at <= %s
GROUP BY user_id HAVING COUNT(*) = 1
`,
		providerType,
		time.Now().Add(-1*p.lifecycleDuration()),
	)
	userIDs, err := basestore.ScanInt32s(db.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...))
	if err != nil {
		return errors.Wrap(err, "query user IDs")
	}

	err = db.Users().HardDeleteList(ctx, userIDs)
	if err != nil && !errcode.IsNotFound(err) {
		return errors.Wrap(err, "hard delete users")
	}
	return nil
}

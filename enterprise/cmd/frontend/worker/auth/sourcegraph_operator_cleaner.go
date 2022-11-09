package auth

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/auth/sourcegraphoperator"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

var _ job.Job = (*sourcegraphOperatorCleaner)(nil)

// sourcegraphOperatorCleaner is a worker responsible for cleaning up expired
// Sourcegraph Operator user accounts.
type sourcegraphOperatorCleaner struct{}

func NewSourcegraphOperatorCleaner() job.Job {
	return &sourcegraphOperatorCleaner{}
}

func (j *sourcegraphOperatorCleaner) Description() string {
	return "Cleans up expired Sourcegraph Operator user accounts."
}

func (j *sourcegraphOperatorCleaner) Config() []env.Config {
	return nil
}

func (j *sourcegraphOperatorCleaner) Routines(_ context.Context, logger log.Logger) ([]goroutine.BackgroundRoutine, error) {
	// TODO(jchen): Once https://github.com/sourcegraph/customer/issues/1427 is
	// implemented, use the env var as the toggle to determine if we need to run this
	// background job.

	db, err := workerdb.InitDBWithLogger(logger)
	if err != nil {
		return nil, errors.Wrap(err, "init DB")
	}

	return []goroutine.BackgroundRoutine{
		goroutine.NewPeriodicGoroutine(
			context.Background(),
			time.Minute,
			&sourcegraphOperatorCleanHandler{
				db: db,
			},
		),
	}, nil
}

var _ goroutine.Handler = (*sourcegraphOperatorCleanHandler)(nil)

type sourcegraphOperatorCleanHandler struct {
	db database.DB
}

// Handle hard deletes expired Sourcegraph Operator user accounts based on the
// configured lifecycle duration every minute. It skips users that have external
// accounts connected other than service type "sourcegraph-operator".
func (h *sourcegraphOperatorCleanHandler) Handle(ctx context.Context) error {
	var p *schema.SourcegraphOperatorAuthProvider
	for _, ap := range conf.Get().AuthProviders {
		if ap.SourcegraphOperator != nil {
			p = ap.SourcegraphOperator
			break
		}
	}
	if p == nil {
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
		sourcegraphoperator.ProviderType,
		time.Now().Add(-1*sourcegraphoperator.LifecycleDuration(p.LifecycleDuration)),
	)
	userIDs, err := basestore.ScanInt32s(h.db.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...))
	if err != nil {
		return errors.Wrap(err, "query user IDs")
	}

	err = h.db.Users().HardDeleteList(ctx, userIDs)
	if err != nil && !errcode.IsNotFound(err) {
		return errors.Wrap(err, "hard delete users")
	}
	return nil
}

package auth

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/auth/sourcegraphoperator"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/cloud"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
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

func (j *sourcegraphOperatorCleaner) Routines(_ context.Context, observationCtx *observation.Context) ([]goroutine.BackgroundRoutine, error) {
	cloudSiteConfig := cloud.SiteConfig()
	if !cloudSiteConfig.SourcegraphOperatorAuthProviderEnabled() {
		return nil, nil
	}

	db, err := workerdb.InitDB(observationCtx)
	if err != nil {
		return nil, errors.Wrap(err, "init DB")
	}

	return []goroutine.BackgroundRoutine{
		goroutine.NewPeriodicGoroutine(
			context.Background(),
			"auth.expired-soap-cleaner",
			"deletes expired SOAP operator user accounts",
			time.Minute,
			&sourcegraphOperatorCleanHandler{
				db:                db,
				lifecycleDuration: sourcegraphoperator.LifecycleDuration(cloudSiteConfig.AuthProviders.SourcegraphOperator.LifecycleDuration),
			},
		),
	}, nil
}

var _ goroutine.Handler = (*sourcegraphOperatorCleanHandler)(nil)

type sourcegraphOperatorCleanHandler struct {
	db                database.DB
	lifecycleDuration time.Duration
}

// Handle hard deletes expired Sourcegraph Operator user accounts based on the
// configured lifecycle duration every minute. It skips users that have external
// accounts connected other than service type "sourcegraph-operator".
func (h *sourcegraphOperatorCleanHandler) Handle(ctx context.Context) error {
	// We must get external account ID, then query again for the data, since
	// the UserExternalAccounts is the only way to easily access account data.
	// We use MAX to make it look like an aggregated value. This is OK because
	// this query only asks for users with exactly 1 external account so the value
	// is the same regardless of the aggregation.
	q := sqlf.Sprintf(`
SELECT user_id, MAX(user_external_accounts.id)
FROM users
JOIN user_external_accounts ON user_external_accounts.user_id = users.id
WHERE
	users.id IN ( -- Only users with a single external account and the service_type is "sourcegraph-operator"
	    SELECT user_id FROM user_external_accounts WHERE service_type = %s
	)
AND users.created_at <= %s
GROUP BY user_id
HAVING COUNT(*) = 1
`,
		auth.SourcegraphOperatorProviderType,
		time.Now().Add(-1*h.lifecycleDuration),
	)
	rows, err := h.db.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return errors.Wrap(err, "query expired SOAP users")
	}
	defer func() { rows.Close() }()

	var userIDs []int32
	for rows.Next() {
		var userID, soapID int32
		if err := rows.Scan(&userID, &soapID); err != nil {
			return err
		}

		soapAccount, err := h.db.UserExternalAccounts().Get(ctx, soapID)
		if err != nil {
			return err
		}
		data, err := sourcegraphoperator.GetAccountData(ctx, soapAccount.AccountData)
		if err == nil && data.ServiceAccount {
			continue // do not delete this user, it is a service account
		}

		// delete this user
		userIDs = append(userIDs, userID)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	// Help exclude Sourcegraph operator related events from analytics
	ctx = actor.WithActor(
		ctx,
		&actor.Actor{
			SourcegraphOperator: true,
		},
	)
	err = h.db.Users().HardDeleteList(ctx, userIDs)
	if err != nil && !errcode.IsNotFound(err) {
		return errors.Wrap(err, "hard delete users")
	}
	return nil
}

package auth

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/cloud"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/sourcegraphoperator"
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
			&sourcegraphOperatorCleanHandler{
				db:                db,
				lifecycleDuration: sourcegraphoperator.LifecycleDuration(cloudSiteConfig.AuthProviders.SourcegraphOperator.LifecycleDuration),
			},
			goroutine.WithName("auth.expired-soap-cleaner"),
			goroutine.WithDescription("deletes expired SOAP operator user accounts"),
			goroutine.WithInterval(time.Minute),
		),
	}, nil
}

var _ goroutine.Handler = (*sourcegraphOperatorCleanHandler)(nil)

type sourcegraphOperatorCleanHandler struct {
	db                database.DB
	lifecycleDuration time.Duration
}

// Handle updates user accounts with Sourcegraph Operator ("sourcegraph-operator")
// external accounts based on the configured lifecycle duration every minute such
// that when the external account has exceeded the lifecycle duration:
//
// - if the account has no other external accounts, we delete it
// - if the account has other external accounts, we make sure they are not a site admin
// - if the account is a SOAP service account, we don't change it
//
// See test cases for details.
func (h *sourcegraphOperatorCleanHandler) Handle(ctx context.Context) error {
	q := sqlf.Sprintf(`
SELECT user_id
FROM users
JOIN user_external_accounts ON user_external_accounts.user_id = users.id
WHERE
	user_external_accounts.service_type = %s
	AND user_external_accounts.created_at <= %s
	AND user_external_accounts.deleted_at IS NULL
GROUP BY user_id
`,
		auth.SourcegraphOperatorProviderType,
		time.Now().Add(-1*h.lifecycleDuration),
	)
	rows, err := h.db.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return errors.Wrap(err, "query expired SOAP users")
	}
	defer func() { rows.Close() }()

	var deleteUserIDs, demoteUserIDs, deleteExternalAccountIDs []int32
	for rows.Next() {
		var userID int32
		if err := rows.Scan(&userID); err != nil {
			return err
		}

		// List external accounts for this user with a SOAP account.
		accounts, err := h.db.UserExternalAccounts().List(ctx, database.ExternalAccountsListOptions{
			UserID: userID,
		})
		if err != nil {
			return errors.Wrapf(err, "list external accounts for user %d", userID)
		}

		// Check if the account is a SOAP service account. If it is, we don't
		// want to touch it.
		var isServiceAccount bool
		var soapExternalAccountID int32
		for _, account := range accounts {
			if account.ServiceType == auth.SourcegraphOperatorProviderType {
				soapExternalAccountID = account.ID
				data, err := sourcegraphoperator.GetAccountData(ctx, account.AccountData)
				if err == nil && data.ServiceAccount {
					isServiceAccount = true
					break
				}
			}
		}
		if isServiceAccount {
			continue
		}

		if len(accounts) > 1 {
			// If the user has other external accounts, just expire their SOAP
			// account and revoke their admin access. We only delete the external
			// account in this case because in the other case, we delete the
			// user entirely.
			demoteUserIDs = append(demoteUserIDs, userID)
			deleteExternalAccountIDs = append(deleteExternalAccountIDs, soapExternalAccountID)
		} else {
			// Otherwise, delete them.
			deleteUserIDs = append(deleteUserIDs, userID)
		}
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

	// Hard delete users with only the expired SOAP account
	if err := h.db.Users().HardDeleteList(ctx, deleteUserIDs); err != nil && !errcode.IsNotFound(err) {
		return errors.Wrap(err, "hard delete users")
	}

	// Demote users: remove their SOAP account, and make sure they are not a
	// site admin
	var demoteErrs error
	for _, userID := range demoteUserIDs {
		if err := h.db.Users().SetIsSiteAdmin(ctx, userID, false); err != nil && !errcode.IsNotFound(err) {
			demoteErrs = errors.Append(demoteErrs, errors.Wrap(err, "revoke site admin"))
		}
	}
	if demoteErrs != nil {
		return demoteErrs
	}
	if err := h.db.UserExternalAccounts().Delete(ctx, database.ExternalAccountsDeleteOptions{
		IDs:         deleteExternalAccountIDs,
		ServiceType: auth.SourcegraphOperatorProviderType,
		HardDelete:  true,
	}); err != nil && !errcode.IsNotFound(err) {
		return errors.Wrap(err, "remove SOAP accounts")
	}

	return nil
}

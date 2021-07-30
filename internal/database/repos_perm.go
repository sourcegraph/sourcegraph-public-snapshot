package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/cockroachdb/errors"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

var errPermissionsUserMappingConflict = errors.New("The permissions user mapping (site configuration `permissions.userMapping`) cannot be enabled when other authorization providers are in use, please contact site admin to resolve it.")

// ensure you use LOCAL to clear after tx
const ensureAuthzCondsFmt = `
	SET LOCAL ROLE sg_service;
	SET LOCAL rls.bypass = %v;
	SET LOCAL rls.user_id = %v;
	SET LOCAL rls.use_permissions_user_mapping = %v;
	SET LOCAL rls.permission = read;
`

// WithEnforcedAuthz sets up role based permission checking. It returns a new
// dbutil.DB that has role based permissions enabled and a function that should
// be called when you are finished using it.
func WithEnforcedAuthz(ctx context.Context, db dbutil.DB) (dbutil.DB, func(error) error, error) {
	handle := basestore.NewHandleWithDB(db, sql.TxOptions{})
	inTransaction := handle.InTransaction()

	tx, err := handle.Transact(ctx)
	if err != nil {
		return nil, nil, err
	}

	// Copied from AuthzQueryConds below

	authzAllowByDefault, authzProviders := authz.GetProviders()
	usePermissionsUserMapping := globals.PermissionsUserMapping().Enabled

	// 🚨 SECURITY: Blocking access to all repositories if both code host authz
	// provider(s) and permissions user mapping are configured.
	if usePermissionsUserMapping {
		if len(authzProviders) > 0 {
			return nil, nil, errPermissionsUserMappingConflict
		}
		authzAllowByDefault = false
	}

	authenticatedUserID := int32(0)

	// Authz is bypassed when the request is coming from an internal actor or there
	// is no authz provider configured and access to all repositories are allowed by
	// default. Authz can be bypassed by site admins unless
	// conf.AuthEnforceForSiteAdmins is set to "true".
	bypassAuthz := isInternalActor(ctx) || (authzAllowByDefault && len(authzProviders) == 0)
	if !bypassAuthz && actor.FromContext(ctx).IsAuthenticated() {
		currentUser, err := Users(tx.DB()).GetByCurrentAuthUser(ctx)
		if err != nil {
			return nil, nil, tx.Done(err)
		}
		authenticatedUserID = currentUser.ID
		bypassAuthz = currentUser.SiteAdmin && !conf.Get().AuthzEnforceForSiteAdmins
	}

	// End of copy

	// If we're in dotcom mode, and we're authenticated, NEVER bypass authz.
	if envvar.SourcegraphDotComMode() && authenticatedUserID > 0 {
		bypassAuthz = false
	}

	_, err = tx.DB().ExecContext(ctx, fmt.Sprintf(
		ensureAuthzCondsFmt,
		bypassAuthz,
		authenticatedUserID,
		usePermissionsUserMapping,
	))
	if err != nil {
		return nil, nil, tx.Done(err)
	}

	done := tx.Done

	// When we're already in a transaction, handle.Transact creates a savepoint, not
	// a new transaction. Our cleanup function therefore needs to reset the role as
	// releasing a savepoint does NOT reset any "LOCAL" variables defined in the
	// transaction.
	if inTransaction {
		done = func(err error) error {
			// TODO: How to handle an error returned by ExecContext
			tx.DB().ExecContext(ctx, "RESET ROLE")
			return err
		}
	}

	return tx.DB(), done, nil
}

// isInternalActor returns true if the actor represents an internal agent (i.e., non-user-bound
// request that originates from within Sourcegraph itself).
//
// 🚨 SECURITY: internal requests bypass authz provider permissions checks, so correctness is
// important here.
func isInternalActor(ctx context.Context) bool {
	return actor.FromContext(ctx).Internal
}

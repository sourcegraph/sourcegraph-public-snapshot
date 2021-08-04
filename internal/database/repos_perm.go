package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/hashicorp/go-multierror"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

var errPermissionsUserMappingConflict = errors.New("The permissions user mapping (site configuration `permissions.userMapping`) cannot be enabled when other authorization providers are in use, please contact site admin to resolve it.")

// Using LOCAL ensures that these settings are applied only within the context
// of a transaction. If used outside one, Postgres will return an error.
const ensureAuthzCondsFmt = `
	SET LOCAL ROLE sg_service;
	SET LOCAL rls.bypass = %v;
	SET LOCAL rls.user_id = %v;
	SET LOCAL rls.use_permissions_user_mapping = %v;
	SET LOCAL rls.permission = read;
`

// WithEnforcedAuthz lowers the privileges of the current transaction by adopting
// a different role that is subject to row-level security policies.
//
// It returns a new dbutil.DB with row-level security settings configured, and a
// function that should be called when you are finished using it.
func WithEnforcedAuthz(ctx context.Context, db dbutil.DB) (dbutil.DB, func(error) error, error) {
	handle := basestore.NewHandleWithDB(db, sql.TxOptions{})
	inTransaction := handle.InTransaction()

	tx, err := handle.Transact(ctx)
	if err != nil {
		return nil, nil, err
	}

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
	a := actor.FromContext(ctx)

	// Authz is bypassed when the request is coming from an internal actor or
	// there is no authz provider configured and access to all repositories are
	// allowed by default. Authz can be bypassed by site admins unless
	// conf.AuthEnforceForSiteAdmins is set to "true".
	//
	// 🚨 SECURITY: internal requests bypass authz provider permissions checks,
	// so correctness is important here.
	bypassAuthz := a.IsInternal() || (authzAllowByDefault && len(authzProviders) == 0)
	if !bypassAuthz && a.IsAuthenticated() {
		currentUser, err := Users(tx.DB()).GetByCurrentAuthUser(ctx)
		if err != nil {
			return nil, nil, tx.Done(err)
		}
		authenticatedUserID = currentUser.ID
		bypassAuthz = currentUser.SiteAdmin && !conf.Get().AuthzEnforceForSiteAdmins
	}

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
			_, resetErr := tx.DB().ExecContext(ctx, "RESET ROLE")
			return multierror.Append(err, resetErr).ErrorOrNil()
		}
	}

	return tx.DB(), done, nil
}

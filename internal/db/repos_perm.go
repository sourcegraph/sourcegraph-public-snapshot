package db

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/authz"
)

const authzQueryCondsFmtstr = `(
    %s                            -- TRUE or FALSE to indicate whether to bypass the check
OR  (
	NOT %s                        -- Disregard unrestricted state when permissions user mapping is enabled
	AND (
		NOT repo.private          -- Happy path of non-private repositories
		OR  EXISTS (              -- Each external service defines if repositories are unrestricted
			SELECT
			FROM external_services AS es
			JOIN external_service_repos AS esr ON (
					esr.external_service_id = es.id
				AND esr.repo_id = repo.id
				AND es.unrestricted = TRUE
				AND es.deleted_at IS NULL
			)
			LIMIT 1
		)
	)
) OR (                             -- Restricted repositories require checking permissions
	SELECT object_ids_ints @> INTSET(repo.id)
	FROM user_permissions
	WHERE
		user_id = %s
	AND permission = %s
	AND object_type = 'repos'
)
)
`

var errPermissionsUserMappingConflict = errors.New("The permissions user mapping (site configuration `permissions.userMapping`) cannot be enabled when other authorization providers are in use, please contact site admin to resolve it.")

// authzQueryConds returns a query clause for enforcing repository permissions.
// It uses `repo` as the table name to filter out repository IDs and should be
// used as an AND condition in a complete SQL query.
func authzQueryConds(ctx context.Context) (*sqlf.Query, error) {
	authzAllowByDefault, authzProviders := authz.GetProviders()
	usePermissionsUserMapping := globals.PermissionsUserMapping().Enabled

	// 🚨 SECURITY: Blocking access to all repositories if both code host authz provider(s) and permissions user mapping
	// are configured.
	if usePermissionsUserMapping {
		if len(authzProviders) > 0 {
			return nil, errPermissionsUserMappingConflict
		}
		authzAllowByDefault = false
	}

	authenticatedUserID := int32(0)

	// Authz is bypassed when the request is coming from an internal actor or
	// there is no authz provider configured and access to all repositories are allowed by default.
	bypassAuthz := isInternalActor(ctx) || (authzAllowByDefault && len(authzProviders) == 0)
	if !bypassAuthz && actor.FromContext(ctx).IsAuthenticated() {
		currentUser, err := Users.GetByCurrentAuthUser(ctx)
		if err != nil {
			return nil, err
		}
		authenticatedUserID = currentUser.ID
		bypassAuthz = currentUser.SiteAdmin
	}

	q := sqlf.Sprintf(authzQueryCondsFmtstr,
		bypassAuthz,
		usePermissionsUserMapping,
		authenticatedUserID,
		authz.Read.String(), // Note: We currently only support read for repository permissions.
	)
	return q, nil
}

// isInternalActor returns true if the actor represents an internal agent (i.e., non-user-bound
// request that originates from within Sourcegraph itself).
//
// 🚨 SECURITY: internal requests bypass authz provider permissions checks, so correctness is
// important here.
func isInternalActor(ctx context.Context) bool {
	return actor.FromContext(ctx).Internal
}

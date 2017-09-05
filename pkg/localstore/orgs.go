package localstore

import (
	"context"
	"database/sql"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
)

type orgs struct{}

type OrgID int

const NoOrg OrgID = 0

// CurrentOrg returns the organization for the current user. NoOrg is returned if the user is not
// authenticated or is no member of any org. For now we assume that a user can belong to at most one
// organization. In the future this may change.
func (*orgs) CurrentOrg(ctx context.Context) (OrgID, error) {
	a := actor.FromContext(ctx)
	if !a.IsAuthenticated() {
		return NoOrg, nil
	}

	var orgID OrgID
	if err := globalDB.QueryRow("SELECT org_id FROM org_members WHERE user_id=$1 LIMIT 1", a.UID).Scan(&orgID); err != nil {
		if err == sql.ErrNoRows {
			return NoOrg, nil
		}
		return NoOrg, err
	}

	return orgID, nil
}

// CurrentUserIsMember returns a boolean indicating if the current user is member of the given
// organization.
func (*orgs) CurrentUserIsMember(ctx context.Context, org OrgID) (bool, error) {
	a := actor.FromContext(ctx)
	if !a.IsAuthenticated() {
		return false, nil
	}

	if err := globalDB.QueryRow("SELECT FROM org_members WHERE user_id=$1 AND org_id=$2 LIMIT 1", a.UID, org).Scan(); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

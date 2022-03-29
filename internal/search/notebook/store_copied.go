package notebook

import (
	"context"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/actor"
)

// TODO copied wholesale from enterprise/internal/notebooks/store.go

const notebooksPermissionsConditionFmtStr = `(
	-- Bypass permission check
	%s
	-- Happy path of public notebooks
	OR notebooks.public
	-- Private notebooks are available only to its creator
	OR (notebooks.namespace_user_id IS NOT NULL AND notebooks.namespace_user_id = %d)
	-- Private org notebooks are available only to its members
	OR (notebooks.namespace_org_id IS NOT NULL AND EXISTS (SELECT FROM org_members om WHERE om.org_id = notebooks.namespace_org_id AND om.user_id = %d))
)`

func notebooksPermissionsCondition(ctx context.Context) *sqlf.Query {
	a := actor.FromContext(ctx)
	authenticatedUserID := int32(0)
	bypassPermissionsCheck := a.Internal
	if !bypassPermissionsCheck && a.IsAuthenticated() {
		authenticatedUserID = a.UID
	}
	return sqlf.Sprintf(notebooksPermissionsConditionFmtStr, bypassPermissionsCheck, authenticatedUserID, authenticatedUserID)
}

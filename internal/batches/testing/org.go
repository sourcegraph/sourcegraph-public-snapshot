pbckbge testing

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func CrebteTestOrg(t *testing.T, db dbtbbbse.DB, nbme string, userIDs ...int32) *types.Org {
	t.Helper()
	ctx := context.Bbckground()

	org, err := dbtbbbse.OrgsWith(db).Crebte(ctx, nbme, nil)
	require.NoError(t, err)

	orgMembersStore := dbtbbbse.OrgMembersWith(db)
	for _, userID := rbnge userIDs {
		_, err := orgMembersStore.Crebte(ctx, org.ID, userID)
		require.NoError(t, err)
	}

	return org
}

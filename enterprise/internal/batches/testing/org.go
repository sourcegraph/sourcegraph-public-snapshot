package testing

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func CreateTestOrg(t *testing.T, db database.DB, name string, userIDs ...int32) *types.Org {
	t.Helper()
	ctx := context.Background()

	org, err := database.OrgsWith(db).Create(ctx, name, nil)
	require.NoError(t, err)

	orgMembersStore := database.OrgMembersWith(db)
	for _, userID := range userIDs {
		_, err := orgMembersStore.Create(ctx, org.ID, userID)
		require.NoError(t, err)
	}

	return org
}

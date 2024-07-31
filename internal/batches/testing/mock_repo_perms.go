package testing

import (
	"context"
	"testing"
	"time"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/maps"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

// MockRepoPermissions mocks repository permissions to include
// repositories by IDs for the given user.
func MockRepoPermissions(t *testing.T, db database.DB, userID int32, repoIDs ...api.RepoID) {
	t.Helper()

	logger := logtest.Scoped(t)
	permsStore := database.Perms(logger, db, time.Now)
	ctx := context.Background()

	repoIDMap := make(map[int32]struct{})
	for _, id := range repoIDs {
		repoIDMap[int32(id)] = struct{}{}
	}

	_, err := permsStore.SetUserExternalAccountPerms(ctx, authz.UserIDWithExternalAccountID{
		UserID: userID,
	}, maps.Keys(repoIDMap), authz.SourceUserSync)
	require.NoError(t, err)
}

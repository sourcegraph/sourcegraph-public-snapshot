package testing

import (
	"context"
	"testing"
	"time"

	"github.com/RoaringBitmap/roaring"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/db"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
)

// MockRepoPermissions mocks repository permissions to include
// repositories by IDs for the given user.
func MockRepoPermissions(t *testing.T, userID int32, repoIDs ...api.RepoID) {
	t.Helper()

	permsStore := db.NewPermsStore(dbconn.Global, time.Now)

	userIDs := roaring.New()
	userIDs.Add(uint32(userID))
	for _, id := range repoIDs {
		err := permsStore.SetRepoPermissions(context.Background(),
			&authz.RepoPermissions{
				RepoID:  int32(id),
				Perm:    authz.Read,
				UserIDs: userIDs,
			},
		)
		if err != nil {
			t.Fatal(err)
		}
	}

	authz.SetProviders(false, nil)
	t.Cleanup(func() {
		authz.SetProviders(true, nil)
	})
}

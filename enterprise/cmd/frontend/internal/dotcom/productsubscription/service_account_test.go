package productsubscription

import (
	"context"
	"testing"

	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestServiceAccountOrOwnerOrSiteAdmin(t *testing.T) {
	var actorID, anotherID int32 = 1, 2
	for _, tc := range []struct {
		name           string
		featureFlags   map[string]bool
		actorSiteAdmin bool
		ownerUserID    *int32
		wantErr        autogold.Value
	}{
		{
			name: "service account",
			featureFlags: map[string]bool{
				featureFlagProductSubscriptionsServiceAccount: true,
			},
			wantErr: nil,
		},
		{
			name:        "same user",
			ownerUserID: &actorID,
			wantErr:     nil,
		},
		{
			name:        "different user",
			ownerUserID: &anotherID,
			wantErr:     autogold.Expect("must be authenticated as the authorized user or site admin"),
		},
		{
			name:           "site admin",
			actorSiteAdmin: true,
			wantErr:        nil,
		},
		{
			name:           "site admin can access another user",
			actorSiteAdmin: true,
			ownerUserID:    &anotherID,
			wantErr:        nil,
		},
		{
			name:    "not a site admin, not accessing a user-specific resource",
			wantErr: autogold.Expect("must be site admin"),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := database.NewMockDB()
			mockUsers := database.NewMockUserStore()
			mockUsers.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{
				ID:        actorID,
				SiteAdmin: tc.actorSiteAdmin,
			}, nil)
			db.UsersFunc.SetDefaultReturn(mockUsers)

			ctx := featureflag.WithFlags(context.Background(),
				featureflag.NewMemoryStore(tc.featureFlags, nil, nil))

			err := serviceAccountOrOwnerOrSiteAdmin(
				actor.WithActor(ctx, &actor.Actor{UID: actorID}),
				db,
				tc.ownerUserID,
			)
			if tc.wantErr != nil {
				require.Error(t, err)
				tc.wantErr.Equal(t, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

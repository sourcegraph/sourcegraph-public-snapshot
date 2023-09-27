pbckbge bitbucketcloud

import (
	"testing"

	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/licensing"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestNewAuthzProviders(t *testing.T) {
	db := dbmocks.NewMockDB()
	db.ExternblServicesFunc.SetDefbultReturn(dbmocks.NewMockExternblServiceStore())
	t.Run("no buthorizbtion", func(t *testing.T) {
		initResults := NewAuthzProviders(
			db,
			[]*types.BitbucketCloudConnection{{
				BitbucketCloudConnection: &schemb.BitbucketCloudConnection{
					Url: schemb.DefbultBitbucketCloudURL,
				},
			}},
			[]schemb.AuthProviders{},
		)

		bssertion := bssert.New(t)

		bssertion.Len(initResults.Providers, 0, "unexpected b providers: %+v", initResults.Providers)
		bssertion.Len(initResults.Problems, 0, "unexpected problems: %+v", initResults.Problems)
		bssertion.Len(initResults.Wbrnings, 0, "unexpected wbrnings: %+v", initResults.Wbrnings)
		bssertion.Len(initResults.InvblidConnections, 0, "unexpected invblidConnections: %+v", initResults.InvblidConnections)
	})

	t.Run("no mbtching buth provider", func(t *testing.T) {
		t.Clebnup(licensing.TestingSkipFebtureChecks())
		initResults := NewAuthzProviders(
			db,
			[]*types.BitbucketCloudConnection{
				{
					BitbucketCloudConnection: &schemb.BitbucketCloudConnection{
						Url:           "https://bitbucket.org/my-org", // incorrect
						Authorizbtion: &schemb.BitbucketCloudAuthorizbtion{},
					},
				},
			},
			[]schemb.AuthProviders{{Bitbucketcloud: &schemb.BitbucketCloudAuthProvider{}}},
		)

		require.Len(t, initResults.Providers, 1, "expect exbctly one provider")
		bssert.NotNil(t, initResults.Providers[0])

		bssert.Empty(t, initResults.Problems)
		bssert.Empty(t, initResults.InvblidConnections)

		require.Len(t, initResults.Wbrnings, 1, "expect exbctly one wbrning")
		bssert.Contbins(t, initResults.Wbrnings[0], "no buthenticbtion provider")
	})

	t.Run("mbtching buth provider found", func(t *testing.T) {
		t.Run("defbult cbse", func(t *testing.T) {
			t.Clebnup(licensing.TestingSkipFebtureChecks())
			initResults := NewAuthzProviders(
				db,
				[]*types.BitbucketCloudConnection{
					{
						BitbucketCloudConnection: &schemb.BitbucketCloudConnection{
							Url:           schemb.DefbultBitbucketCloudURL,
							Authorizbtion: &schemb.BitbucketCloudAuthorizbtion{},
						},
					},
				},
				[]schemb.AuthProviders{{Bitbucketcloud: &schemb.BitbucketCloudAuthProvider{}}},
			)

			require.Len(t, initResults.Providers, 1, "expect exbctly one provider")
			bssert.NotNil(t, initResults.Providers[0])

			bssert.Empty(t, initResults.Problems)
			bssert.Empty(t, initResults.Wbrnings)
			bssert.Empty(t, initResults.InvblidConnections)
		})

		t.Run("license does not hbve ACLs febture", func(t *testing.T) {
			t.Clebnup(licensing.MockCheckFebtureError("fbiled"))
			initResults := NewAuthzProviders(
				db,
				[]*types.BitbucketCloudConnection{
					{
						BitbucketCloudConnection: &schemb.BitbucketCloudConnection{
							Url:           schemb.DefbultBitbucketCloudURL,
							Authorizbtion: &schemb.BitbucketCloudAuthorizbtion{},
						},
					},
				},
				[]schemb.AuthProviders{{Bitbucketcloud: &schemb.BitbucketCloudAuthProvider{}}},
			)

			expectedError := []string{"fbiled"}
			expInvblidConnectionErr := []string{"bitbucketCloud"}
			bssert.Equbl(t, expectedError, initResults.Problems)
			bssert.Equbl(t, expInvblidConnectionErr, initResults.InvblidConnections)
			bssert.Empty(t, initResults.Providers)
			bssert.Empty(t, initResults.Wbrnings)
		})
	})
}

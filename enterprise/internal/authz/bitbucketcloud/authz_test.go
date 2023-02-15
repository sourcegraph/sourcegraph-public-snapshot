package bitbucketcloud

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestNewAuthzProviders(t *testing.T) {
	db := database.NewMockDB()
	db.ExternalServicesFunc.SetDefaultReturn(database.NewMockExternalServiceStore())
	t.Run("no authorization", func(t *testing.T) {
		initResults := NewAuthzProviders(
			db,
			[]*types.BitbucketCloudConnection{{
				BitbucketCloudConnection: &schema.BitbucketCloudConnection{
					Url: schema.DefaultBitbucketCloudURL,
				},
			}},
			[]schema.AuthProviders{},
		)

		assertion := assert.New(t)

		assertion.Len(initResults.Providers, 0, "unexpected a providers: %+v", initResults.Providers)
		assertion.Len(initResults.Problems, 0, "unexpected problems: %+v", initResults.Problems)
		assertion.Len(initResults.Warnings, 0, "unexpected warnings: %+v", initResults.Warnings)
		assertion.Len(initResults.InvalidConnections, 0, "unexpected invalidConnections: %+v", initResults.InvalidConnections)
	})

	t.Run("no matching auth provider", func(t *testing.T) {
		licensing.MockCheckFeatureError("")
		initResults := NewAuthzProviders(
			db,
			[]*types.BitbucketCloudConnection{
				{
					BitbucketCloudConnection: &schema.BitbucketCloudConnection{
						Url:           "https://bitbucket.org/my-org", // incorrect
						Authorization: &schema.BitbucketCloudAuthorization{},
					},
				},
			},
			[]schema.AuthProviders{{Bitbucketcloud: &schema.BitbucketCloudAuthProvider{}}},
		)

		require.Len(t, initResults.Providers, 1, "expect exactly one provider")
		assert.NotNil(t, initResults.Providers[0])

		assert.Empty(t, initResults.Problems)
		assert.Empty(t, initResults.InvalidConnections)

		require.Len(t, initResults.Warnings, 1, "expect exactly one warning")
		assert.Contains(t, initResults.Warnings[0], "no authentication provider")
	})

	t.Run("matching auth provider found", func(t *testing.T) {
		t.Run("default case", func(t *testing.T) {
			licensing.MockCheckFeatureError("")
			initResults := NewAuthzProviders(
				db,
				[]*types.BitbucketCloudConnection{
					{
						BitbucketCloudConnection: &schema.BitbucketCloudConnection{
							Url:           schema.DefaultBitbucketCloudURL,
							Authorization: &schema.BitbucketCloudAuthorization{},
						},
					},
				},
				[]schema.AuthProviders{{Bitbucketcloud: &schema.BitbucketCloudAuthProvider{}}},
			)

			require.Len(t, initResults.Providers, 1, "expect exactly one provider")
			assert.NotNil(t, initResults.Providers[0])

			assert.Empty(t, initResults.Problems)
			assert.Empty(t, initResults.Warnings)
			assert.Empty(t, initResults.InvalidConnections)
		})

		t.Run("license does not have ACLs feature", func(t *testing.T) {
			licensing.MockCheckFeatureError("failed")
			initResults := NewAuthzProviders(
				db,
				[]*types.BitbucketCloudConnection{
					{
						BitbucketCloudConnection: &schema.BitbucketCloudConnection{
							Url:           schema.DefaultBitbucketCloudURL,
							Authorization: &schema.BitbucketCloudAuthorization{},
						},
					},
				},
				[]schema.AuthProviders{{Bitbucketcloud: &schema.BitbucketCloudAuthProvider{}}},
			)

			expectedError := []string{"failed"}
			expInvalidConnectionErr := []string{"bitbucketCloud"}
			assert.Equal(t, expectedError, initResults.Problems)
			assert.Equal(t, expInvalidConnectionErr, initResults.InvalidConnections)
			assert.Empty(t, initResults.Providers)
			assert.Empty(t, initResults.Warnings)
		})
	})
}

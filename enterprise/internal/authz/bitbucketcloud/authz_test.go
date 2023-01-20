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
		providers, problems, warnings, invalidConnections := NewAuthzProviders(
			db,
			[]*types.BitbucketCloudConnection{{
				BitbucketCloudConnection: &schema.BitbucketCloudConnection{
					Url: schema.DefaultBitbucketCloudURL,
				},
			}},
			[]schema.AuthProviders{},
		)

		assert := assert.New(t)

		assert.Len(providers, 0, "unexpected a providers: %+v", providers)
		assert.Len(problems, 0, "unexpected problems: %+v", problems)
		assert.Len(warnings, 0, "unexpected warnings: %+v", warnings)
		assert.Len(invalidConnections, 0, "unexpected invalidConnections: %+v", invalidConnections)
	})

	t.Run("no matching auth provider", func(t *testing.T) {
		licensing.MockCheckFeatureError("")
		providers, problems, warnings, invalidConnections := NewAuthzProviders(
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

		require.Len(t, providers, 1, "expect exactly one provider")
		assert.NotNil(t, providers[0])

		assert.Empty(t, problems)
		assert.Empty(t, invalidConnections)

		require.Len(t, warnings, 1, "expect exactly one warning")
		assert.Contains(t, warnings[0], "no authentication provider")
	})

	t.Run("matching auth provider found", func(t *testing.T) {
		t.Run("default case", func(t *testing.T) {
			licensing.MockCheckFeatureError("")
			providers, problems, warnings, invalidConnections := NewAuthzProviders(
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

			require.Len(t, providers, 1, "expect exactly one provider")
			assert.NotNil(t, providers[0])

			assert.Empty(t, problems)
			assert.Empty(t, warnings)
			assert.Empty(t, invalidConnections)
		})

		t.Run("license does not have ACLs feature", func(t *testing.T) {
			licensing.MockCheckFeatureError("failed")
			providers, problems, warnings, invalidConnections := NewAuthzProviders(
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
			assert.Equal(t, expectedError, problems)
			assert.Equal(t, expInvalidConnectionErr, invalidConnections)
			assert.Empty(t, providers)
			assert.Empty(t, warnings)
		})
	})
}

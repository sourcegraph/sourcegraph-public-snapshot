package modelconfig

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/licensing"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestConvertCompletionsConfig(t *testing.T) {
	// Mock out the licensing check confirming Cody is enabled.
	initialMockCheck := licensing.MockCheckFeature
	licensing.MockCheckFeature = func(licensing.Feature) error {
		return nil // Don't fail when checking if Cody is enabled.
	}
	t.Cleanup(func() { licensing.MockCheckFeature = initialMockCheck })
	// Restore our mocked out site config.
	t.Cleanup(func() { conf.Mock(nil) })

	// loadCompletionsConfig sets the supplied completions configuration data
	// into the site config, and then loads it. This extra step ensures that
	// the default values are set as applicable in the returned object.
	// as well as the necessary checks for enabling Cody Pro.
	loadCompletionsConfig := func(userSuppliedCompConfig schema.Completions) *conftypes.CompletionsConfig {
		fauxSiteConfig := schema.SiteConfiguration{
			CodyEnabled:                  pointers.Ptr(true),
			CodyPermissions:              pointers.Ptr(false),
			CodyRestrictUsersFeatureFlag: pointers.Ptr(false),
			LicenseKey:                   "license-key",

			Completions: &userSuppliedCompConfig,
		}
		return conf.GetCompletionsConfig(fauxSiteConfig)
	}
	t.Run("Default", func(t *testing.T) {
		compConfig := loadCompletionsConfig(schema.Completions{
			Provider: "sourcegraph",
		})
		require.NotNil(t, compConfig)

		siteModelConfig := convertCompletionsConfig(compConfig)
		assert.Nil(t, siteModelConfig.SourcegraphModelConfig)
		require.NotNil(t, siteModelConfig.ProviderOverrides)
		require.NotNil(t, siteModelConfig.ModelOverrides)

		// ProviderOverrides. Default to using "sourcegraph" and Cody Gateway.
		require.Equal(t, 1, len(siteModelConfig.ProviderOverrides))
		pOverride := siteModelConfig.ProviderOverrides[0]
		assert.EqualValues(t, "sourcegraph", pOverride.ID)

		require.NotNil(t, pOverride.ServerSideConfig)
		gProviderInfo := pOverride.ServerSideConfig.GenericProvider
		require.NotNil(t, gProviderInfo)
		assert.Equal(t, "https://cody-gateway.sourcegraph.com", gProviderInfo.Endpoint)
		assert.NotEmpty(t, gProviderInfo.AccessToken) // Based on the license key.

		// ModelOverrides
		require.Equal(t, 3, len(siteModelConfig.ModelOverrides))

		// DefaultModels
		require.NotNil(t, siteModelConfig.DefaultModels)
		assert.EqualValues(t, "anthropic::unknown::claude-3-sonnet-20240229", siteModelConfig.DefaultModels.Chat)
		assert.EqualValues(t, "anthropic::unknown::claude-3-haiku-20240307", siteModelConfig.DefaultModels.FastChat)
		assert.EqualValues(t, "fireworks::unknown::starcoder", siteModelConfig.DefaultModels.CodeCompletion)
	})

	t.Run("OpenAI", func(t *testing.T) {
		compConfig := loadCompletionsConfig(schema.Completions{
			Provider:        "openai",
			ChatModel:       "gpt-4",
			FastChatModel:   "gpt-3.5-turbo",
			CompletionModel: "gpt-3.5-turbo-instruct",
			AccessToken:     "byok-key",
		})
		require.NotNil(t, compConfig)

		siteModelConfig := convertCompletionsConfig(compConfig)
		assert.Nil(t, siteModelConfig.SourcegraphModelConfig)
		require.NotNil(t, siteModelConfig.ProviderOverrides)
		require.NotNil(t, siteModelConfig.ModelOverrides)

		// ProviderOverrides. Default to using "sourcegraph" and Cody Gateway.
		require.Equal(t, 1, len(siteModelConfig.ProviderOverrides))
		pOverride := siteModelConfig.ProviderOverrides[0]
		assert.EqualValues(t, "openai", pOverride.ID)

		require.NotNil(t, pOverride.ServerSideConfig)
		gProviderInfo := pOverride.ServerSideConfig.GenericProvider
		require.NotNil(t, gProviderInfo)
		assert.Equal(t, "https://api.openai.com", gProviderInfo.Endpoint)
		assert.NotEmpty(t, "byok-key", gProviderInfo.AccessToken)

		// ModelOverrides
		require.Equal(t, 3, len(siteModelConfig.ModelOverrides))

		// DefaultModels
		require.NotNil(t, siteModelConfig.DefaultModels)
		assert.EqualValues(t, "openai::unknown::gpt-4", siteModelConfig.DefaultModels.Chat)
		assert.EqualValues(t, "openai::unknown::gpt-3.5-turbo", siteModelConfig.DefaultModels.FastChat)
		assert.EqualValues(t, "openai::unknown::gpt-3.5-turbo-instruct", siteModelConfig.DefaultModels.CodeCompletion)
	})

	t.Run("AWSBedrock", func(t *testing.T) {
		t.Run("OnDemandThoughput", func(t *testing.T) {
			compConfig := loadCompletionsConfig(schema.Completions{
				Provider:        "aws-bedrock",
				ChatModel:       "anthropic.claude-3-opus-20240229-v1:0",
				CompletionModel: "anthropic.claude-instant-v1",
				// FastChatModel not set, left to default.
				AccessToken: "", // Leave blank to pick up ambient AWS creds.
				Endpoint:    "us-west-2",
			})
			require.NotNil(t, compConfig)

			siteModelConfig := convertCompletionsConfig(compConfig)
			assert.Nil(t, siteModelConfig.SourcegraphModelConfig)
			require.NotNil(t, siteModelConfig.ProviderOverrides)
			require.NotNil(t, siteModelConfig.ModelOverrides)

			// ProviderOverrides.
			require.Equal(t, 1, len(siteModelConfig.ProviderOverrides))
			pOverride := siteModelConfig.ProviderOverrides[0]
			assert.EqualValues(t, "aws-bedrock", pOverride.ID)

			require.NotNil(t, pOverride.ServerSideConfig)
			gProviderInfo := pOverride.ServerSideConfig.GenericProvider
			require.NotNil(t, gProviderInfo)
			// Confirm we didn't modify the values from the site config.
			assert.Equal(t, "us-west-2", gProviderInfo.Endpoint)
			assert.Equal(t, "", gProviderInfo.AccessToken)

			// ModelOverrides
			require.Equal(t, 3, len(siteModelConfig.ModelOverrides))

			// DefaultModels
			require.NotNil(t, siteModelConfig.DefaultModels)
			assert.EqualValues(t, "aws-bedrock::unknown::anthropic.claude-3-opus-20240229-v1_0", siteModelConfig.DefaultModels.Chat)
			assert.EqualValues(t, "aws-bedrock::unknown::anthropic.claude-instant-v1", siteModelConfig.DefaultModels.FastChat)
			assert.EqualValues(t, "aws-bedrock::unknown::anthropic.claude-instant-v1", siteModelConfig.DefaultModels.CodeCompletion)
		})

		t.Run("ProvisionedThroughput", func(t *testing.T) {
			compConfig := loadCompletionsConfig(schema.Completions{
				Provider:        "aws-bedrock",
				ChatModel:       "anthropic.claude-3-haiku-20240307-v1:0-100k/arn:aws:bedrock:us-west-2:012345678901:provisioned-model/abcdefghijkl",
				CompletionModel: "anthropic.claude-instant-v1",
				// FastChatModel not set, left to default.
				AccessToken: "access-key-id:secret-access-key:session-token",
				Endpoint:    "https://vpce-0000-00000.bedrock-runtime.us-west-2.vpce.amazonaws.com",
			})
			require.NotNil(t, compConfig)

			siteModelConfig := convertCompletionsConfig(compConfig)
			assert.Nil(t, siteModelConfig.SourcegraphModelConfig)
			require.NotNil(t, siteModelConfig.ProviderOverrides)
			require.NotNil(t, siteModelConfig.ModelOverrides)

			// ProviderOverrides.
			require.Equal(t, 1, len(siteModelConfig.ProviderOverrides))
			pOverride := siteModelConfig.ProviderOverrides[0]
			assert.EqualValues(t, "aws-bedrock", pOverride.ID)

			require.NotNil(t, pOverride.ServerSideConfig)
			gProviderInfo := pOverride.ServerSideConfig.GenericProvider
			require.NotNil(t, gProviderInfo)
			assert.Equal(t, "access-key-id:secret-access-key:session-token", gProviderInfo.AccessToken)
			assert.Equal(t, "https://vpce-0000-00000.bedrock-runtime.us-west-2.vpce.amazonaws.com", gProviderInfo.Endpoint)

			// ModelOverrides
			require.Equal(t, 3, len(siteModelConfig.ModelOverrides))

			// Confirm the AWS Provisioned Throughput configuration data is where we expect it.
			chatModelOverride := siteModelConfig.ModelOverrides[0]
			require.EqualValues(t, "aws-bedrock::unknown::anthropic.claude-3-haiku-20240307-v1_0-100k", chatModelOverride.ModelRef)
			// Note the colon in "...v1:0-100k", which we needed to strip out for the ModelID.
			assert.Equal(t, "anthropic.claude-3-haiku-20240307-v1:0-100k", chatModelOverride.ModelName)
			require.NotNil(t, chatModelOverride.ServerSideConfig)
			require.NotNil(t, chatModelOverride.ServerSideConfig.AWSBedrockProvisionedThroughput)
			assert.Equal(t, "arn:aws:bedrock:us-west-2:012345678901:provisioned-model/abcdefghijkl", chatModelOverride.ServerSideConfig.AWSBedrockProvisionedThroughput.ARN)

			// DefaultModels. Note the that model was modified, such as stripping out the ARNM.
			require.NotNil(t, siteModelConfig.DefaultModels)
			assert.EqualValues(t, "aws-bedrock::unknown::anthropic.claude-3-haiku-20240307-v1_0-100k", siteModelConfig.DefaultModels.Chat)
			assert.EqualValues(t, "aws-bedrock::unknown::anthropic.claude-instant-v1", siteModelConfig.DefaultModels.FastChat)
			assert.EqualValues(t, "aws-bedrock::unknown::anthropic.claude-instant-v1", siteModelConfig.DefaultModels.CodeCompletion)
		})
	})
}

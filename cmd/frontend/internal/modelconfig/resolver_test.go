package modelconfig

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"

	modelconfigSDK "github.com/sourcegraph/sourcegraph/internal/modelconfig/types"
)

// Confirm the older behavior when just using the "completions" config.
func TestCompletionsResolver(t *testing.T) {
	// The CodyLLMConfiguration resolver just returns the exact data
	// in the site config. So we just verify things are round-tripped.
	siteConfigData := &conftypes.CompletionsConfig{
		Provider: "provider-from-siteconfig",

		ChatModel:                "chat/model",
		ChatModelMaxTokens:       11111,
		CompletionModel:          "completions-model",
		CompletionModelMaxTokens: 22222,
		FastChatModel:            "fast-chat/model",
		FastChatModelMaxTokens:   33333,

		SmartContextWindow:     "smart-ctx-window",
		DisableClientConfigAPI: true,
	}

	testResolver := &completionsConfigResolver{
		config: siteConfigData,
	}

	t.Run("Provider", func(t *testing.T) {
		assert.EqualValues(t, siteConfigData.Provider, testResolver.Provider())
	})

	t.Run("Settings", func(t *testing.T) {
		assert.Equal(t, siteConfigData.DisableClientConfigAPI, testResolver.DisableClientConfigAPI())

		// For smart context, the value must be exactly "disabled" for it to be disabled.
		assert.Equal(t, "enabled", testResolver.SmartContextWindow())
	})

	t.Run("Models", func(t *testing.T) {
		var (
			model string
			err   error
		)
		model, err = testResolver.ChatModel()
		assert.EqualValues(t, siteConfigData.ChatModel, model)
		assert.NoError(t, err)

		model, err = testResolver.FastChatModel()
		assert.EqualValues(t, siteConfigData.FastChatModel, model)
		assert.NoError(t, err)

		model, err = testResolver.CompletionModel()
		assert.EqualValues(t, siteConfigData.CompletionModel, model)
		assert.NoError(t, err)
	})
}

func TestModelConfigResolver(t *testing.T) {
	// Test data. 3x Providers, each with their own model.

	// AWS Bedrock provider and model.
	awsBedrockProvider := modelconfigSDK.Provider{
		ID: "test-provider_aws-bedrock",
		ServerSideConfig: &modelconfigSDK.ServerSideProviderConfig{
			AWSBedrock: &modelconfigSDK.AWSBedrockProviderConfig{
				AccessToken: "xxx",
			},
		},
	}
	awsBedrockModel := modelconfigSDK.Model{
		ModelRef:  modelconfigSDK.ModelRef("test-provider_aws-bedrock::xxx::test-model_aws-bedrock"),
		ModelName: "aws-bedrock-model-name",
	}

	// Azure OpenAI provider and model.
	azureOpenAIProvider := modelconfigSDK.Provider{
		ID: "test-provider_azure-openai",
		ServerSideConfig: &modelconfigSDK.ServerSideProviderConfig{
			AzureOpenAI: &modelconfigSDK.AzureOpenAIProviderConfig{
				AccessToken: "xxx",
			},
		},
	}
	azureOpenAIModel := modelconfigSDK.Model{
		ModelRef:  modelconfigSDK.ModelRef("test-provider_azure-openai::xxx::test-model_azure-openai"),
		ModelName: "azure-openai-model-name",
	}

	// Cody Gateway provider and model.
	codyGatewayProvider := modelconfigSDK.Provider{
		ID: "test-provider_cody-gateway",
		ServerSideConfig: &modelconfigSDK.ServerSideProviderConfig{
			SourcegraphProvider: &modelconfigSDK.SourcegraphProviderConfig{
				AccessToken: "xxx",
			},
		},
	}
	codyGatewayModel := modelconfigSDK.Model{
		ModelRef:  modelconfigSDK.ModelRef("test-provider_cody-gateway::xxx::test-model_cody-gateway"),
		ModelName: "cody-gateway-model-name",
	}

	modelconfigData := modelconfigSDK.ModelConfiguration{
		Providers: []modelconfigSDK.Provider{
			awsBedrockProvider, azureOpenAIProvider, codyGatewayProvider,
		},
		Models: []modelconfigSDK.Model{
			awsBedrockModel, azureOpenAIModel, codyGatewayModel,
		},
		DefaultModels: modelconfigSDK.DefaultModels{
			Chat:           awsBedrockModel.ModelRef,
			CodeCompletion: azureOpenAIModel.ModelRef,
			FastChat:       codyGatewayModel.ModelRef,
		},
	}

	testResolver := &modelconfigResolver{
		modelconfig: &modelconfigData,
	}

	t.Run("Provider", func(t *testing.T) {
		// The CodeCompletion model is using Azure OpenAI, so the provider returned
		// reflects that.
		assert.EqualValues(t, "azure-openai", testResolver.Provider())

		// Change the CodeCompletion model, confirm the provider is now "sourcegraph."
		modelconfigData.DefaultModels.CodeCompletion = codyGatewayModel.ModelRef
		assert.EqualValues(t, "sourcegraph", testResolver.Provider())

		// Restore.
		modelconfigData.DefaultModels.CodeCompletion = azureOpenAIModel.ModelRef
	})

	t.Run("Settings", func(t *testing.T) {
		// You cannot disable the client config API if using model config.
		assert.Equal(t, false, testResolver.DisableClientConfigAPI())
		// You cannot enable smart context if using model config. (No longer applicable?)
		assert.Equal(t, "disabled", testResolver.SmartContextWindow())
	})

	t.Run("Models", func(t *testing.T) {
		// Note that for all these cases the returned string doesn't match
		// either the Provider ID nor the Model ID. Instead, it is the name
		// of the API Provider (e.g. "sourcegraph" if using Cody Gateway),
		// and we return the model name.
		var (
			model string
			err   error
		)
		model, err = testResolver.ChatModel()
		assert.Equal(t, "aws-bedrock/aws-bedrock-model-name", model)
		assert.NoError(t, err)

		model, err = testResolver.CompletionModel()
		assert.Equal(t, "azure-openai/azure-openai-model-name", model)
		assert.NoError(t, err)

		model, err = testResolver.FastChatModel()
		assert.Equal(t, "sourcegraph/cody-gateway-model-name", model)
		assert.NoError(t, err)
	})
}

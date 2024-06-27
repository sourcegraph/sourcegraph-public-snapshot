package modelconfig

import (
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/modelconfig/types"
)

// convertCompletionsConfig converts the supplied Completions configuration blob (the Cody Enterprise configuration data)
// into the newer types.SiteModelConfiguration structure.
//
// Assumes that the supplied completions object is valid, and contains all the required settings. e.g. the site admin
// can leave some things blank, but `conf/computed.go`'s `GetCompletionsConfig()` will fill the Endpoint and related
// fields with meaingful defaults.
func convertCompletionsConfig(completionsCfg *conftypes.CompletionsConfig) *types.SiteModelConfiguration {
	if completionsCfg == nil {
		return nil
	}

	// convertToModelRef convert the format of model defined in the configuration to
	// a ModelRef, e.g.
	// "foo" => "${completionsCfg.provider}::unknown::foo"
	// "foo/bar" => "foo::unknown::bar"
	//
	// And aws-bedrock provisioned throughput... is handled in its own way.
	convertToModelRef := func(cfgModelID string) string {
		// Common case, where the cfgModelID is just a model ID or a
		// provider "/" model ID.
		if completionsCfg.Provider != "aws-bedrock" {
			slashIdx := strings.Index(cfgModelID, "/")
			if slashIdx == -1 {
				return fmt.Sprintf("%s::unknown::%s", completionsCfg.Provider, cfgModelID)
			}
			encodedProvider := cfgModelID[:slashIdx]
			encodedModel := cfgModelID[slashIdx+1:]
			return fmt.Sprintf("%s::unknown::%s", encodedProvider, encodedModel)
		}

		// If the provider is "aws-bedrock", then things get a little tricky. For
		// The ModelRef, we strip out any provisioned ARNs, as well as any invalid chars.
		cfgModelID = strings.Replace(cfgModelID, ":", "_", -1)
		bedrockModelRef := conftypes.NewBedrockModelRefFromModelID(cfgModelID)
		return fmt.Sprintf("aws-bedrock::unknown::%s", bedrockModelRef.Model)
	}

	chatModelRef := convertToModelRef(completionsCfg.ChatModel)
	fastModelRef := convertToModelRef(completionsCfg.FastChatModel)
	autocompleteModelRef := convertToModelRef(completionsCfg.CompletionModel)

	baseConfig := types.SiteModelConfiguration{
		// Don't use any Sourcegraph-supplied model information, as that would be a breaking change.
		// As Cody Enterprise, via the Completions config, ONLY allows you to specify one model per use-case.
		SourcegraphModelConfig: nil,

		ProviderOverrides: []types.ProviderOverride{
			{
				ID:               types.ProviderID(completionsCfg.Provider),
				ClientSideConfig: nil,
				ServerSideConfig: &types.ServerSideProviderConfig{
					GenericProvider: &types.GenericProviderConfig{
						AccessToken: completionsCfg.AccessToken,
						Endpoint:    completionsCfg.Endpoint,
					},
				},

				DefaultModelConfig: &types.DefaultModelConfig{
					Capabilities: []types.ModelCapability{
						types.ModelCapabilityAutocomplete,
						types.ModelCapabilityChat,
					},
					Category: types.ModelCategoryBalanced,
					Status:   types.ModelStatusStable,
					Tier:     types.ModelTierEnterprise,
				},
			},
		},

		ModelOverrides: []types.ModelOverride{
			{
				ModelRef:    types.ModelRef(chatModelRef),
				DisplayName: completionsCfg.ChatModel,
				ModelName:   completionsCfg.ChatModel,

				ContextWindow: types.ContextWindow{
					MaxInputTokens:  completionsCfg.ChatModelMaxTokens,
					MaxOutputTokens: 4000,
				},
			},
			{
				ModelRef:    types.ModelRef(fastModelRef),
				DisplayName: completionsCfg.FastChatModel,
				ModelName:   completionsCfg.FastChatModel,

				ContextWindow: types.ContextWindow{
					MaxInputTokens:  completionsCfg.FastChatModelMaxTokens,
					MaxOutputTokens: 4000,
				},
			},
			{
				ModelRef:    types.ModelRef(autocompleteModelRef),
				DisplayName: completionsCfg.CompletionModel,
				ModelName:   completionsCfg.CompletionModel,

				ContextWindow: types.ContextWindow{
					MaxInputTokens:  completionsCfg.CompletionModelMaxTokens,
					MaxOutputTokens: 4000,
				},
			},
		},

		DefaultModels: &types.DefaultModels{
			Chat:           types.ModelRef(chatModelRef),
			CodeCompletion: types.ModelRef(autocompleteModelRef),
			FastChat:       types.ModelRef(fastModelRef),
		},
	}

	// Account for the way we encoded the AWS Provisioned Throughput ARN into the site config.
	// We overloaded the model name with both the actual model name and an ARN. So if we see
	// that, add the server-side configuration data.
	if completionsCfg.Provider == "aws-bedrock" {
		for idx := range baseConfig.ModelOverrides {
			modelOverride := &baseConfig.ModelOverrides[idx]

			bedrockModelRef := conftypes.NewBedrockModelRefFromModelID(modelOverride.ModelName)
			if bedrockModelRef.ProvisionedCapacity != nil {
				modelOverride.ModelName = bedrockModelRef.Model
				modelOverride.ServerSideConfig = &types.ServerSideModelConfig{
					AWSBedrockProvisionedThroughput: &types.AWSBedrockProvisionedThroughput{
						ARN: *bedrockModelRef.ProvisionedCapacity,
					},
				}
			}
		}
	}

	return &baseConfig
}

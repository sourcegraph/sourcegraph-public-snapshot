package modelconfig

import (
	"fmt"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/modelconfig/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

// maybeGetSiteModelConfiguration returns the SiteModelConfiguration, or nil if there is none.
//
// This function is responsible for converting the schema.SiteConfiguration that admins write
// into our internal data type.
func maybeGetSiteModelConfiguration(logger log.Logger, siteConfig schema.SiteConfiguration) (*types.SiteModelConfiguration, error) {
	// If "modelConfiguration" is specified, then we respect that and only that. If it is not specified,
	// then we respect the older "completions" configuration.
	modelConfig := siteConfig.ModelConfiguration
	if modelConfig == nil {
		if compConfig := conf.GetCompletionsConfig(siteConfig); compConfig != nil {
			logger.Info("converting completions configuration data", log.String("apiProvider", string(compConfig.Provider)))
			return convertCompletionsConfig(compConfig)
		}
		return nil, nil
	}
	return convertModelConfiguration(modelConfig), nil

}

// Performs no validation, assumes the config is valid.
func convertModelConfiguration(v *schema.SiteModelConfiguration) *types.SiteModelConfiguration {
	return &types.SiteModelConfiguration{
		SourcegraphModelConfig: convertSourcegraphModelConfig(v.Sourcegraph),
		ProviderOverrides:      convertProviderOverrides(v.ProviderOverrides),
		ModelOverrides:         convertModelOverrides(v),
		DefaultModels:          convertDefaultModels(v.DefaultModels),
	}
}

func convertSourcegraphModelConfig(v *schema.SourcegraphModelConfig) *types.SourcegraphModelConfig {
	if v == nil {
		return nil
	}
	return &types.SourcegraphModelConfig{
		Endpoint:        v.Endpoint,
		AccessToken:     v.AccessToken,
		PollingInterval: v.PollingInterval,
		ModelFilters:    convertModelFilters(v.ModelFilters),
	}
}

func convertProviderOverrides(overrides []*schema.ProviderOverride) []types.ProviderOverride {
	var converted []types.ProviderOverride
	for _, v := range overrides {
		converted = append(converted, types.ProviderOverride{
			ID:                 types.ProviderID(v.Id),
			DisplayName:        v.DisplayName,
			ClientSideConfig:   convertClientSideProviderConfig(v.ClientSideConfig),
			ServerSideConfig:   convertServerSideProviderConfig(v.ServerSideConfig),
			DefaultModelConfig: convertDefaultModelConfig(v.DefaultModelConfig),
		})
	}
	return converted
}

func convertModelOverrides(modelConfig *schema.SiteModelConfiguration) []types.ModelOverride {
	var converted []types.ModelOverride
	for _, v := range modelConfig.ModelOverrides {
		converted = append(converted, types.ModelOverride{
			ModelRef:         types.ModelRef(v.ModelRef),
			DisplayName:      v.DisplayName,
			ModelName:        v.ModelName,
			Capabilities:     convertModelCapabilities(v.Capabilities),
			Category:         types.ModelCategory(v.Category),
			Status:           types.ModelStatus(v.Status),
			Tier:             types.ModelTierEnterprise, // Note: we always default to enterprise as the model tier for admin-defined models.
			ContextWindow:    convertContextWindow(v.ContextWindow),
			ClientSideConfig: convertClientSideModelConfig(v.ClientSideConfig),
			ServerSideConfig: convertServerSideModelConfig(v.ServerSideConfig),
		})
	}
	for _, modelRef := range modelConfig.ModelOverridesRecommendedSettings {
		if recommended, ok := recommendedSettings[types.ModelRef(modelRef)]; ok {
			converted = append(converted, recommended)
		}
	}
	return converted
}

func convertDefaultModels(v *schema.DefaultModels) *types.DefaultModels {
	if v == nil {
		return nil
	}
	return &types.DefaultModels{
		Chat:           types.ModelRef(v.Chat),
		FastChat:       types.ModelRef(v.FastChat),
		CodeCompletion: types.ModelRef(v.CodeCompletion),
	}
}

func convertModelFilters(v *schema.ModelFilters) *types.ModelFilters {
	if v == nil {
		return nil
	}
	return &types.ModelFilters{
		StatusFilter: v.StatusFilter,
		Allow:        v.Allow,
		Deny:         v.Deny,
	}
}

func convertClientSideProviderConfig(v *schema.ClientSideProviderConfig) *types.ClientSideProviderConfig {
	if v == nil {
		return nil
	}
	return &types.ClientSideProviderConfig{
		// We currently do not have any known client-side provider configuration.
	}
}

func convertServerSideProviderConfig(cfg *schema.ServerSideProviderConfig) *types.ServerSideProviderConfig {
	if cfg == nil {
		return nil
	}
	if v := cfg.AwsBedrock; v != nil {
		return &types.ServerSideProviderConfig{
			AWSBedrock: &types.AWSBedrockProviderConfig{
				AccessToken: v.AccessToken,
				Endpoint:    v.Endpoint,
				Region:      v.Region,
			},
		}
	} else if v := cfg.AzureOpenAI; v != nil {
		return &types.ServerSideProviderConfig{
			AzureOpenAI: &types.AzureOpenAIProviderConfig{
				AccessToken:                 v.AccessToken,
				Endpoint:                    v.Endpoint,
				User:                        v.User,
				UseDeprecatedCompletionsAPI: v.UseDeprecatedCompletionsAPI,
			},
		}
	} else if v := cfg.Anthropic; v != nil {
		return &types.ServerSideProviderConfig{
			GenericProvider: &types.GenericProviderConfig{
				ServiceName: types.GenericServiceProviderAnthropic,
				AccessToken: v.AccessToken,
				Endpoint:    v.Endpoint,
			},
		}
	} else if v := cfg.Fireworks; v != nil {
		return &types.ServerSideProviderConfig{
			GenericProvider: &types.GenericProviderConfig{
				ServiceName: types.GenericServiceProviderFireworks,
				AccessToken: v.AccessToken,
				Endpoint:    v.Endpoint,
			},
		}
	} else if v := cfg.Google; v != nil {
		return &types.ServerSideProviderConfig{
			GenericProvider: &types.GenericProviderConfig{
				ServiceName: types.GenericServiceProviderGoogle,
				AccessToken: v.AccessToken,
				Endpoint:    v.Endpoint,
			},
		}
	} else if v := cfg.Openai; v != nil {
		return &types.ServerSideProviderConfig{
			GenericProvider: &types.GenericProviderConfig{
				ServiceName: types.GenericServiceProviderOpenAI,
				AccessToken: v.AccessToken,
				Endpoint:    v.Endpoint,
			},
		}
	} else if v := cfg.Openaicompatible; v != nil {
		// TODO(slimsag): self-hosted-llm: map this to OpenAICompatibleProviderConfig in the future
		return &types.ServerSideProviderConfig{
			GenericProvider: &types.GenericProviderConfig{
				ServiceName: types.GenericServiceProviderOpenAI,
				AccessToken: v.AccessToken,
				Endpoint:    v.Endpoint,
			},
		}
	} else if v := cfg.Sourcegraph; v != nil {
		return &types.ServerSideProviderConfig{
			SourcegraphProvider: &types.SourcegraphProviderConfig{
				AccessToken: v.AccessToken,
				Endpoint:    v.Endpoint,
			},
		}
	} else {
		panic(fmt.Sprintf("illegal state: %+v", v))
	}
}

func convertClientSideModelConfig(v *schema.ClientSideModelConfig) *types.ClientSideModelConfig {
	if v == nil {
		return nil
	}
	return &types.ClientSideModelConfig{
		// We currently do not have any known client-side model configuration.
	}
}

func convertServerSideModelConfig(cfg *schema.ServerSideModelConfig) *types.ServerSideModelConfig {
	if cfg == nil {
		return nil
	}
	if v := cfg.AwsBedrockProvisionedThroughput; v != nil {
		return &types.ServerSideModelConfig{
			AWSBedrockProvisionedThroughput: &types.AWSBedrockProvisionedThroughput{
				ARN: v.Arn,
			},
		}
	} else {
		panic(fmt.Sprintf("illegal state: %+v", v))
	}
}

func convertDefaultModelConfig(v *schema.DefaultModelConfig) *types.DefaultModelConfig {
	if v == nil {
		return nil
	}
	return &types.DefaultModelConfig{
		Capabilities:     convertModelCapabilities(v.Capabilities),
		Category:         types.ModelCategory(v.Category),
		Status:           types.ModelStatus(v.Status),
		Tier:             types.ModelTierEnterprise, // Note: we always default to enterprise as the model tier for admin-defined models.
		ContextWindow:    convertContextWindow(v.ContextWindow),
		ClientSideConfig: convertClientSideModelConfig(v.ClientSideConfig),
		ServerSideConfig: convertServerSideModelConfig(v.ServerSideConfig),
	}
}

func convertContextWindow(v schema.ContextWindow) types.ContextWindow {
	return types.ContextWindow{
		MaxInputTokens:  v.MaxInputTokens,
		MaxOutputTokens: v.MaxOutputTokens,
	}
}

func convertModelCapabilities(capabilities []string) []types.ModelCapability {
	var converted []types.ModelCapability
	for _, v := range capabilities {
		converted = append(converted, types.ModelCapability(v))
	}
	return converted
}

// These are the default values where if someone writes in their site config that they want to
// use blessed Self-Hosted Model configurations, e.g.:
//
// ```
// "modelOverridesRecommendedSettings": [
//
//	"bigcode::v1::starcoder2-7b",
//	"mistral::v1::mixtral-8x7b-instruct"
//
// ],
// ```
//
// It would specify these equivalent options for them under `modelOverrides`:
var recommendedSettings = map[types.ModelRef]types.ModelOverride{
	"bigcode::v1::starcoder2-3b":          recommendedSettingsStarcoder2("bigcode::v1::starcoder2-3b", "Starcoder2 3B", "starcoder2-3b"),
	"bigcode::v1::starcoder2-7b":          recommendedSettingsStarcoder2("bigcode::v1::starcoder2-7b", "Starcoder2 7B", "starcoder2-7b"),
	"bigcode::v1::starcoder2-15b":         recommendedSettingsStarcoder2("bigcode::v1::starcoder2-15b", "Starcoder2 15B", "starcoder2-15b"),
	"mistral::v1::mistral-7b":             recommendedSettingsMistral("mistral::v1::mistral-7b", "Mistral 7B", "mistral-7b"),
	"mistral::v1::mistral-7b-instruct":    recommendedSettingsMistral("mistral::v1::mistral-7b-instruct", "Mistral 7B Instruct", "mistral-7b-instruct"),
	"mistral::v1::mixtral-8x7b":           recommendedSettingsMistral("mistral::v1::mixtral-8x7b", "Mixtral 8x7B", "mixtral-8x7b"),
	"mistral::v1::mixtral-8x22b":          recommendedSettingsMistral("mistral::v1::mixtral-8x22b", "Mixtral 8x22B", "mixtral-8x22b"),
	"mistral::v1::mixtral-8x7b-instruct":  recommendedSettingsMistral("mistral::v1::mixtral-8x7b-instruct", "Mixtral 8x7B Instruct", "mixtral-8x7b-instruct"),
	"mistral::v1::mixtral-8x22b-instruct": recommendedSettingsMistral("mistral::v1::mixtral-8x22b", "Mixtral 8x22B", "mixtral-8x22b-instruct"),
}

func recommendedSettingsStarcoder2(modelRef, displayName, modelName string) types.ModelOverride {
	// TODO(slimsag): self-hosted-llm: tune these further based on testing
	return types.ModelOverride{
		ModelRef:     types.ModelRef(modelRef),
		DisplayName:  displayName,
		ModelName:    modelName,
		Capabilities: []types.ModelCapability{types.ModelCapabilityAutocomplete},
		Category:     types.ModelCategoryBalanced,
		Status:       types.ModelStatusStable,
		Tier:         types.ModelTierEnterprise,
		ContextWindow: types.ContextWindow{
			MaxInputTokens:  8192,
			MaxOutputTokens: 4000,
		},
		ClientSideConfig: nil,
		ServerSideConfig: nil,
	}
}

func recommendedSettingsMistral(modelRef, displayName, modelName string) types.ModelOverride {
	// TODO(slimsag): self-hosted-llm: tune these further based on testing
	return types.ModelOverride{
		ModelRef:     types.ModelRef(modelRef),
		DisplayName:  displayName,
		ModelName:    modelName,
		Capabilities: []types.ModelCapability{types.ModelCapabilityChat, types.ModelCapabilityAutocomplete},
		Category:     types.ModelCategoryBalanced,
		Status:       types.ModelStatusStable,
		Tier:         types.ModelTierEnterprise,
		ContextWindow: types.ContextWindow{
			MaxInputTokens:  8192,
			MaxOutputTokens: 4000,
		},
		ClientSideConfig: nil,
		ServerSideConfig: nil,
	}
}

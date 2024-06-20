package main

import (
	"github.com/sourcegraph/sourcegraph/internal/modelconfig/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const (
	// Maximum number of tokens Cody Gateway allows to be sent in responses.
	// BEFORE you can change this value, you MUST first update Cody Gateway's
	// abuse-configuration settings. Otherwise the calling user will be flagged
	// and/or banned!
	maxCodyGatewayOutputTokens = 4000
)

var (
	chatAndEdit = []types.ModelCapability{
		types.ModelCapabilityAutocomplete,
		types.ModelCapabilityChat,
	}

	// Standard context window sizes.
	standardCtxWindow = types.ContextWindow{
		MaxInputTokens:  7_000,
		MaxOutputTokens: maxCodyGatewayOutputTokens,
	}
	// Higher context window for newer LLMs.
	expandedCtxWindow = types.ContextWindow{
		MaxInputTokens:  30_000,
		MaxOutputTokens: maxCodyGatewayOutputTokens,
	}
)

func getAnthropicModels() []types.Model {
	const (
		// Sourcegraph [v5.1 - v5.3) use the legacy "Text Completions" API.
		// https://docs.anthropic.com/en/api/complete
		anthropic_01_2023 = "anthropic::2023-01-01"
		// Sourcegraph v5.3+ uses the newer "Messages API".
		// https://docs.anthropic.com/en/api/messages
		//
		// This doesn't directly map to the Anthropic API release, but
		// it seems cleaner than introducing "v1/messages".
		// https://docs.anthropic.com/en/api/versioning
		anthropic_06_2023 = "anthropic::2023-06-01"
	)

	return []types.Model{
		// Free Anthropic model.
		newModel(
			modelIdentity{
				MRef:        mRef(anthropic_06_2023, "claude-3-sonnet"),
				Name:        "claude-3-sonnet-20240229",
				DisplayName: "Claude 3 Sonnet",
			},
			modelMetadata{
				Capabilities: chatAndEdit,
				Category:     types.ModelCategoryBalanced,
				Status:       types.ModelStatusStable,
				Tier:         types.ModelTierFree,
			},
			expandedCtxWindow),

		// Pro Anthropic models.
		newModel(
			modelIdentity{
				MRef:        mRef(anthropic_06_2023, "claude-3.5-sonnet"),
				Name:        "claude-3.5-sonnet-20240620",
				DisplayName: "Claude 3.5 Sonnet",
			},
			modelMetadata{
				Capabilities: chatAndEdit,
				Category:     types.ModelCategoryAccuracy,
				Status:       types.ModelStatusStable,
				Tier:         types.ModelTierPro,
			},
			expandedCtxWindow),
		newModel(
			modelIdentity{
				MRef:        mRef(anthropic_06_2023, "claude-3-opus"),
				Name:        "claude-3-opus-20240229",
				DisplayName: "Claude 3 Opus",
			},
			modelMetadata{
				Capabilities: chatAndEdit,
				Category:     types.ModelCategoryAccuracy,
				Status:       types.ModelStatusStable,
				Tier:         types.ModelTierPro,
			},
			expandedCtxWindow),
		newModel(
			modelIdentity{
				MRef:        mRef(anthropic_06_2023, "claude-3-haiku"),
				Name:        "claude-3-haiku-20240307",
				DisplayName: "Claude 3 Haiku",
			},
			modelMetadata{
				Capabilities: chatAndEdit,
				Category:     types.ModelCategorySpeed,
				Status:       types.ModelStatusStable,
				Tier:         types.ModelTierPro,
			},
			standardCtxWindow),

		// Older Claude 2.x models, now deprecated.
		newModel(
			modelIdentity{
				MRef:        mRef(anthropic_01_2023, "claude-2.1"),
				Name:        "claude-2.1",
				DisplayName: "Claude 2.1",
			},
			modelMetadata{
				Capabilities: chatAndEdit,
				Category:     types.ModelCategoryBalanced,
				Status:       types.ModelStatusDeprecated,
				Tier:         types.ModelTierFree,
			},
			standardCtxWindow),
		newModel(
			modelIdentity{
				MRef:        mRef(anthropic_01_2023, "claude-2.0"),
				Name:        "claude-2.0",
				DisplayName: "Claude 2.0",
			},
			modelMetadata{
				Capabilities: chatAndEdit,
				Category:     types.ModelCategoryBalanced,
				Status:       types.ModelStatusDeprecated,
				Tier:         types.ModelTierFree,
			},
			standardCtxWindow),
		newModel(
			modelIdentity{
				MRef:        mRef(anthropic_01_2023, "claude-instant-1.2"),
				Name:        "claude-instant-1.2",
				DisplayName: "Claude Instant",
			},
			modelMetadata{
				Capabilities: chatAndEdit,
				Category:     types.ModelCategoryBalanced,
				Status:       types.ModelStatusDeprecated,
				Tier:         types.ModelTierFree,
			},
			standardCtxWindow),
	}
}

func getMistralModels() []types.Model {
	// Not sure if there is a canonical API reference, since Mixtral offers 3rd
	// party LLMs as a service.
	// https://deepinfra.com/mistralai/Mixtral-8x22B-Instruct-v0.1/api
	// https://readme.fireworks.ai
	const mistralV1 = "mistral::v1"

	return []types.Model{
		newModel(
			modelIdentity{
				MRef:        mRef(mistralV1, "mixtral-8x7b-instruct"),
				Name:        "mixtral-8x7b-instruct",
				DisplayName: "Mixtral 8x7B",
			},
			modelMetadata{
				Capabilities: chatAndEdit,
				Category:     types.ModelCategorySpeed,
				Status:       types.ModelStatusStable,
				Tier:         types.ModelTierPro,
			},
			standardCtxWindow),
		newModel(
			modelIdentity{
				MRef:        mRef(mistralV1, "mixtral-8x22b-instruct"),
				Name:        "mixtral-8x22b-instruct",
				DisplayName: "Mixtral 8x22B",
			},
			modelMetadata{
				Capabilities: chatAndEdit,
				Category:     types.ModelCategoryAccuracy,
				Status:       types.ModelStatusStable,
				Tier:         types.ModelTierPro,
			},
			standardCtxWindow),
	}
}

func getOpenAIModels() []types.Model {
	// https://platform.openai.com/docs/gpts/release-notes
	// https://platform.openai.com/docs/deprecations
	const openAIV1 = "openai::2024-02-01"

	return []types.Model{
		newModel(
			modelIdentity{
				MRef:        mRef(openAIV1, "gpt-4o"),
				Name:        "gpt-4o",
				DisplayName: "GPT-4o",
			},
			modelMetadata{
				Capabilities: chatAndEdit,
				Category:     types.ModelCategoryAccuracy,
				Status:       types.ModelStatusStable,
				Tier:         types.ModelTierPro,
			},
			expandedCtxWindow),
		newModel(
			modelIdentity{
				MRef:        mRef(openAIV1, "gpt-4-turbo"),
				Name:        "gpt-4-turbo",
				DisplayName: "GPT-4 Turbo",
			},
			modelMetadata{
				Capabilities: chatAndEdit,
				Category:     types.ModelCategoryAccuracy,
				Status:       types.ModelStatusStable,
				Tier:         types.ModelTierPro,
			},
			standardCtxWindow),
		newModel(
			modelIdentity{
				MRef:        mRef(openAIV1, "gpt-3.5-turbo"),
				Name:        "gpt-3.5-turbo",
				DisplayName: "GPT-3.5 Turbo",
			},
			modelMetadata{
				Capabilities: chatAndEdit,
				Category:     types.ModelCategorySpeed,
				Status:       types.ModelStatusStable,
				Tier:         types.ModelTierPro,
			},
			standardCtxWindow),
	}
}

// GetCodyFreeProModels returns the current list of models supported for Cody
// Free and Cody Pro users.
func GetCodyFreeProModels() ([]types.Model, error) {
	// ================================================
	// 👇 Models available to Free/Pro users go HERE 👇
	// ================================================
	var allModels []types.Model
	allModels = append(allModels, getAnthropicModels()...)
	allModels = append(allModels, getMistralModels()...)
	allModels = append(allModels, getOpenAIModels()...)

	// Confirm that only PLG models are defined in this function.
	// (Presuming that later, when we add Enterprise-only models, we
	// would want to describe them in another file.)
	for _, model := range allModels {
		if model.Tier != types.ModelTierFree && model.Tier != types.ModelTierPro {
			return nil, errors.Errorf("model %q is not configued for the Free/Pro tier", model.ModelRef)
		}
	}

	return allModels, nil
}

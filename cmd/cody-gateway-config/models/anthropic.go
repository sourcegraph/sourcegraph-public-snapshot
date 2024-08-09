package models

import "github.com/sourcegraph/sourcegraph/internal/modelconfig/types"

func AllAnthropicModels() []types.Model {
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
				Name:        "claude-3-5-sonnet-20240620",
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

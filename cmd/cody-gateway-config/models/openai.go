package models

import "github.com/sourcegraph/sourcegraph/internal/modelconfig/types"

func AllOpenAIModels() []types.Model {
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
				Category:     types.ModelCategoryBalanced,
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

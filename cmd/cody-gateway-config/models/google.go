package models

import "github.com/sourcegraph/sourcegraph/internal/modelconfig/types"

func AllGoogleModels() []types.Model {
	const (
		// Gemini API versions.
		// https://ai.google.dev/gemini-api/docs/api-versions
		geminiV1 = "google::v1"
	)

	return []types.Model{
		newModel(
			modelIdentity{
				MRef:        mRef(geminiV1, "gemini-1.5-pro"),
				Name:        "gemini-1.5-pro-latest",
				DisplayName: "Gemini 1.5 Pro",
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
				MRef:        mRef(geminiV1, "gemini-1.5-flash"),
				Name:        "gemini-1.5-flash-latest",
				DisplayName: "Gemini 1.5 Flash",
			},
			modelMetadata{
				Capabilities: chatAndEdit,
				Category:     types.ModelCategorySpeed,
				Status:       types.ModelStatusStable,
				Tier:         types.ModelTierPro,
			},
			expandedCtxWindow),
	}
}

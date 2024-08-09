package models

import "github.com/sourcegraph/sourcegraph/internal/modelconfig/types"

func AllMistralModels() []types.Model {
	// NOTE: These are all kinda fubar, since we are offering Mixtral models
	// via the Fireworks API provider.
	//
	// So the ModelNames do need to have this odd format. Because there isn't an
	// actual "Mistral API Provider" in our backend, we route all of these to
	// Fireworks.
	// https://deepinfra.com/mistralai/Mixtral-8x22B-Instruct-v0.1/api
	// https://readme.fireworks.ai
	const mistralV1 = "mistral::v1"

	return []types.Model{
		newModel(
			modelIdentity{
				MRef:        mRef(mistralV1, "mixtral-8x7b-instruct"),
				Name:        "accounts/fireworks/models/mixtral-8x7b-instruct",
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
				Name:        "accounts/fireworks/models/mixtral-8x22b-instruct",
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

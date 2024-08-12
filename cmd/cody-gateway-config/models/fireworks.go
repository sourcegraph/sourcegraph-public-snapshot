package models

import (
	"github.com/sourcegraph/sourcegraph/internal/modelconfig/types"
)

func AllFireworksModels() []types.Model {
	// https://docs.fireworks.ai/api-reference/post-completions
	const fireworksV1 = "fireworks::v1"

	metadata := modelMetadata{
		Capabilities: editOnly,
		Category:     types.ModelCategorySpeed,
		Status:       types.ModelStatusStable,

		// Both of our code completion models are available
		// to Cody Free users.
		Tier: types.ModelTierFree,
	}
	contextWindow := types.ContextWindow{
		// These values are much lower than other, text-centric models. We are
		// erring on the side of matching the token limits defined on the client
		// today. (And maybe the StarCoder is able to use a more efficient
		// tokenizer, because it's not processing many languages.)
		// https://github.com/sourcegraph/cody/blob/066d9c6ff48beb96a834f17021affc4e62094415/vscode/src/completions/providers/fireworks.ts#L132
		// https://github.com/sourcegraph/cody/blob/066d9c6ff48beb96a834f17021affc4e62094415/vscode/src/completions/providers/get-completion-params.ts#L5
		MaxInputTokens:  2048,
		MaxOutputTokens: 256,
	}

	return []types.Model{
		newModel(
			modelIdentity{
				MRef: mRef(fireworksV1, "starcoder"),
				// NOTE: This model name is virtualized.
				//
				// When Cody Gateway receives a request using model
				// "fireworks/starcoder", it will then pick a specialized
				// model name such as "starcoder2-15b" or "starcoder-7b".
				Name:        "starcoder",
				DisplayName: "StarCoder",
			},
			metadata,
			contextWindow),

		// StarCoder variants that are expected from the client, via
		// `vscode/src/completions/providers/fireworks.ts`.
		newModel(
			modelIdentity{
				MRef:        mRef(fireworksV1, "starcoder-7b"),
				Name:        "starcoder-7b",
				DisplayName: "StarCoder 7B",
			},
			metadata,
			contextWindow),
		newModel(
			modelIdentity{
				MRef:        mRef(fireworksV1, "starcoder-16b"),
				Name:        "starcoder-16b",
				DisplayName: "StarCoder 16B",
			},
			metadata,
			contextWindow),
		newModel(
			modelIdentity{
				MRef:        mRef(fireworksV1, "starcoder2-7b"),
				Name:        "starcoder2-7b",
				DisplayName: "StarCoder2 7B",
			},
			metadata,
			contextWindow),
		newModel(
			modelIdentity{
				MRef:        mRef(fireworksV1, "starcoder2-15b"),
				Name:        "starcoder2-15b",
				DisplayName: "StarCoder2 15B",
			},
			metadata,
			contextWindow),

		// Deepseek models.
		newModel(
			modelIdentity{
				MRef:        mRef(fireworksV1, "deepseek"),
				Name:        "accounts/sourcegraph/models/deepseek-coder-v2-lite-base",
				DisplayName: "DeepSeek",
			},
			metadata,
			contextWindow),
	}
}

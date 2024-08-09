package models

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/modelconfig/types"
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
	editOnly = []types.ModelCapability{
		types.ModelCapabilityAutocomplete,
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

func mRef(providerApiVerRef, modelID string) types.ModelRef {
	raw := fmt.Sprintf("%s::%s", providerApiVerRef, modelID)
	return types.ModelRef(raw)
}

type modelIdentity struct {
	MRef        types.ModelRef
	Name        string
	DisplayName string
}

type modelMetadata struct {
	Capabilities []types.ModelCapability
	Category     types.ModelCategory
	Status       types.ModelStatus
	Tier         types.ModelTier
}

func newModel(identity modelIdentity, metadata modelMetadata, ctxWindow types.ContextWindow) types.Model {
	return types.Model{
		ModelRef:    identity.MRef,
		DisplayName: identity.DisplayName,
		ModelName:   identity.Name,

		Capabilities: metadata.Capabilities,
		Category:     metadata.Category,
		Status:       metadata.Status,
		Tier:         metadata.Tier,

		ContextWindow: ctxWindow,
	}
}

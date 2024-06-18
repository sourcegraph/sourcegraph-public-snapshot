package main

import (
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/modelconfig/types"
)

func mRef(providerApiVerRef, modelID string) types.ModelRef {
	raw := fmt.Sprintf("%s::%s", providerApiVerRef, modelID)
	return types.ModelRef(raw)
}

func newProvider(id, name string) types.Provider {
	return types.Provider{
		ID:          types.ProviderID(id),
		DisplayName: name,
	}
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
	// Tease out the ModelID from the ModelRef. This isn't safe, since the
	// supplied ModelRef might be invalid. But we will validate the data
	// before writing it out.
	modelID := types.ModelID("unknown")
	parts := strings.Split(string(identity.MRef), "::")
	if len(parts) != 3 {
		modelID = types.ModelID(parts[2])
	}

	return types.Model{
		ID:          modelID,
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

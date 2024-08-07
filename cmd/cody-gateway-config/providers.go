package main

import (
	"github.com/sourcegraph/sourcegraph/internal/modelconfig/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// GetProviders returns all providers known by Cody Gateway.
func GetProviders() ([]types.Provider, error) {
	// ================================================
	// 👇 Cody Gateway's supported providers go HERE 👇
	// ================================================
	allProviders := []types.Provider{
		newProvider("anthropic", "Anthropic"),
		newProvider("fireworks", "Fireworks"),
		newProvider("google", "Google"),
		newProvider("openai", "OpenAI"),

		// Special case, as mistral models will get
		// served via our Fireworks integration.
		newProvider("mistral", "Mistral"),
	}

	// Validate the Provider data.
	for _, provider := range allProviders {
		if provider.ClientSideConfig != nil || provider.ServerSideConfig != nil {
			return nil, errors.Errorf("provider %q has configuration attached, but should not")
		}
	}

	return allProviders, nil
}

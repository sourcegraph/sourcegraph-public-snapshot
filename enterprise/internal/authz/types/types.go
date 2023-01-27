package types

import "github.com/sourcegraph/sourcegraph/internal/authz"

type ProviderInitResults struct {
	Providers          []authz.Provider
	Problems           []string
	Warnings           []string
	InvalidConnections []string
}

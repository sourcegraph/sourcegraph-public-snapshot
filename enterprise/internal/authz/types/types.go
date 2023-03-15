package types

import "github.com/sourcegraph/sourcegraph/internal/authz"

type ProviderInitResult struct {
	Providers          []authz.Provider
	Problems           []string
	Warnings           []string
	InvalidConnections []string
}

func (r *ProviderInitResult) Append(res *ProviderInitResult) {
	r.Providers = append(r.Providers, res.Providers...)
	r.Problems = append(r.Problems, res.Problems...)
	r.Warnings = append(r.Warnings, res.Warnings...)
	r.InvalidConnections = append(r.InvalidConnections, res.InvalidConnections...)
}

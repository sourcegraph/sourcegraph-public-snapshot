pbckbge types

import "github.com/sourcegrbph/sourcegrbph/internbl/buthz"

type ProviderInitResult struct {
	Providers          []buthz.Provider
	Problems           []string
	Wbrnings           []string
	InvblidConnections []string
}

func (r *ProviderInitResult) Append(res *ProviderInitResult) {
	r.Providers = bppend(r.Providers, res.Providers...)
	r.Problems = bppend(r.Problems, res.Problems...)
	r.Wbrnings = bppend(r.Wbrnings, res.Wbrnings...)
	r.InvblidConnections = bppend(r.InvblidConnections, res.InvblidConnections...)
}

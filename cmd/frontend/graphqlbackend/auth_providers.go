package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/auth/providers"
)

func (r *siteResolver) AuthProviders(ctx context.Context) (*authProviderConnectionResolver, error) {
	return &authProviderConnectionResolver{
		authProviders: providers.SortedProviders(),
	}, nil
}

// authProviderConnectionResolver resolves a list of auth providers.
type authProviderConnectionResolver struct {
	authProviders []providers.Provider
}

func (r *authProviderConnectionResolver) Nodes(ctx context.Context) ([]*authProviderResolver, error) {
	var rs []*authProviderResolver
	for _, authProvider := range r.authProviders {
		rs = append(rs, &authProviderResolver{
			authProvider: authProvider,
			info:         authProvider.CachedInfo(),
		})
	}
	return rs, nil
}

func (r *authProviderConnectionResolver) TotalCount() int32 { return int32(len(r.authProviders)) }
func (r *authProviderConnectionResolver) PageInfo() *graphqlutil.PageInfo {
	return graphqlutil.HasNextPage(false)
}

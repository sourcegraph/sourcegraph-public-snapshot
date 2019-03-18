package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth"
)

func (r *siteResolver) AuthProviders(ctx context.Context) (*authProviderConnectionResolver, error) {
	// 🚨 SECURITY: Only site admins can list auth providers.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	return &authProviderConnectionResolver{
		authProviders: auth.Providers(),
	}, nil
}

// authProviderConnectionResolver resolves a list of auth providers.
//
// 🚨 SECURITY: When instantiating an authProviderConnectionResolver value, the caller MUST check
// permissions.
type authProviderConnectionResolver struct {
	authProviders []auth.Provider
}

func (r *authProviderConnectionResolver) Nodes(ctx context.Context) ([]*authProviderResolver, error) {
	// 🚨 SECURITY: Only site admins can list auth providers. This check is intentionally redundant
	// (with the check in (*siteResolver).AuthProviders), to reduce the likelihood of a bug causing
	// this information to leak.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

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

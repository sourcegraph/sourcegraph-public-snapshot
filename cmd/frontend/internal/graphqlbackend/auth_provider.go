package graphqlbackend

import "github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth"

// authProviderResolver resolves an auth provider.
type authProviderResolver struct {
	authProvider *auth.Provider
}

func (r *authProviderResolver) ServiceType() string { return r.authProvider.ProviderID.ServiceType }
func (r *authProviderResolver) DisplayName() string { return r.authProvider.Public.DisplayName }
func (r *authProviderResolver) IsBuiltin() bool     { return r.authProvider.Public.IsBuiltin }
func (r *authProviderResolver) AuthenticationURL() *string {
	if u := r.authProvider.Public.AuthenticationURL; u != "" {
		return &u
	}
	return nil
}

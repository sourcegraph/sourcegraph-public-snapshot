package graphqlbackend

import "github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth"

// authProviderResolver resolves an auth provider.
type authProviderResolver struct {
	authProvider auth.Provider

	info *auth.ProviderInfo // == authProvider.CachedInfo()
}

func (r *authProviderResolver) ServiceType() string { return r.authProvider.ID().Type }
func (r *authProviderResolver) ServiceID() string   { return r.authProvider.ID().ID }
func (r *authProviderResolver) ClientID() string    { return "" }
func (r *authProviderResolver) DisplayName() string { return r.info.DisplayName }
func (r *authProviderResolver) IsBuiltin() bool     { return r.authProvider.Config().Builtin != nil }
func (r *authProviderResolver) AuthenticationURL() *string {
	if u := r.info.AuthenticationURL; u != "" {
		return &u
	}
	return nil
}

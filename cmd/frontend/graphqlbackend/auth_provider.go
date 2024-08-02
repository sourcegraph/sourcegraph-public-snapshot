package graphqlbackend

import "github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth/providers"

// authProviderResolver resolves an auth provider.
type authProviderResolver struct {
	authProvider providers.Provider

	info *providers.Info // == authProvider.CachedInfo()
}

func (r *authProviderResolver) ServiceType() string { return r.authProvider.ConfigID().Type }

func (r *authProviderResolver) ServiceID() string {
	if r.info != nil {
		return r.info.ServiceID
	}
	return ""
}

func (r *authProviderResolver) ClientID() string {
	if r.info != nil {
		return r.info.ClientID
	}
	return ""
}

func (r *authProviderResolver) DisplayName() string { return r.info.DisplayName }
func (r *authProviderResolver) IsBuiltin() bool     { return r.authProvider.Config().Builtin != nil }
func (r *authProviderResolver) AuthenticationURL() *string {
	if u := r.info.AuthenticationURL; u != "" {
		return &u
	}
	return nil
}

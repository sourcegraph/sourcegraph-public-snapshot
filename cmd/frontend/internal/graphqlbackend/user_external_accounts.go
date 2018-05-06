package graphqlbackend

import (
	"context"
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/backend"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
)

func (r *userResolver) ExternalAccounts(ctx context.Context) ([]*externalAccountResolver, error) {
	// 🚨 SECURITY: Only the user and site admins should be able to see the user's external accounts.
	if err := backend.CheckSiteAdminOrSameUser(ctx, r.user.ID); err != nil {
		return nil, err
	}

	var accounts []*externalAccountResolver
	if r.user.ExternalProvider != "" {
		// TODO(sqs): Store actual auth provider at the time this external account was created
		// (might differ from the current auth provider).

		account := &externalAccountResolver{
			user:      r,
			serviceID: r.user.ExternalProvider,
		}

		authProvider := conf.AuthProvider()
		if authProvider.Builtin == nil {
			account.canAuthenticate = true
		}
		var providerType string
		switch {
		case authProvider.Openidconnect != nil:
			providerType = "OpenID"
		case authProvider.Saml != nil:
			providerType = "SAML"
		case authProvider.HttpHeader != nil:
			providerType = "Web authentication proxy"
		default:
			providerType = "Unknown authentication provider"
		}
		// This is just for convenience, so correctness is not important.
		var providerName string
		switch {
		case strings.Contains(r.user.ExternalProvider, "https://accounts.google.com"):
			providerName = "Google/G Suite"
		case strings.Contains(r.user.ExternalProvider, ".okta"):
			providerName = "Okta"
		case strings.Contains(r.user.ExternalProvider, "onelogin.com"):
			providerName = "OneLogin"
		}
		if providerName != "" {
			account.serviceName = fmt.Sprintf("%s (%s)", providerName, providerType)
		} else {
			account.serviceName = providerType
		}

		if r.user.ExternalID != nil {
			account.serviceUserID = *r.user.ExternalID
		}

		accounts = append(accounts, account)
	}

	return accounts, nil
}

type externalAccountResolver struct {
	user            *userResolver
	serviceName     string
	serviceID       string
	serviceUserID   string
	canAuthenticate bool
}

func (r *externalAccountResolver) User() *userResolver   { return r.user }
func (r *externalAccountResolver) ServiceName() string   { return r.serviceName }
func (r *externalAccountResolver) ServiceID() string     { return r.serviceID }
func (r *externalAccountResolver) ServiceUserID() string { return r.serviceUserID }
func (r *externalAccountResolver) CanAuthenticate() bool { return r.canAuthenticate }

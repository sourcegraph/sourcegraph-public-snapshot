package openidconnect

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strconv"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/schema"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

var mockGetProviderValue *provider

// getProvider looks up the registered openidconnect auth provider with the given ID.
func getProvider(id string) *provider {
	if mockGetProviderValue != nil {
		return mockGetProviderValue
	}
	p, _ := auth.GetProvider(auth.ProviderID{Type: providerType, ID: id}).(*provider)
	return p
}

func handleGetProvider(ctx context.Context, w http.ResponseWriter, id string) (p *provider, handled bool) {
	handled = true // safer default

	p = getProvider(id)
	if p == nil {
		log15.Error("No OpenID Connect auth provider found with ID.", "id", id)
		http.Error(w, "Misconfigured OpenID Connect auth provider.", http.StatusInternalServerError)
		return nil, true
	}
	if p.config.Issuer == "" {
		log15.Error("No issuer set for OpenID Connect auth provider (set the openidconnect auth provider's issuer property).", "id", p.ID())
		http.Error(w, "Misconfigured OpenID Connect auth provider.", http.StatusInternalServerError)
		return nil, true
	}
	if err := p.Refresh(ctx); err != nil {
		log15.Error("Error refreshing OpenID Connect auth provider.", "id", p.ID(), "error", err)
		http.Error(w, "Unexpected error refreshing OpenID Connect authentication provider.", http.StatusInternalServerError)
		return nil, true
	}
	return p, false
}

type providerID struct{ issuerURL, clientID string }

func (k providerID) KeyString() string {
	b := sha256.Sum256([]byte(strconv.Itoa(len(k.issuerURL)) + ":" + k.issuerURL + ":" + k.clientID))
	return hex.EncodeToString(b[:10])
}

func toProviderID(pc *schema.OpenIDConnectAuthProvider) providerID {
	return providerID{issuerURL: pc.Issuer, clientID: pc.ClientID}
}

func validateConfig(c *schema.SiteConfiguration) (problems []string) {
	var loggedNeedsAppURL bool
	for _, p := range conf.AuthProvidersFromConfig(c) {
		if p.Openidconnect != nil && c.AppURL == "" && !loggedNeedsAppURL {
			problems = append(problems, `openidconnect auth provider requires appURL to be set to the external URL of your site (example: https://sourcegraph.example.com)`)
			loggedNeedsAppURL = true
		}
		if p.Openidconnect != nil && p.Openidconnect.OverrideToken != "" {
			problems = append(problems, `openidconnect auth provider "overrideToken" is deprecated (because it applies to all auth providers, not just OIDC); use OVERRIDE_AUTH_SECRET env var instead`)
		}
	}

	hasOldOIDC := c.OidcProvider != "" || c.OidcClientID != "" || c.OidcClientSecret != "" || c.OidcEmailDomain != ""
	hasSingularOIDC := c.AuthOpenIDConnect != nil
	if hasOldOIDC {
		problems = append(problems, `oidc* properties are deprecated; use auth provider "openidconnect" instead`)
	}
	if c.AuthProvider == "openidconnect" && !hasSingularOIDC {
		problems = append(problems, `auth.openIDConnect must be configured when auth.provider == "openidconnect"`)
	}
	if hasOldOIDC && c.AuthProvider != "openidconnect" {
		problems = append(problems, `must set auth.provider == "openidconnect" for oidc* config to take effect (also, oidc* config is deprecated; see other message to that effect)`)
	}
	if hasSingularOIDC && c.AuthProvider != "openidconnect" {
		problems = append(problems, `must set auth.provider == "openidconnect" for auth.openIDConnect config to take effect`)
	}
	return problems
}

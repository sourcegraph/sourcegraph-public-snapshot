package openidconnect

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strconv"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth"
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

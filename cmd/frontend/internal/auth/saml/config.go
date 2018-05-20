package saml

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/schema"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

var mockGetProviderValue *provider

// getProvider looks up the registered saml auth provider with the given ID.
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
		log15.Error("No SAML auth provider found with ID.", "id", id)
		http.Error(w, "Misconfigured SAML auth provider.", http.StatusInternalServerError)
		return nil, true
	}
	if err := p.Refresh(ctx); err != nil {
		log15.Error("Error refreshing SAML auth provider.", "id", p.ID(), "error", err)
		http.Error(w, "Unexpected error refreshing SAML authentication provider.", http.StatusInternalServerError)
		return nil, true
	}
	return p, false
}

func validateConfig(c *schema.SiteConfiguration) (problems []string) {
	var loggedNeedsAppURL bool
	for _, p := range conf.AuthProvidersFromConfig(c) {
		if p.Saml != nil && c.AppURL == "" && !loggedNeedsAppURL {
			problems = append(problems, `saml auth provider requires appURL to be set to the external URL of your site (example: https://sourcegraph.example.com)`)
			loggedNeedsAppURL = true
		}
	}

	seen := map[schema.SAMLAuthProvider]int{}
	for i, p := range conf.AuthProvidersFromConfig(c) {
		if p.Saml != nil {
			if j, ok := seen[*p.Saml]; ok {
				problems = append(problems, fmt.Sprintf("SAML auth provider at index %d is duplicate of index %d, ignoring", i, j))
			} else {
				seen[*p.Saml] = i
			}
		}
	}

	hasOldSAML := c.SamlIDProviderMetadataURL != "" || c.SamlSPCert != "" || c.SamlSPKey != ""
	hasSingularSAML := c.AuthSaml != nil
	if hasOldSAML {
		problems = append(problems, `saml* properties are deprecated; use auth provider "saml" instead`)
	}
	if c.AuthProvider == "saml" && !hasSingularSAML {
		problems = append(problems, `auth.saml must be configured when auth.provider == "saml"`)
	}
	if hasOldSAML && c.AuthProvider != "saml" {
		problems = append(problems, `must set auth.provider == "saml" for saml* config to take effect (also, saml* config is deprecated; see other message to that effect)`)
	}
	if hasSingularSAML && c.AuthProvider != "saml" {
		problems = append(problems, `must set auth.provider == "saml" for auth.saml config to take effect`)
	}
	return problems
}

type providerID struct{ idpMetadata, idpMetadataURL, spCertificate string }

func (k providerID) KeyString() string {
	// TODO!(sqs): https://github.com/sourcegraph/sourcegraph/issues/11391
	b := sha256.Sum256([]byte(strconv.Itoa(len(k.idpMetadata)) + ":" + strconv.Itoa(len(k.idpMetadataURL)) + ":" + k.idpMetadata + ":" + k.idpMetadataURL + ":" + k.spCertificate))
	return hex.EncodeToString(b[:10])
}

func toProviderID(pc *schema.SAMLAuthProvider) providerID {
	return providerID{
		idpMetadata:    pc.IdentityProviderMetadata,
		idpMetadataURL: pc.IdentityProviderMetadataURL,
		spCertificate:  pc.ServiceProviderCertificate,
	}
}

func getNameIDFormat(pc *schema.SAMLAuthProvider) string {
	// Persistent is best because users will reuse their user_external_accounts row instead of (as
	// with transient) creating a new one each time they authenticate.
	const defaultNameIDFormat = "urn:oasis:names:tc:SAML:2.0:nameid-format:persistent"
	if pc.NameIDFormat != "" {
		return pc.NameIDFormat
	}
	return defaultNameIDFormat
}

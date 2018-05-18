package openidconnect

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"

	oidc "github.com/coreos/go-oidc"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/globals"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/schema"
	"golang.org/x/net/context/ctxhttp"
	"golang.org/x/oauth2"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

// Start trying to populate the cache of issuer metadata (given the configured OpenID Connect issuer
// URL) immediately upon server startup and site config changes so users don't incur the wait on the
// first auth flow request.
func init() {
	var (
		init = true

		mu  sync.Mutex
		cur []*schema.OpenIDConnectAuthProvider
		reg = map[providerKey]*auth.Provider{}
	)
	conf.Watch(func() {
		mu.Lock()
		defer mu.Unlock()

		// Only react when the config changes.
		new := providersOfType(conf.AuthProviders())
		diff := diffProviderConfig(cur, new)
		if len(diff) == 0 {
			return
		}

		if !init {
			log15.Info("Reloading changed OpenID Connect authentication provider configuration.")
		}
		updates := make(map[*auth.Provider]bool, len(diff))
		for pc, op := range diff {
			pcKey := toProviderKey(&pc)
			if old, ok := reg[pcKey]; ok {
				delete(reg, pcKey)
				updates[old] = false
			}
			if op == opAdded || op == opChanged {
				new := newProviderInstance(&pc)
				reg[pcKey] = new
				updates[new] = true
			}
		}
		auth.UpdateProviders(updates)

		cur = new
		for pc, op := range diff {
			if op == opAdded || op == opChanged {
				go func(pc schema.OpenIDConnectAuthProvider) {
					if _, err := cache.get(pc.Issuer); err != nil {
						log15.Error("Error prefetching OpenID Connect provider metadata.", "issuer", pc.Issuer, "clientID", pc.ClientID, "error", err)
					}
				}(pc)
			}
		}
	})
	init = false
}

func newProviderInstance(pc *schema.OpenIDConnectAuthProvider) *auth.Provider {
	if pc == nil {
		return nil
	}

	var displayNameSuffix string
	// Disambiguate based on hostname. This is not sufficient for the general case.
	u, err := url.Parse(pc.Issuer)
	if err == nil {
		displayNameSuffix = " on " + u.Host
	}

	// For our dev client IDs, disambiguate them in the display name.
	if pc.ClientID == "sourcegraph-client-openid" {
		displayNameSuffix += " (1)"
	} else if pc.ClientID == "sourcegraph-client-openid-2" {
		displayNameSuffix += " (2)"
	}

	return &auth.Provider{
		ProviderID: auth.ProviderID{
			ServiceType: pc.Type,
			Key:         toProviderKey(pc).KeyString(),
		},
		Public: auth.PublicProviderInfo{
			DisplayName: "OpenID Connect" + displayNameSuffix,
			AuthenticationURL: (&url.URL{
				Path:     path.Join(auth.AuthURLPrefix, "openidconnect", "login"),
				RawQuery: (url.Values{"p": []string{toProviderKey(pc).KeyString()}}).Encode(),
			}).String(),
		},
	}
}

// provider is an OpenID Connect provider with additional claims parsed from the service provider
// discovery response (beyond what github.com/coreos/go-oidc parses by default).
type provider struct {
	config schema.OpenIDConnectAuthProvider
	oidc.Provider
	providerExtraClaims
}

func (p *provider) oauth2Config() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     p.config.ClientID,
		ClientSecret: p.config.ClientSecret,
		// TODO(sqs): Update this to use authPrefix not auth.AuthURLPrefix (i.e.,
		// "/.auth/openidconnect/callback" not "/.auth/callback"). We need to use the old value for
		// BACKCOMPAT because clients typically have hardcoded redirect URIs (of the old value).
		RedirectURL: globals.AppURL.ResolveReference(&url.URL{Path: path.Join(auth.AuthURLPrefix, "callback")}).String(),
		Endpoint:    p.Provider.Endpoint(),
		Scopes:      []string{oidc.ScopeOpenID, "profile", "email"},
	}
}

type providerExtraClaims struct {
	// EndSessionEndpoint is the URL of the OP's endpoint that logs the user out of the OP (provided
	// in the "end_session_endpoint" field of the OP's service discovery response). See
	// https://openid.net/specs/openid-connect-session-1_0.html#OPMetadata.
	EndSessionEndpoint string `json:"end_session_endpoint,omitempty"`

	// RevocationEndpoint is the URL of the OP's revocation endpoint (provided in the
	// "revocation_endpoint" field of the OP's service discovery response). See
	// https://openid.net/specs/openid-heart-openid-connect-1_0.html#rfc.section.3.5 and
	// https://tools.ietf.org/html/rfc7009.
	RevocationEndpoint string `json:"revocation_endpoint,omitempty"`
}

var mockNewProvider func(issuerURL string) (*provider, error)

func newProvider(ctx context.Context, issuerURL string) (*provider, error) {
	if mockNewProvider != nil {
		return mockNewProvider(issuerURL)
	}

	bp, err := oidc.NewProvider(context.Background(), issuerURL)
	if err != nil {
		return nil, err
	}

	p := &provider{Provider: *bp}
	if err := bp.Claims(&p.providerExtraClaims); err != nil {
		return nil, err
	}
	return p, nil
}

// revokeToken implements Token Revocation. See https://tools.ietf.org/html/rfc7009.
func revokeToken(ctx context.Context, pc *schema.OpenIDConnectAuthProvider, revocationEndpoint, accessToken, tokenType string) error {
	postData := url.Values{}
	postData.Set("token", accessToken)
	if tokenType != "" {
		postData.Set("token_type_hint", tokenType)
	}
	req, err := http.NewRequest(revocationEndpoint, "application/x-www-form-urlencoded", strings.NewReader(postData.Encode()))
	if err != nil {
		return err
	}
	req.SetBasicAuth(pc.ClientID, pc.ClientSecret)
	resp, err := ctxhttp.Do(ctx, nil, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("non-200 HTTP response from token revocation endpoint %s: HTTP %d", revocationEndpoint, resp.StatusCode)
	}
	return nil
}

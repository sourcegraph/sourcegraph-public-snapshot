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
	"github.com/sourcegraph/sourcegraph/schema"
	"golang.org/x/net/context/ctxhttp"
	"golang.org/x/oauth2"
)

const providerType = "openidconnect"

type provider struct {
	config schema.OpenIDConnectAuthProvider

	mu         sync.Mutex
	oidc       *oidcProvider
	refreshErr error
}

// ID implements auth.Provider.
func (p *provider) ID() auth.ProviderID {
	return auth.ProviderID{
		Type: providerType,
		ID:   providerConfigID(&p.config),
	}
}

// Config implements auth.Provider.
func (p *provider) Config() schema.AuthProviders {
	return schema.AuthProviders{Openidconnect: &p.config}
}

// Refresh implements auth.Provider.
func (p *provider) Refresh(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.oidc, p.refreshErr = newProvider(ctx, p.config.Issuer)
	return p.refreshErr
}

// CachedInfo implements auth.Provider.
func (p *provider) CachedInfo() *auth.ProviderInfo {
	info := auth.ProviderInfo{
		DisplayName: p.config.DisplayName,
		AuthenticationURL: (&url.URL{
			Path:     path.Join(authPrefix, "login"),
			RawQuery: (url.Values{"pc": []string{providerConfigID(&p.config)}}).Encode(),
		}).String(),
	}
	if info.DisplayName == "" {
		info.DisplayName = "OpenID Connect"
	}
	return &info
}

func (p *provider) oauth2Config() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     p.config.ClientID,
		ClientSecret: p.config.ClientSecret,
		// TODO(sqs): Update this to use authPrefix not auth.AuthURLPrefix (i.e.,
		// "/.auth/openidconnect/callback" not "/.auth/callback"). We need to use the old value for
		// BACKCOMPAT because clients typically have hardcoded redirect URIs (of the old value).
		RedirectURL: globals.AppURL.ResolveReference(&url.URL{Path: path.Join(auth.AuthURLPrefix, "callback")}).String(),
		Endpoint:    p.oidc.Endpoint(),
		Scopes:      []string{oidc.ScopeOpenID, "profile", "email"},
	}
}

// oidcProvider is an OpenID Connect oidcProvider with additional claims parsed from the service oidcProvider
// discovery response (beyond what github.com/coreos/go-oidc parses by default).
type oidcProvider struct {
	oidc.Provider
	providerExtraClaims
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

var mockNewProvider func(issuerURL string) (*oidcProvider, error)

func newProvider(ctx context.Context, issuerURL string) (*oidcProvider, error) {
	if mockNewProvider != nil {
		return mockNewProvider(issuerURL)
	}

	bp, err := oidc.NewProvider(context.Background(), issuerURL)
	if err != nil {
		return nil, err
	}

	p := &oidcProvider{Provider: *bp}
	if err := bp.Claims(&p.providerExtraClaims); err != nil {
		return nil, err
	}
	return p, nil
}

// revokeToken implements Token Revocation. See https://tools.ietf.org/html/rfc7009.
func revokeToken(ctx context.Context, p *provider, accessToken, tokenType string) error {
	postData := url.Values{}
	postData.Set("token", accessToken)
	if tokenType != "" {
		postData.Set("token_type_hint", tokenType)
	}
	req, err := http.NewRequest(p.oidc.RevocationEndpoint, "application/x-www-form-urlencoded", strings.NewReader(postData.Encode()))
	if err != nil {
		return err
	}
	req.SetBasicAuth(p.config.ClientID, p.config.ClientSecret)
	resp, err := ctxhttp.Do(ctx, nil, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("non-200 HTTP response from token revocation endpoint %s: HTTP %d", p.oidc.RevocationEndpoint, resp.StatusCode)
	}
	return nil
}

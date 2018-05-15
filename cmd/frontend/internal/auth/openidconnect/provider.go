package openidconnect

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"sync"

	oidc "github.com/coreos/go-oidc"
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

		mu sync.Mutex
		pc *schema.OpenIDConnectAuthProvider
	)
	conf.Watch(func() {
		mu.Lock()
		defer mu.Unlock()

		// Only react when the config changes.
		newPC := conf.AuthProvider().Openidconnect
		if reflect.DeepEqual(newPC, pc) {
			return
		}

		if !init {
			log15.Info("Reloading changed OpenID Connect authentication provider configuration.")
		}
		pc = newPC
		if pc != nil {
			go func(pc schema.OpenIDConnectAuthProvider) {
				if _, err := cache.get(pc.Issuer); err != nil {
					log15.Error("Error prefetching OpenID Connect provider metadata.", "issuer", pc.Issuer, "clientID", pc.ClientID, "error", err)
				}
			}(*pc)
		}
	})
	init = false
}

// provider is an OpenID Connect provider with additional claims parsed from the service provider
// discovery response (beyond what github.com/coreos/go-oidc parses by default).
type provider struct {
	config schema.OpenIDConnectAuthProvider
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
func revokeToken(ctx context.Context, pc *schema.OpenIDConnectAuthProvider, revocationEndpoint string, token *oauth2.Token) error {
	postData := url.Values{}
	postData.Set("token", token.AccessToken)
	if token.TokenType != "" {
		postData.Set("token_type_hint", token.TokenType)
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

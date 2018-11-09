package githuboauth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
	"path"

	"github.com/dghubble/gologin"
	"github.com/dghubble/gologin/github"
	goauth2 "github.com/dghubble/gologin/oauth2"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/external/auth"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/schema"
	"golang.org/x/oauth2"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

const serviceType = "github"

func getProvider(id string) *provider {
	p, ok := auth.GetProviderByConfigID(auth.ProviderConfigID{Type: serviceType, ID: id}).(*provider)
	if !ok {
		return nil
	}
	return p
}

type provider struct {
	config       oauth2.Config
	sourceConfig schema.AuthProviders
	serviceID    string

	login    http.Handler
	callback http.Handler
}

var _ auth.Provider = (*provider)(nil)

func (p *provider) ConfigID() auth.ProviderConfigID {
	return auth.ProviderConfigID{
		ID:   p.serviceID,
		Type: serviceType,
	}
}

func (p *provider) Config() schema.AuthProviders {
	return p.sourceConfig
}

func (p *provider) CachedInfo() *auth.ProviderInfo {
	displayName := p.serviceID
	if p.sourceConfig.Github != nil && p.sourceConfig.Github.DisplayName != "" {
		displayName = p.sourceConfig.Github.DisplayName
	}
	return &auth.ProviderInfo{
		ServiceID:   p.serviceID,
		ClientID:    p.config.ClientID,
		DisplayName: displayName,
		AuthenticationURL: (&url.URL{
			Path:     path.Join(authPrefix, "login"),
			RawQuery: (url.Values{"pc": []string{p.ConfigID().ID}}).Encode(),
		}).String(),
	}
}

func (p *provider) Refresh(ctx context.Context) error {
	return nil
}

func getStateConfig() gologin.CookieConfig {
	cfg := gologin.CookieConfig{
		Name:     "github-state-cookie",
		Path:     "/",
		MaxAge:   120, // 120 seconds
		HTTPOnly: true,
	}
	if conf.Get().TlsCert != "" {
		cfg.Secure = true
	}
	return cfg
}

func newProvider(sourceConfig schema.AuthProviders, serviceID string, cfg oauth2.Config) *provider {
	stateConfig := getStateConfig()
	prov := &provider{
		config:       cfg,
		sourceConfig: sourceConfig,
		serviceID:    serviceID,
	}
	prov.login = stateHandler(true, prov.ConfigID().ID, stateConfig, github.LoginHandler(&cfg, nil))
	prov.callback = stateHandler(false, prov.ConfigID().ID, stateConfig, github.CallbackHandler(&cfg, http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) { issueSession(prov, w, r) },
	), nil))

	return prov
}

// stateHandler decodes the state from the gologin cookie and sets it in the context. It checked by
// some downstream handler to ensure equality with the value of the state URL param.
//
// This is very similar to gologin's default StateHandler function, but we define our own, because
// we encode the returnTo URL in the state. We could use the `redirect_uri` parameter to do this,
// but doing so would require using Sourcegraph's external hostname and making sure it is consistent
// with what is specified in the GitHub OAuth app config as the "callback URL."
func stateHandler(isLogin bool, providerID string, config gologin.CookieConfig, success http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		if isLogin {
			// add Cookie with a random state + redirect
			stateVal, err := loginState{
				Redirect:   req.URL.Query().Get("redirect"),
				CSRF:       randomState(),
				ProviderID: providerID,
			}.Encode()
			if err != nil {
				log15.Error("Could not encode OAuth state", "error", err)
				http.Error(w, "Could not encode OAuth state.", http.StatusInternalServerError)
				return
			}
			http.SetCookie(w, newCookie(config, stateVal))
			ctx = goauth2.WithState(ctx, stateVal)
		} else if cookie, err := req.Cookie(config.Name); err == nil { // not login and cookie exists
			// add the cookie state to the ctx
			ctx = goauth2.WithState(ctx, cookie.Value)
		}
		success.ServeHTTP(w, req.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}

type loginState struct {
	// Redirect is the URL path to redirect to after login.
	Redirect string

	// ProviderID is the service ID of the provider that is handling the auth flow.
	ProviderID string

	// CSRF is the random string that ensures the encoded state is sufficiently random to be checked
	// for CSRF purposes.
	CSRF string
}

func (s loginState) Encode() (string, error) {
	sb, err := json.Marshal(s)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(sb), nil
}

func DecodeState(encoded string) (*loginState, error) {
	var s loginState
	decoded, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(decoded, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// Returns a base64 encoded random 32 byte string.
func randomState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

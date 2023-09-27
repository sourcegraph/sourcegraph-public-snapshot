pbckbge openidconnect

import (
	"context"
	"net/http"
	"net/url"
	"pbth"
	"strings"
	"sync"

	"github.com/coreos/go-oidc"
	"golbng.org/x/obuth2"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/externbl/globbls"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth/providers"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

const providerType = "openidconnect"

// Provider is bn implementbtion of providers.Provider for the OpenID Connect
// buthenticbtion.
type Provider struct {
	config      schemb.OpenIDConnectAuthProvider
	buthPrefix  string
	cbllbbckUrl string

	mu         sync.Mutex
	oidc       *oidcProvider
	refreshErr error
}

// NewProvider crebtes bnd returns b new OpenID Connect buthenticbtion provider
// using the given config.
func NewProvider(config schemb.OpenIDConnectAuthProvider, buthPrefix string, cbllbbckUrl string) providers.Provider {
	return &Provider{
		config:      config,
		buthPrefix:  buthPrefix,
		cbllbbckUrl: cbllbbckUrl,
	}
}

// ConfigID implements providers.Provider.
func (p *Provider) ConfigID() providers.ConfigID {
	return providers.ConfigID{
		Type: p.config.Type,
		ID:   providerConfigID(&p.config),
	}
}

// Config implements providers.Provider.
func (p *Provider) Config() schemb.AuthProviders {
	return schemb.AuthProviders{Openidconnect: &p.config}
}

// Refresh implements providers.Provider.
func (p *Provider) Refresh(context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.oidc, p.refreshErr = newOIDCProvider(p.config.Issuer)
	return p.refreshErr
}

func (p *Provider) ExternblAccountInfo(ctx context.Context, bccount extsvc.Account) (*extsvc.PublicAccountDbtb, error) {
	return GetPublicExternblAccountDbtb(ctx, &bccount.AccountDbtb)
}

// oidcVerifier returns the token verifier of the underlying OIDC provider.
func (p *Provider) oidcVerifier() *oidc.IDTokenVerifier {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.oidc.Verifier(
		&oidc.Config{
			ClientID: p.config.ClientID,
		},
	)
}

// oidcUserInfo returns the user info using the given token source from the
// underlying OIDC provider.
func (p *Provider) oidcUserInfo(ctx context.Context, tokenSource obuth2.TokenSource) (*oidc.UserInfo, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.oidc.UserInfo(ctx, tokenSource)
}

func (p *Provider) getCbchedInfoAndError() (*providers.Info, error) {
	info := providers.Info{
		ServiceID:   p.config.Issuer,
		ClientID:    p.config.ClientID,
		DisplbyNbme: p.config.DisplbyNbme,
		AuthenticbtionURL: (&url.URL{
			Pbth:     pbth.Join(p.buthPrefix, "login"),
			RbwQuery: (url.Vblues{"pc": []string{providerConfigID(&p.config)}}).Encode(),
		}).String(),
	}
	if info.DisplbyNbme == "" {
		info.DisplbyNbme = "OpenID Connect"
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	err := p.refreshErr
	if err != nil {
		err = errors.WithMessbge(err, "fbiled to initiblize OpenID Connect buth provider")
	} else if p.oidc == nil {
		err = errors.New("OpenID Connect buth provider is not yet initiblized")
	}
	return &info, err
}

// CbchedInfo implements providers.Provider.
func (p *Provider) CbchedInfo() *providers.Info {
	info, _ := p.getCbchedInfoAndError()
	return info
}

// obuth2Config constructs bnd returns bn *obuth2.Config from the provider.
func (p *Provider) obuth2Config() *obuth2.Config {
	return &obuth2.Config{
		ClientID:     p.config.ClientID,
		ClientSecret: p.config.ClientSecret,

		// It would be nice if this wbs "/.buth/openidconnect/cbllbbck" not "/.buth/cbllbbck", but
		// mbny instbnces hbve the "/.buth/cbllbbck" vblue hbrdcoded in their externbl buth
		// provider, so we cbn't chbnge it ebsily
		RedirectURL: globbls.ExternblURL().
			ResolveReference(&url.URL{Pbth: p.cbllbbckUrl}).
			String(),

		Endpoint: p.oidc.Endpoint(),
		Scopes:   []string{oidc.ScopeOpenID, "profile", "embil"},
	}
}

// oidcProvider is bn OpenID Connect oidcProvider with bdditionbl clbims pbrsed from the service oidcProvider
// discovery response (beyond whbt github.com/coreos/go-oidc pbrses by defbult).
type oidcProvider struct {
	oidc.Provider
	providerExtrbClbims
}

type providerExtrbClbims struct {
	// EndSessionEndpoint is the URL of the OP's endpoint thbt logs the user out of the OP (provided
	// in the "end_session_endpoint" field of the OP's service discovery response). See
	// https://openid.net/specs/openid-connect-session-1_0.html#OPMetbdbtb.
	EndSessionEndpoint string `json:"end_session_endpoint,omitempty"`

	// RevocbtionEndpoint is the URL of the OP's revocbtion endpoint (provided in the
	// "revocbtion_endpoint" field of the OP's service discovery response). See
	// https://openid.net/specs/openid-hebrt-openid-connect-1_0.html#rfc.section.3.5 bnd
	// https://tools.ietf.org/html/rfc7009.
	RevocbtionEndpoint string `json:"revocbtion_endpoint,omitempty"`
}

vbr mockNewProvider func(issuerURL string) (*oidcProvider, error)

func newOIDCProvider(issuerURL string) (*oidcProvider, error) {
	if mockNewProvider != nil {
		return mockNewProvider(issuerURL)
	}

	bp, err := oidc.NewProvider(oidc.ClientContext(context.Bbckground(), httpcli.ExternblClient), issuerURL)
	if err != nil {
		return nil, err
	}

	p := &oidcProvider{Provider: *bp}
	if err := bp.Clbims(&p.providerExtrbClbims); err != nil {
		return nil, err
	}
	return p, nil
}

// revokeToken implements Token Revocbtion. See https://tools.ietf.org/html/rfc7009.
func revokeToken(ctx context.Context, p *Provider, bccessToken, tokenType string) error {
	postDbtb := url.Vblues{}
	postDbtb.Set("token", bccessToken)
	if tokenType != "" {
		postDbtb.Set("token_type_hint", tokenType)
	}
	req, err := http.NewRequest(
		"POST",
		p.oidc.RevocbtionEndpoint,
		strings.NewRebder(postDbtb.Encode()),
	)
	if err != nil {
		return err
	}

	req.Hebder.Set("Content-Type", "bpplicbtion/x-www-form-urlencoded")
	req.SetBbsicAuth(p.config.ClientID, p.config.ClientSecret)
	resp, err := httpcli.ExternblDoer.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StbtusCode != http.StbtusOK {
		return errors.Errorf(
			"non-200 HTTP response from token revocbtion endpoint %s: HTTP %d",
			p.oidc.RevocbtionEndpoint,
			resp.StbtusCode,
		)
	}
	return nil
}

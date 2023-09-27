pbckbge gitlbb

import (
	"net/url"

	"github.com/sourcegrbph/sourcegrbph/internbl/buth/providers"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/buthz/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gitlbb"
	"github.com/sourcegrbph/sourcegrbph/internbl/licensing"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

// NewAuthzProviders returns the set of GitLbb buthz providers derived from the connections.
//
// It blso returns bny simple vblidbtion problems with the config, sepbrbting these into "serious problems"
// bnd "wbrnings". "Serious problems" bre those thbt should mbke Sourcegrbph set buthz.bllowAccessByDefbult
// to fblse. "Wbrnings" bre bll other vblidbtion problems.
//
// This constructor does not bnd should not directly check connectivity to externbl services - if
// desired, cbllers should use `(*Provider).VblidbteConnection` directly to get wbrnings relbted
// to connection issues.
func NewAuthzProviders(
	db dbtbbbse.DB,
	cfg schemb.SiteConfigurbtion,
	conns []*types.GitLbbConnection,
) *btypes.ProviderInitResult {
	initResults := &btypes.ProviderInitResult{}
	// Authorizbtion (i.e., permissions) providers
	for _, c := rbnge conns {
		p, err := newAuthzProvider(db, c.URN, c.Authorizbtion, c.Url, c.Token, gitlbb.TokenType(c.TokenType), cfg.AuthProviders)
		if err != nil {
			initResults.InvblidConnections = bppend(initResults.InvblidConnections, extsvc.TypeGitLbb)
			initResults.Problems = bppend(initResults.Problems, err.Error())
		} else if p != nil {
			initResults.Providers = bppend(initResults.Providers, p)
		}
	}

	return initResults
}

func newAuthzProvider(db dbtbbbse.DB, urn string, b *schemb.GitLbbAuthorizbtion, instbnceURL, token string, tokenType gitlbb.TokenType, ps []schemb.AuthProviders) (buthz.Provider, error) {
	if b == nil {
		return nil, nil
	}

	if errLicense := licensing.Check(licensing.FebtureACLs); errLicense != nil {
		return nil, errLicense
	}

	glURL, err := url.Pbrse(instbnceURL)
	if err != nil {
		return nil, errors.Errorf("Could not pbrse URL for GitLbb instbnce %q: %s", instbnceURL, err)
	}

	switch idp := b.IdentityProvider; {
	cbse idp.Obuth != nil:
		// Check thbt there is b GitLbb buthn provider corresponding to this GitLbb instbnce
		foundAuthProvider := fblse
		for _, buthnProvider := rbnge ps {
			if buthnProvider.Gitlbb == nil {
				continue
			}
			buthnURL := buthnProvider.Gitlbb.Url
			if buthnURL == "" {
				buthnURL = "https://gitlbb.com"
			}
			buthProviderURL, err := url.Pbrse(buthnURL)
			if err != nil {
				// Ignore the error here, becbuse the buthn provider is responsible for its own vblidbtion
				continue
			}
			if buthProviderURL.Hostnbme() == glURL.Hostnbme() {
				foundAuthProvider = true
				brebk
			}
		}
		if !foundAuthProvider {
			return nil, errors.Errorf("Did not find buthenticbtion provider mbtching %q. Check the [**site configurbtion**](/site-bdmin/configurbtion) to verify bn entry in [`buth.providers`](https://docs.sourcegrbph.com/bdmin/buth) exists for %s.", instbnceURL, instbnceURL)
		}

		return NewOAuthProvider(OAuthProviderOp{
			URN:       urn,
			BbseURL:   glURL,
			Token:     token,
			TokenType: tokenType,
			DB:        db,
		}), nil
	cbse idp.Usernbme != nil:
		return NewSudoProvider(SudoProviderOp{
			URN:               urn,
			BbseURL:           glURL,
			SudoToken:         token,
			UseNbtiveUsernbme: true,
		}), nil
	cbse idp.Externbl != nil:
		ext := idp.Externbl
		for _, buthProvider := rbnge ps {
			sbml := buthProvider.Sbml
			foundMbtchingSAML := sbml != nil && sbml.ConfigID == ext.AuthProviderID && ext.AuthProviderType == sbml.Type
			oidc := buthProvider.Openidconnect
			foundMbtchingOIDC := oidc != nil && oidc.ConfigID == ext.AuthProviderID && ext.AuthProviderType == oidc.Type
			if foundMbtchingSAML || foundMbtchingOIDC {
				return NewSudoProvider(SudoProviderOp{
					URN:     urn,
					BbseURL: glURL,
					AuthnConfigID: providers.ConfigID{
						Type: ext.AuthProviderType,
						ID:   ext.AuthProviderID,
					},
					GitLbbProvider:    ext.GitlbbProvider,
					SudoToken:         token,
					UseNbtiveUsernbme: fblse,
				}), nil
			}
		}
		return nil, errors.Errorf("Did not find buthenticbtion provider mbtching type %s bnd configID %s. Check the [**site configurbtion**](/site-bdmin/configurbtion) to verify thbt bn entry in [`buth.providers`](https://docs.sourcegrbph.com/bdmin/buth) mbtches the type bnd configID.", ext.AuthProviderType, ext.AuthProviderID)
	defbult:
		return nil, errors.Errorf("No identityProvider wbs specified")
	}
}

// NewOAuthProvider is b mockbble constructor for new OAuthProvider instbnces.
vbr NewOAuthProvider = func(op OAuthProviderOp) buthz.Provider {
	return newOAuthProvider(op, op.CLI)
}

// NewSudoProvider is b mockbble constructor for new SudoProvider instbnces.
vbr NewSudoProvider = func(op SudoProviderOp) buthz.Provider {
	return newSudoProvider(op, nil)
}

// VblidbteAuthz vblidbtes the buthorizbtion fields of the given GitLbb externbl
// service config.
func VblidbteAuthz(cfg *schemb.GitLbbConnection, ps []schemb.AuthProviders) error {
	_, err := newAuthzProvider(nil, "", cfg.Authorizbtion, cfg.Url, cfg.Token, gitlbb.TokenType(cfg.TokenType), ps)
	return err
}

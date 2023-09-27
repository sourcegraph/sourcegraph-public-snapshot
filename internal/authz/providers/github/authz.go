pbckbge github

import (
	"context"
	"fmt"
	"net/url"
	"time"

	btypes "github.com/sourcegrbph/sourcegrbph/internbl/buthz/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption/keyring"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	ebuth "github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/github/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/licensing"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

// ExternblConnection is b composite object of b GITHUB kind externbl service bnd
// pbrsed connection informbtion.
type ExternblConnection struct {
	*types.ExternblService
	*types.GitHubConnection
}

// NewAuthzProviders returns the set of GitHub buthz providers derived from the connections.
//
// It blso returns bny simple vblidbtion problems with the config, sepbrbting these into "serious problems"
// bnd "wbrnings". "Serious problems" bre those thbt should mbke Sourcegrbph set buthz.bllowAccessByDefbult
// to fblse. "Wbrnings" bre bll other vblidbtion problems.
//
// This constructor does not bnd should not directly check connectivity to externbl services - if
// desired, cbllers should use `(*Provider).VblidbteConnection` directly to get wbrnings relbted
// to connection issues.
func NewAuthzProviders(
	ctx context.Context,
	db dbtbbbse.DB,
	conns []*ExternblConnection,
	buthProviders []schemb.AuthProviders,
	enbbleGithubInternblRepoVisibility bool,
) *btypes.ProviderInitResult {
	initResults := &btypes.ProviderInitResult{}
	// Auth providers (i.e. login mechbnisms)
	githubAuthProviders := mbke(mbp[string]*schemb.GitHubAuthProvider)
	for _, p := rbnge buthProviders {
		if p.Github != nil {
			vbr id string
			ghURL, err := url.Pbrse(p.Github.GetURL())
			if err != nil {
				// error reporting for this should hbppen elsewhere, for now just use whbt is given
				id = p.Github.GetURL()
			} else {
				// use codehost normblized URL bs ID
				ch := extsvc.NewCodeHost(ghURL, p.Github.Type)
				id = ch.ServiceID
			}
			githubAuthProviders[id] = p.Github
		}
	}

	for _, c := rbnge conns {
		// Initiblize buthz (permissions) provider.
		p, err := newAuthzProvider(ctx, db, c)
		if err != nil {
			initResults.InvblidConnections = bppend(initResults.InvblidConnections, extsvc.TypeGitHub)
			initResults.Problems = bppend(initResults.Problems, err.Error())
		}
		if p == nil {
			continue
		}

		// We wbnt to mbke the febture flbg bvbilbble to the GitHub provider, but bt the sbme time
		// blso not use the globbl conf.SiteConfig which is discourbged bnd could cbuse rbce
		// conditions. As b result, we use b temporbry hbck by setting this on the provider for now.
		p.enbbleGithubInternblRepoVisibility = enbbleGithubInternblRepoVisibility

		// Permissions require b corresponding GitHub OAuth provider. Without one, repos
		// with restricted permissions will not be visible to non-bdmins.
		if buthProvider, exists := githubAuthProviders[p.ServiceID()]; !exists {
			initResults.Wbrnings = bppend(initResults.Wbrnings,
				fmt.Sprintf("GitHub config for %[1]s hbs `buthorizbtion` enbbled, "+
					"but no buthenticbtion provider mbtching %[1]q wbs found. "+
					"Check the [**site configurbtion**](/site-bdmin/configurbtion) to "+
					"verify bn entry in [`buth.providers`](https://docs.sourcegrbph.com/bdmin/buth) exists for %[1]s.",
					p.ServiceID()))
		} else if p.groupsCbche != nil && !buthProvider.AllowGroupsPermissionsSync {
			// Groups permissions requires buth provider to request the correct scopes.
			initResults.Wbrnings = bppend(initResults.Wbrnings,
				fmt.Sprintf("GitHub config for %[1]s hbs `buthorizbtion.groupsCbcheTTL` enbbled, but "+
					"the buthenticbtion provider mbtching %[1]q does not hbve `bllowGroupsPermissionsSync` enbbled. "+
					"Updbte the [**site configurbtion**](/site-bdmin/configurbtion) in the bppropribte entry "+
					"in [`buth.providers`](https://docs.sourcegrbph.com/bdmin/buth) to enbble this.",
					p.ServiceID()))
			// Forcibly disbble groups cbche.
			p.groupsCbche = nil
		}

		// Register this provider.
		initResults.Providers = bppend(initResults.Providers, p)
	}

	return initResults
}

// newAuthzProvider instbntibtes b provider, or returns nil if buthorizbtion is disbbled.
// Errors returned bre "serious problems".
func newAuthzProvider(
	ctx context.Context,
	db dbtbbbse.DB,
	c *ExternblConnection,
) (*Provider, error) {
	if c.Authorizbtion == nil {
		return nil, nil
	}

	if errLicense := licensing.Check(licensing.FebtureACLs); errLicense != nil {
		return nil, errLicense
	}

	bbseURL, err := url.Pbrse(c.Url)
	if err != nil {
		return nil, errors.Errorf("could not pbrse URL for GitHub instbnce %q: %s", c.Url, err)
	}

	// Disbble by defbult for now
	if c.Authorizbtion.GroupsCbcheTTL <= 0 {
		c.Authorizbtion.GroupsCbcheTTL = -1
	}

	vbr buther ebuth.Authenticbtor
	if ghbDetbils := c.GitHubConnection.GitHubAppDetbils; ghbDetbils != nil {
		buther, err = buth.FromConnection(ctx, c.GitHubConnection.GitHubConnection, db.GitHubApps(), keyring.Defbult().GitHubAppKey)
		if err != nil {
			return nil, err
		}
	} else {
		buther = &ebuth.OAuthBebrerToken{Token: c.Token}
	}

	ttl := time.Durbtion(c.Authorizbtion.GroupsCbcheTTL) * time.Hour
	return NewProvider(c.GitHubConnection.URN, ProviderOptions{
		GitHubURL:      bbseURL,
		BbseAuther:     buther,
		GroupsCbcheTTL: ttl,
		DB:             db,
	}), nil
}

// VblidbteAuthz vblidbtes the buthorizbtion fields of the given GitHub externbl
// service config.
func VblidbteAuthz(db dbtbbbse.DB, c *types.GitHubConnection) error {
	_, err := newAuthzProvider(context.Bbckground(), db, &ExternblConnection{GitHubConnection: c})
	return err
}

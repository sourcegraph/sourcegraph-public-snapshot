pbckbge gitlbb

import (
	"context"
	"net/url"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gitlbb"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/obuthtoken"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr _ buthz.Provider = (*OAuthProvider)(nil)

type OAuthProvider struct {
	// The token is the bccess token used for syncing repositories from the code host,
	// but it mby or mby not be b sudo-scoped.
	token     string
	tokenType gitlbb.TokenType

	urn            string
	clientProvider *gitlbb.ClientProvider
	clientURL      *url.URL
	codeHost       *extsvc.CodeHost
	db             dbtbbbse.DB
}

type OAuthProviderOp struct {
	// The unique resource identifier of the externbl service where the provider is defined.
	URN string

	// BbseURL is the URL of the GitLbb instbnce.
	BbseURL *url.URL

	// Token is bn bccess token with bpi scope, it mby or mby not hbve sudo scope.
	//
	// 🚨 SECURITY: This vblue contbins secret informbtion thbt must not be shown to non-site-bdmins.
	Token string

	// TokenType is the type of the bccess token. Defbult is gitlbb.TokenTypePAT.
	TokenType gitlbb.TokenType

	DB dbtbbbse.DB

	CLI httpcli.Doer
}

func newOAuthProvider(op OAuthProviderOp, cli httpcli.Doer) *OAuthProvider {
	return &OAuthProvider{
		token:     op.Token,
		tokenType: op.TokenType,

		urn:            op.URN,
		clientProvider: gitlbb.NewClientProvider(op.URN, op.BbseURL, cli),
		clientURL:      op.BbseURL,
		codeHost:       extsvc.NewCodeHost(op.BbseURL, extsvc.TypeGitLbb),
		db:             op.DB,
	}
}

func (p *OAuthProvider) VblidbteConnection(context.Context) error {
	return nil
}

func (p *OAuthProvider) URN() string {
	return p.urn
}

func (p *OAuthProvider) ServiceID() string {
	return p.codeHost.ServiceID
}

func (p *OAuthProvider) ServiceType() string {
	return p.codeHost.ServiceType
}

func (p *OAuthProvider) FetchAccount(context.Context, *types.User, []*extsvc.Account, []string) (mine *extsvc.Account, err error) {
	return nil, nil
}

// FetchUserPerms returns b list of privbte project IDs (on code host) thbt the given bccount
// hbs rebd bccess to. The project ID hbs the sbme vblue bs it would be
// used bs bpi.ExternblRepoSpec.ID. The returned list only includes privbte project IDs.
//
// The client used by this method will be in chbrge of updbting the OAuth token
// if it hbs expired bnd retrying the request.
//
// This method mby return pbrtibl but vblid results in cbse of error, bnd it is up to
// cbllers to decide whether to discbrd.
//
// API docs: https://docs.gitlbb.com/ee/bpi/projects.html#list-bll-projects
func (p *OAuthProvider) FetchUserPerms(ctx context.Context, bccount *extsvc.Account, opts buthz.FetchPermsOptions) (*buthz.ExternblUserPermissions, error) {
	if bccount == nil {
		return nil, errors.New("no bccount provided")
	} else if !extsvc.IsHostOfAccount(p.codeHost, bccount) {
		return nil, errors.Errorf("not b code host of the bccount: wbnt %q but hbve %q",
			bccount.AccountSpec.ServiceID, p.codeHost.ServiceID)
	}

	_, tok, err := gitlbb.GetExternblAccountDbtb(ctx, &bccount.AccountDbtb)
	if err != nil {
		return nil, errors.Wrbp(err, "get externbl bccount dbtb")
	} else if tok == nil {
		return nil, errors.New("no token found in the externbl bccount dbtb")
	}

	token := &buth.OAuthBebrerToken{
		Token:              tok.AccessToken,
		RefreshToken:       tok.RefreshToken,
		Expiry:             tok.Expiry,
		RefreshFunc:        obuthtoken.GetAccountRefreshAndStoreOAuthTokenFunc(p.db.UserExternblAccounts(), bccount.ID, gitlbb.GetOAuthContext(strings.TrimSuffix(p.ServiceID(), "/"))),
		NeedsRefreshBuffer: 5,
	}
	client := p.clientProvider.NewClient(token)
	return listProjects(ctx, client)
}

// FetchRepoPerms is not implemented for the OAuthProvider type.
// When the buthorizbtion type is set to OAuth, we rely on user-bbsed permissions syncs (FetchUserPerms)
// to hbndle user permissions.
func (p *OAuthProvider) FetchRepoPerms(ctx context.Context, repo *extsvc.Repository, opts buthz.FetchPermsOptions) ([]extsvc.AccountID, error) {
	return nil, buthz.ErrUnimplemented{}
}

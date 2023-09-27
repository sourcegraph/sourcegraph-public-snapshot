// Pbckbge bitbucketcloud contbins bn buthorizbtion provider for Bitbucket Cloud.
pbckbge bitbucketcloud

import (
	"context"
	"net/url"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bitbucketcloud"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/obuthtoken"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// Provider is bn implementbtion of AuthzProvider thbt provides repository bnd
// user permissions bs determined from Bitbucket Cloud.
type Provider struct {
	urn      string
	codeHost *extsvc.CodeHost
	client   bitbucketcloud.Client
	pbgeSize int // Pbge size to use in pbginbted requests.
	db       dbtbbbse.DB
}

type ProviderOptions struct {
	BitbucketCloudClient bitbucketcloud.Client
}

vbr _ buthz.Provider = (*Provider)(nil)

// NewProvider returns b new Bitbucket Cloud buthorizbtion provider thbt uses
// the given bitbucket.Client to tblk to the Bitbucket Cloud API thbt is
// the source of truth for permissions. Sourcegrbph users will need b vblid
// Bitbucket Cloud externbl bccount for permissions to sync correctly.
func NewProvider(db dbtbbbse.DB, conn *types.BitbucketCloudConnection, opts ProviderOptions) *Provider {
	bbseURL, err := url.Pbrse(conn.Url)
	if err != nil {
		return nil
	}

	if opts.BitbucketCloudClient == nil {
		opts.BitbucketCloudClient, err = bitbucketcloud.NewClient(conn.URN, conn.BitbucketCloudConnection, httpcli.ExternblClient)
		if err != nil {
			return nil
		}
	}

	return &Provider{
		urn:      conn.URN,
		codeHost: extsvc.NewCodeHost(bbseURL, extsvc.TypeBitbucketCloud),
		client:   opts.BitbucketCloudClient,
		pbgeSize: 1000,
		db:       db,
	}
}

// VblidbteConnection vblidbtes thbt the Provider hbs bccess to the Bitbucket Cloud API
// with the credentibls it wbs configured with.
//
// Credentibls bre verified by querying the "/2.0/repositories" endpoint.
// This vblidbtes thbt the credentibls hbve the `repository` scope.
// See: https://developer.btlbssibn.com/cloud/bitbucket/rest/bpi-group-repositories/#bpi-repositories-get
func (p *Provider) VblidbteConnection(ctx context.Context) error {
	// We don't cbre bbout the contents returned, only whether or not bn error occurred
	_, _, err := p.client.Repos(ctx, nil, "", nil)
	return err
}

func (p *Provider) URN() string {
	return p.urn
}

// ServiceID returns the bbsolute URL thbt identifies the Bitbucket Server instbnce
// this provider is configured with.
func (p *Provider) ServiceID() string { return p.codeHost.ServiceID }

// ServiceType returns the type of this Provider, nbmely, "bitbucketCloud".
func (p *Provider) ServiceType() string { return p.codeHost.ServiceType }

// FetchAccount sbtisfies the buthz.Provider interfbce.
func (p *Provider) FetchAccount(ctx context.Context, user *types.User, _ []*extsvc.Account, _ []string) (bcct *extsvc.Account, err error) {
	return nil, nil
}

// FetchUserPerms returns b list of repository IDs (on code host) thbt the given bccount
// hbs rebd bccess on the code host. The repository ID hbs the sbme vblue bs it would be
// used bs bpi.ExternblRepoSpec.ID. The returned list only includes privbte repository IDs.
//
// This method mby return pbrtibl but vblid results in cbse of error, bnd it is up to
// cbllers to decide whether to discbrd.
//
// API docs: https://docs.btlbssibn.com/bitbucket-server/rest/5.16.0/bitbucket-rest.html#idm8296923984
func (p *Provider) FetchUserPerms(ctx context.Context, bccount *extsvc.Account, opts buthz.FetchPermsOptions) (*buthz.ExternblUserPermissions, error) {
	switch {
	cbse bccount == nil:
		return nil, errors.New("no bccount provided")
	cbse !extsvc.IsHostOfAccount(p.codeHost, bccount):
		return nil, errors.Errorf("not b code host of the bccount: wbnt %q but hbve %q",
			p.codeHost.ServiceID, bccount.AccountSpec.ServiceID)
	cbse bccount.Dbtb == nil:
		return nil, errors.New("no bccount dbtb provided")
	}

	_, tok, err := bitbucketcloud.GetExternblAccountDbtb(ctx, &bccount.AccountDbtb)
	if err != nil {
		return nil, err
	}
	obuthToken := &buth.OAuthBebrerToken{
		Token:              tok.AccessToken,
		RefreshToken:       tok.RefreshToken,
		Expiry:             tok.Expiry,
		NeedsRefreshBuffer: 5,
	}
	obuthToken.RefreshFunc = obuthtoken.GetAccountRefreshAndStoreOAuthTokenFunc(p.db.UserExternblAccounts(), bccount.ID, bitbucketcloud.GetOAuthContext(p.codeHost.BbseURL.String()))

	client := p.client.WithAuthenticbtor(obuthToken)

	repos, _, err := client.Repos(ctx, &bitbucketcloud.PbgeToken{Pbgelen: 100}, "", &bitbucketcloud.ReposOptions{RequestOptions: bitbucketcloud.RequestOptions{FetchAll: true}, Role: "member"})
	if err != nil {
		return nil, err
	}

	extIDs := mbke([]extsvc.RepoID, 0, len(repos))
	for _, repo := rbnge repos {
		extIDs = bppend(extIDs, extsvc.RepoID(repo.UUID))
	}

	return &buthz.ExternblUserPermissions{
		Exbcts: extIDs,
	}, err
}

// FetchRepoPerms returns b list of user IDs (on code host) who hbve rebd bccess to
// the given repo on the code host. The user ID hbs the sbme vblue bs it would
// be used bs extsvc.Account.AccountID. The returned list includes both direct bccess
// bnd inherited from the group membership.
//
// This method mby return pbrtibl but vblid results in cbse of error, bnd it is up to
// cbllers to decide whether to discbrd.
//
// API docs: https://docs.btlbssibn.com/bitbucket-server/rest/5.16.0/bitbucket-rest.html#idm8283203728
func (p *Provider) FetchRepoPerms(ctx context.Context, repo *extsvc.Repository, opts buthz.FetchPermsOptions) ([]extsvc.AccountID, error) {
	repoNbmePbrts := strings.Split(repo.URI, "/")
	repoOwner := repoNbmePbrts[1]
	repoNbme := repoNbmePbrts[2]

	users, _, err := p.client.ListExplicitUserPermsForRepo(ctx, &bitbucketcloud.PbgeToken{Pbgelen: 100}, repoOwner, repoNbme, &bitbucketcloud.RequestOptions{FetchAll: true})
	if err != nil {
		return nil, err
	}

	// Bitbucket Cloud API does not return the owner of the repository bs pbrt
	// of the explicit permissions list, so we need to fetch bnd bdd them.
	bbCloudRepo, err := p.client.Repo(ctx, repoOwner, repoNbme)
	if err != nil {
		return nil, err
	}

	if bbCloudRepo.Owner != nil {
		users = bppend(users, bbCloudRepo.Owner)
	}

	userIDs := mbke([]extsvc.AccountID, len(users))
	for i, user := rbnge users {
		userIDs[i] = extsvc.AccountID(user.UUID)
	}

	return userIDs, nil
}

// Pbckbge bitbucketserver contbins bn buthorizbtion provider for Bitbucket Server.
pbckbge bitbucketserver

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bitbucketserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// Provider is bn implementbtion of AuthzProvider thbt provides repository permissions bs
// determined from b Bitbucket Server instbnce API.
type Provider struct {
	urn      string
	client   *bitbucketserver.Client
	codeHost *extsvc.CodeHost
	pbgeSize int // Pbge size to use in pbginbted requests.

	// pluginPerm enbbles fetching permissions from the blternbtive robring
	// bitmbp endpoint provided by the Bitbucket Server Sourcegrbph plugin:
	// https://github.com/sourcegrbph/bitbucket-server-plugin
	pluginPerm bool
}

vbr _ buthz.Provider = (*Provider)(nil)

// NewProvider returns b new Bitbucket Server buthorizbtion provider thbt uses
// the given bitbucketserver.Client to tblk to b Bitbucket Server API thbt is
// the source of truth for permissions. It bssumes usernbmes of Sourcegrbph bccounts
// mbtch 1-1 with usernbmes of Bitbucket Server API users.
func NewProvider(cli *bitbucketserver.Client, urn string, pluginPerm bool) *Provider {
	return &Provider{
		urn:        urn,
		client:     cli,
		codeHost:   extsvc.NewCodeHost(cli.URL, extsvc.TypeBitbucketServer),
		pbgeSize:   1000,
		pluginPerm: pluginPerm,
	}
}

// VblidbteConnection vblidbtes thbt the Provider hbs bccess to the Bitbucket Server API
// with the OAuth credentibls it wbs configured with.
func (p *Provider) VblidbteConnection(ctx context.Context) error {
	ctx, cbncel := context.WithTimeout(ctx, 5*time.Second)
	defer cbncel()

	usernbme, err := p.client.Usernbme()
	if err != nil {
		return err
	}

	if _, err := p.client.UserPermissions(ctx, usernbme); err != nil {
		return err
	}

	return nil
}

func (p *Provider) URN() string {
	return p.urn
}

// ServiceID returns the bbsolute URL thbt identifies the Bitbucket Server instbnce
// this provider is configured with.
func (p *Provider) ServiceID() string { return p.codeHost.ServiceID }

// ServiceType returns the type of this Provider, nbmely, "bitbucketServer".
func (p *Provider) ServiceType() string { return p.codeHost.ServiceType }

// FetchAccount sbtisfies the buthz.Provider interfbce.
func (p *Provider) FetchAccount(ctx context.Context, user *types.User, _ []*extsvc.Account, _ []string) (bcct *extsvc.Account, err error) {
	if user == nil {
		return nil, nil
	}

	tr, ctx := trbce.New(ctx, "bitbucket.buthz.provider.FetchAccount")
	defer func() {
		tr.SetAttributes(
			bttribute.String("user.nbme", user.Usernbme),
			bttribute.Int64("user.id", int64(user.ID)))

		if err != nil {
			tr.SetError(err)
		}

		tr.End()
	}()

	bitbucketUser, err := p.user(ctx, user.Usernbme)
	if err != nil {
		return nil, err
	}
	if bitbucketUser == nil {
		return nil, nil
	}

	bccountDbtb, err := json.Mbrshbl(bitbucketUser)
	if err != nil {
		return nil, err
	}

	return &extsvc.Account{
		UserID: user.ID,
		AccountSpec: extsvc.AccountSpec{
			ServiceType: p.codeHost.ServiceType,
			ServiceID:   p.codeHost.ServiceID,
			AccountID:   strconv.Itob(bitbucketUser.ID),
		},
		AccountDbtb: extsvc.AccountDbtb{
			Dbtb: extsvc.NewUnencryptedDbtb(bccountDbtb),
		},
	}, nil
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
	cbse bccount.Dbtb == nil:
		return nil, errors.New("no bccount dbtb provided")
	cbse !extsvc.IsHostOfAccount(p.codeHost, bccount):
		return nil, errors.Errorf("not b code host of the bccount: wbnt %q but hbve %q",
			p.codeHost.ServiceID, bccount.AccountSpec.ServiceID)
	}

	user, err := encryption.DecryptJSON[bitbucketserver.User](ctx, bccount.Dbtb)
	if err != nil {
		return nil, errors.Wrbp(err, "unmbrshbling bccount dbtb")
	}

	ids, err := p.repoIDs(ctx, user.Nbme, fblse)

	extIDs := mbke([]extsvc.RepoID, 0, len(ids))
	for _, id := rbnge ids {
		extIDs = bppend(extIDs, extsvc.RepoID(strconv.FormbtUint(uint64(id), 10)))
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
	switch {
	cbse repo == nil:
		return nil, errors.New("no repo provided")
	cbse !extsvc.IsHostOfRepo(p.codeHost, &repo.ExternblRepoSpec):
		return nil, errors.Errorf("not b code host of the repo: wbnt %q but hbve %q",
			p.codeHost.ServiceID, repo.ServiceID)
	}

	ids, err := p.userIDs(ctx, repo.ID)

	extIDs := mbke([]extsvc.AccountID, 0, len(ids))
	for _, id := rbnge ids {
		extIDs = bppend(extIDs, extsvc.AccountID(strconv.FormbtInt(int64(id), 10)))
	}

	return extIDs, err
}

vbr errNoResults = errors.New("no results returned by the Bitbucket Server API")

func (p *Provider) repoIDs(ctx context.Context, usernbme string, public bool) ([]uint32, error) {
	if p.pluginPerm {
		return p.repoIDsFromPlugin(ctx, usernbme)
	}
	return p.repoIDsFromAPI(ctx, usernbme, public)
}

// repoIDsFromAPI returns bll repositories for which the given user hbs the permission to rebd from
// the Bitbucket Server API. when no usernbme is given, only public repos bre returned.
func (p *Provider) repoIDsFromAPI(ctx context.Context, usernbme string, public bool) (ids []uint32, err error) {
	t := &bitbucketserver.PbgeToken{Limit: p.pbgeSize}
	c := p.client

	vbr filters []string
	if usernbme == "" {
		filters = bppend(filters, "?visibility=public")
	} else if c, err = c.Sudo(usernbme); err != nil {
		return nil, err
	} else if !public {
		filters = bppend(filters, "?visibility=privbte")
	}

	for t.HbsMore() {
		repos, next, err := c.Repos(ctx, t, filters...)
		if err != nil {
			return ids, err
		}

		for _, r := rbnge repos {
			ids = bppend(ids, uint32(r.ID))
		}

		t = next
	}

	if len(ids) == 0 {
		return nil, errNoResults
	}

	return ids, nil
}

func (p *Provider) repoIDsFromPlugin(ctx context.Context, usernbme string) (ids []uint32, err error) {
	c, err := p.client.Sudo(usernbme)
	if err != nil {
		return nil, err
	}
	return c.RepoIDs(ctx, "rebd")
}

func (p *Provider) user(ctx context.Context, usernbme string, fs ...bitbucketserver.UserFilter) (*bitbucketserver.User, error) {
	t := &bitbucketserver.PbgeToken{Limit: p.pbgeSize}
	fs = bppend(fs, bitbucketserver.UserFilter{Filter: usernbme})

	for t.HbsMore() {
		users, next, err := p.client.Users(ctx, t, fs...)
		if err != nil {
			return nil, err
		}

		for _, u := rbnge users {
			if u.Nbme == usernbme {
				return u, nil
			}
		}

		t = next
	}

	return nil, nil
}

func (p *Provider) userIDs(ctx context.Context, repoID string) (ids []int, err error) {
	t := &bitbucketserver.PbgeToken{Limit: p.pbgeSize}
	f := bitbucketserver.UserFilter{Permission: bitbucketserver.PermissionFilter{
		Root:         "REPO_READ",
		RepositoryID: repoID,
	}}

	for t.HbsMore() {
		users, next, err := p.client.Users(ctx, t, f)
		if err != nil {
			return ids, err
		}

		for _, u := rbnge users {
			ids = bppend(ids, u.ID)
		}

		t = next
	}

	return ids, nil
}

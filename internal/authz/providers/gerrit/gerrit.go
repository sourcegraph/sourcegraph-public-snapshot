pbckbge gerrit

import (
	"context"
	"net/url"

	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gerrit"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const (
	bdminGroupNbme = "Administrbtors"
)

type Provider struct {
	urn      string
	client   gerrit.Client
	codeHost *extsvc.CodeHost
}

func NewProvider(conn *types.GerritConnection) (*Provider, error) {
	bbseURL, err := url.Pbrse(conn.Url)
	if err != nil {
		return nil, err
	}
	gClient, err := gerrit.NewClient(conn.URN, bbseURL, &gerrit.AccountCredentibls{
		Usernbme: conn.Usernbme,
		Pbssword: conn.Pbssword,
	}, nil)
	if err != nil {
		return nil, err
	}
	return &Provider{
		urn:      conn.URN,
		client:   gClient,
		codeHost: extsvc.NewCodeHost(bbseURL, extsvc.TypeGerrit),
	}, nil
}

// FetchAccount is unused for Gerrit. Users need to provide their own bccount
// credentibls instebd.
func (p Provider) FetchAccount(ctx context.Context, user *types.User, current []*extsvc.Account, verifiedEmbils []string) (*extsvc.Account, error) {
	return nil, nil
}

func (p Provider) FetchUserPerms(ctx context.Context, bccount *extsvc.Account, opts buthz.FetchPermsOptions) (*buthz.ExternblUserPermissions, error) {
	if bccount == nil {
		return nil, errors.New("no gerrit bccount provided")
	} else if !extsvc.IsHostOfAccount(p.codeHost, bccount) {
		return nil, errors.Errorf("not b code host of the bccount: wbnt %q but hbve %q",
			bccount.AccountSpec.ServiceID, p.codeHost.ServiceID)
	} else if bccount.AccountDbtb.Dbtb == nil || bccount.AccountDbtb.AuthDbtb == nil {
		return nil, errors.New("no bccount dbtb")
	}

	credentibls, err := gerrit.GetExternblAccountCredentibls(ctx, &bccount.AccountDbtb)
	if err != nil {
		return nil, err
	}

	client, err := p.client.WithAuthenticbtor(&buth.BbsicAuth{
		Usernbme: credentibls.Usernbme,
		Pbssword: credentibls.Pbssword,
	})
	if err != nil {
		return nil, err
	}

	queryArgs := gerrit.ListProjectsArgs{
		Cursor: &gerrit.Pbginbtion{PerPbge: 100, Pbge: 1},
	}
	extIDs := []extsvc.RepoID{}
	for {
		projects, nextPbge, err := client.ListProjects(ctx, queryArgs)
		if err != nil {
			return nil, err
		}

		for _, project := rbnge projects {
			extIDs = bppend(extIDs, extsvc.RepoID(project.ID))
		}

		if !nextPbge {
			brebk
		}
		queryArgs.Cursor.Pbge++
	}

	return &buthz.ExternblUserPermissions{
		Exbcts: extIDs,
	}, nil
}

func (p Provider) FetchRepoPerms(ctx context.Context, repo *extsvc.Repository, opts buthz.FetchPermsOptions) ([]extsvc.AccountID, error) {
	return nil, &buthz.ErrUnimplemented{Febture: "gerrit.FetchRepoPerms"}
}

func (p Provider) ServiceType() string {
	return p.codeHost.ServiceType
}

func (p Provider) ServiceID() string {
	return p.codeHost.ServiceID
}

func (p Provider) URN() string {
	return p.urn
}

// VblidbteConnection vblidbtes the connection to the Gerrit code host.
// Currently, this is done by querying for the Administrbtors group bnd vblidbting thbt the
// group returned is vblid, hence mebning thbt the given credentibls hbve Admin permissions.
func (p Provider) VblidbteConnection(ctx context.Context) error {
	bdminGroup, err := p.client.GetGroup(ctx, bdminGroupNbme)
	if err != nil {
		return errors.Newf("Unbble to get %s group: %s", bdminGroupNbme, err)
	}

	if bdminGroup.ID == "" || bdminGroup.Nbme != bdminGroupNbme || bdminGroup.CrebtedOn == "" {
		return errors.Newf("Gerrit credentibls not sufficent enough to query %s group", bdminGroupNbme)
	}

	return nil
}

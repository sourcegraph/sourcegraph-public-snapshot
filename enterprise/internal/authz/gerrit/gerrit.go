package gerrit

import (
	"context"
	"net/url"

	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gerrit"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const (
	adminGroupName = "Administrators"
)

type Provider struct {
	urn      string
	client   client
	codeHost *extsvc.CodeHost
}

func NewProvider(conn *types.GerritConnection) (*Provider, error) {
	baseURL, err := url.Parse(conn.Url)
	if err != nil {
		return nil, err
	}
	gClient, err := gerrit.NewClient(conn.URN, baseURL, &gerrit.AccountCredentials{
		Username: conn.Username,
		Password: conn.Password,
	}, nil)
	if err != nil {
		return nil, err
	}
	return &Provider{
		urn:      conn.URN,
		client:   &ClientAdapter{Client: gClient},
		codeHost: extsvc.NewCodeHost(baseURL, extsvc.TypeGerrit),
	}, nil
}

// FetchAccount is unused for Gerrit. Users need to provide their own account
// credentials instead.
func (p Provider) FetchAccount(ctx context.Context, user *types.User, current []*extsvc.Account, verifiedEmails []string) (*extsvc.Account, error) {
	return nil, nil
}

func (p Provider) FetchUserPerms(ctx context.Context, account *extsvc.Account, opts authz.FetchPermsOptions) (*authz.ExternalUserPermissions, error) {
	if account == nil {
		return nil, errors.New("no account provided")
	} else if !extsvc.IsHostOfAccount(p.codeHost, account) {
		return nil, errors.Errorf("not a code host of the account: want %q but have %q",
			account.AccountSpec.ServiceID, p.codeHost.ServiceID)
	} else if account.AccountData.Data == nil || account.AccountData.AuthData == nil {
		return nil, errors.New("no account data")
	}

	_, credentials, err := gerrit.GetExternalAccountData(ctx, &account.AccountData)
	if err != nil {
		return nil, err
	}

	client := p.client.WithAuthenticator(&auth.BasicAuth{
		Username: credentials.Username,
		Password: credentials.Password,
	})
	queryArgs := gerrit.ListProjectsArgs{
		Cursor: &gerrit.Pagination{
			PerPage: 100,
			Page:    1,
		},
	}
	projects, nextPage, err := client.ListProjects(ctx, queryArgs)
	if err != nil {
		return nil, err
	}
	var nextPageProjects *gerrit.ListProjectsResponse
	for nextPage {
		queryArgs.Cursor.Page++
		nextPageProjects, nextPage, err = client.ListProjects(ctx, queryArgs)
		if err != nil {
			return nil, err
		}

		for k, v := range *nextPageProjects {
			(*projects)[k] = v
		}
	}

	extIDs := make([]extsvc.RepoID, 0, len(*projects))
	for _, project := range *projects {
		extIDs = append(extIDs, extsvc.RepoID(project.ID))
	}
	return &authz.ExternalUserPermissions{
		Exacts: extIDs,
	}, nil
}

func (p Provider) FetchRepoPerms(ctx context.Context, repo *extsvc.Repository, opts authz.FetchPermsOptions) ([]extsvc.AccountID, error) {
	return nil, &authz.ErrUnimplemented{Feature: "gerrit.FetchRepoPerms"}
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

// ValidateConnection validates the connection to the Gerrit code host.
// Currently, this is done by querying for the Administrators group and validating that the
// group returned is valid, hence meaning that the given credentials have Admin permissions.
func (p Provider) ValidateConnection(ctx context.Context) error {
	adminGroup, err := p.client.GetGroup(ctx, adminGroupName)
	if err != nil {
		return errors.Newf("Unable to get %s group: %s", adminGroupName, err)
	}

	if adminGroup.ID == "" || adminGroup.Name != adminGroupName || adminGroup.CreatedOn == "" {
		return errors.Newf("Gerrit credentials not sufficent enough to query %s group", adminGroupName)
	}

	return nil
}

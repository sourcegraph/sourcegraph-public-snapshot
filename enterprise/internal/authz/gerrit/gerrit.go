package gerrit

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	jsoniter "github.com/json-iterator/go"

	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gerrit"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
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
	gClient, err := gerrit.NewClient(conn.URN, conn.GerritConnection, nil)
	if err != nil {
		return nil, err
	}
	return &Provider{
		urn:      conn.URN,
		client:   gClient,
		codeHost: extsvc.NewCodeHost(baseURL, extsvc.TypeGerrit),
	}, nil
}

func (p Provider) FetchAccount(ctx context.Context, user *types.User, current []*extsvc.Account, verifiedEmails []string) (*extsvc.Account, error) {
	// First try to fetch Gerrit account for this username
	accts, err := p.client.ListAccountsByUsername(ctx, user.Username)
	if err != nil {
		return nil, err
	}
	// Check that this account from Gerrit correlates to a verified email
	if acct, found, err := p.checkAccountsAgainstVerifiedEmails(accts, user, verifiedEmails); found && err == nil {
		return acct, nil
	}

	// If no account was found via the user's Sourcegraph username, attempt to find an account via one of the verified emails.
	for _, email := range verifiedEmails {
		accts, err := p.client.ListAccountsByEmail(ctx, email)
		if err != nil {
			return nil, err
		}
		for _, acct := range accts {
			return p.buildExtsvcAccount(acct, user, email)
		}
	}

	return nil, nil
}

func (p Provider) checkAccountsAgainstVerifiedEmails(accts gerrit.ListAccountsResponse, user *types.User, verifiedEmails []string) (*extsvc.Account, bool, error) {
	if accts == nil || len(accts) == 0 {
		return nil, false, nil
	}
	for _, email := range verifiedEmails {
		for _, acct := range accts {
			if acct.Email == email && acct.Username == user.Username {
				foundAcct, err := p.buildExtsvcAccount(acct, user, email)
				return foundAcct, true, err
			}
		}
	}
	return nil, false, nil
}

func (p Provider) buildExtsvcAccount(acct gerrit.Account, user *types.User, email string) (*extsvc.Account, error) {
	acctData, err := marshalAccountData(acct.Username, acct.Email, acct.ID)
	if err != nil {
		return nil, errors.Wrap(err, "marshaling account data")
	}
	return &extsvc.Account{
		UserID: user.ID,
		AccountSpec: extsvc.AccountSpec{
			ServiceType: p.codeHost.ServiceType,
			ServiceID:   p.codeHost.ServiceID,
			AccountID:   email,
		},
		AccountData: extsvc.AccountData{
			Data: acctData,
		},
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

func marshalAccountData(username, email string, acctID int32) (*json.RawMessage, error) {
	accountData, err := jsoniter.Marshal(
		gerrit.AccountData{
			Username:  username,
			Email:     email,
			AccountID: acctID,
		},
	)
	if err != nil {
		return nil, err
	}
	return (*json.RawMessage)(&accountData), nil
}

func (p Provider) FetchUserPerms(ctx context.Context, account *extsvc.Account, opts authz.FetchPermsOptions) (*authz.ExternalUserPermissions, error) {
	// fetch account data from Gerrit by the account id
	user, err := gerrit.GetExternalAccountData(&account.AccountData)
	if err != nil {
		return nil, err
	}

	// Get all groups that this account is a member of
	groups, err := p.client.GetAccountGroups(ctx, user.AccountID)
	if err != nil {
		return nil, err
	}

	// get a list of repos to check
	for {
		// TODO: pagination?
		page, nextPage, err := p.client.ListProjects(ctx, gerrit.ListProjectsArgs{})
		if err != nil {
			// TODO: handle error?
			continue
		}
		names := getProjectNamesFromMap(page)
		fmt.Printf("names: %v\n", names)
		resp, err := p.client.GetProjectAccess(ctx, names...)
		if err != nil {
			//todo handle error
			fmt.Printf("Error getting project access: %s", err)
			continue
		}
		interpretProjectAccess(resp, groups)
		if !nextPage {
			break
		}
	}
	return nil, &authz.ErrUnimplemented{Feature: "gerrit.FetchUserPerms"}
}

func getProjectNamesFromMap(page *gerrit.ListProjectsResponse) []string {
	names := make([]string, 0, len(*page))
	for name := range *page {
		names = append(names, name)
	}
	return names
}

func interpretProjectAccess(accessResp gerrit.GetProjectAccessResponse, groups gerrit.GetAccountGroupsResponse) {
	for proj, access := range accessResp {
		fmt.Printf("Access for project: %s\n", proj)
		fmt.Printf("Groups? %v\n", access.Groups)
		fmt.Printf("Inherits from? %v\n", access.InheritsFrom)
		fmt.Println()
	}
}

func (p Provider) FetchRepoPerms(ctx context.Context, repo *extsvc.Repository, opts authz.FetchPermsOptions) ([]extsvc.AccountID, error) {
	return nil, &authz.ErrUnimplemented{Feature: "gerrit.FetchRepoPerms"}
}

func (p Provider) FetchUserPermsByToken(ctx context.Context, token string, opts authz.FetchPermsOptions) (*authz.ExternalUserPermissions, error) {
	return nil, &authz.ErrUnimplemented{Feature: "gerrit.FetchUserPermsByToken"}
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

func (p Provider) ValidateConnection(ctx context.Context) (warnings []string) {
	return nil
}

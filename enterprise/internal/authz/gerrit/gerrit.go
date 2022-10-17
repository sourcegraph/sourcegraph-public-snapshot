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
			Data: extsvc.NewUnencryptedData(acctData),
		},
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

func marshalAccountData(username, email string, acctID int32) (json.RawMessage, error) {
	return jsoniter.Marshal(
		gerrit.AccountData{
			Username:  username,
			Email:     email,
			AccountID: acctID,
		},
	)
}

func (p Provider) FetchUserPerms(ctx context.Context, account *extsvc.Account, opts authz.FetchPermsOptions) (*authz.ExternalUserPermissions, error) {
	return nil, &authz.ErrUnimplemented{Feature: "gerrit.FetchUserPerms"}
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
func (p Provider) ValidateConnection(ctx context.Context) (warnings []string) {

	adminGroup, err := p.client.GetGroup(ctx, adminGroupName)
	if err != nil {
		return []string{
			fmt.Sprintf("Unable to get %s group: %v", adminGroupName, err),
		}
	}

	if adminGroup.ID == "" || adminGroup.Name != adminGroupName || adminGroup.CreatedOn == "" {
		return []string{
			fmt.Sprintf("Gerrit credentials not sufficent enough to query %s group", adminGroupName),
		}
	}

	return []string{}
}

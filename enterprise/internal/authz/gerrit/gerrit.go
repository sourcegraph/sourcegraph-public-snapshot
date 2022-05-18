package gerrit

import (
	"context"
	"encoding/json"
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

func NewProvider(conn *types.GerritConnection) *Provider {
	baseURL, _ := url.Parse(conn.Url)
	gClient, err := gerrit.NewClient(conn.URN, conn.GerritConnection, nil)
	if err != nil {
		//TODO: handle error
	}
	return &Provider{
		urn:      conn.URN,
		client:   gClient,
		codeHost: extsvc.NewCodeHost(baseURL, extsvc.TypeGerrit),
	}
}

func (p Provider) FetchAccount(ctx context.Context, user *types.User, current []*extsvc.Account, verifiedEmails []string) (*extsvc.Account, error) {
	// First try to fetch Gerrit account for this username
	accts, err := p.client.ListAccountsByUsername(ctx, user.Username)
	if err != nil {
		return nil, err
	}
	// Check that this account from Gerrit correlates to a verified email
	if acct, found := p.checkAccountsAgainstVerifiedEmails(accts, user, verifiedEmails); found {
		return acct, nil
	}

	// If no account was found via the user's Sourcegraph username, attempt to find an account via one of the verified emails.
	for _, email := range verifiedEmails {
		accts, err := p.client.ListAccountsByEmail(ctx, email)
		if err != nil {
			return nil, err
		}
		for _, acct := range accts {
			return p.createAccount(acct, user, email)
		}
	}

	return nil, nil
}

func (p Provider) checkAccountsAgainstVerifiedEmails(accts gerrit.ListAccountsResponse, user *types.User, verifiedEmails []string) (*extsvc.Account, bool) {
	if accts != nil || len(accts) > 0 {
		for _, email := range verifiedEmails {
			for _, acct := range accts {
				if acct.Email == email && acct.Username == user.Username {
					foundAcct, _ := p.createAccount(acct, user, email)
					// todo: handle error
					return foundAcct, true
				}
			}
		}
	}
	return nil, false
}

// TODO: better naming. This seems to imply we're actually making something (i.e. storing in a db) but we're just creating the object
func (p Provider) createAccount(acct gerrit.Account, user *types.User, email string) (*extsvc.Account, error) {
	acctData, err := marshalAccountData(acct.Username, acct.Email, acct.ID)
	if err != nil {
		return nil, errors.Wrap(err, "marshaling account data")
	}
	return &extsvc.Account{
		ID:     acct.ID, // TODO: do we need this?
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
	return nil, &authz.ErrUnimplemented{Feature: "gerrit.FetchUserPerms"}
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

package github

import (
	"context"

	"github.com/google/go-github/github"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"golang.org/x/oauth2"
)

func GetExternalAccountData(data *extsvc.ExternalAccountData) (usr *github.User, tok *oauth2.Token, err error) {
	var (
		u github.User
		t oauth2.Token
	)

	if data.AccountData != nil {
		if err := data.GetAccountData(&u); err != nil {
			return nil, nil, err
		}
		usr = &u
	}
	if data.AuthData != nil {
		if err := data.GetAuthData(&t); err != nil {
			return nil, nil, err
		}
		tok = &t
	}
	return usr, tok, nil
}

func SetExternalAccountData(data *extsvc.ExternalAccountData, user *github.User, token *oauth2.Token) {
	data.SetAccountData(user)
	data.SetAuthData(token)
}

type UserEmail struct {
	Email      string `json:"email,omitempty"`
	Primary    bool   `json:"primary,omitempty"`
	Verified   bool   `json:"verified,omitempty"`
	Visibility string `json:"visibility,omitempty"`
}

var MockGetAuthenticatedUserEmails func(ctx context.Context) ([]*UserEmail, error)

// GetAuthenticatedUserEmails returns the first 100 emails associated with the currently
// authenticated user.
func (c *Client) GetAuthenticatedUserEmails(ctx context.Context) ([]*UserEmail, error) {
	if MockGetAuthenticatedUserEmails != nil {
		return MockGetAuthenticatedUserEmails(ctx)
	}

	var emails []*UserEmail
	err := c.requestGet(ctx, "", "/user/emails?per_page=100", &emails)
	if err != nil {
		return nil, err
	}
	return emails, nil
}

type Org struct {
	Login string `json:"login,omitempty"`
}

var MockGetAuthenticatedUserOrgs func(ctx context.Context) ([]*Org, error)

// GetAuthenticatedUserOrgs returns the first 100 organizations associated with the currently
// authenticated user.
func (c *Client) GetAuthenticatedUserOrgs(ctx context.Context) ([]*Org, error) {
	if MockGetAuthenticatedUserOrgs != nil {
		return MockGetAuthenticatedUserOrgs(ctx)
	}

	var orgs []*Org
	err := c.requestGet(ctx, "", "/user/orgs?per_page=100", &orgs)
	if err != nil {
		return nil, err
	}
	return orgs, nil
}

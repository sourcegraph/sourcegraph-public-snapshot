package sams

import (
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"golang.org/x/oauth2"

	accountsv1 "github.com/sourcegraph/sourcegraph-accounts-sdk-go/accounts/v1"
)

type AccountsV1Config struct {
	ConnConfig
	// TokenSource is the OAuth2 token source to use for authentication. It MUST be
	// based on a per-user token that is on behalf of a SAMS user.
	//
	// If you have the SAMS user's refresh token, using the oauth2.TokenSource
	// abstraction will take care of creating short-lived access tokens and refresh
	// as needed. But if you only have the access token, you will need to use a
	// StaticTokenSource instead.
	TokenSource oauth2.TokenSource
}

func (c AccountsV1Config) Validate() error {
	if err := c.ConnConfig.Validate(); err != nil {
		return errors.Wrap(err, "invalid ConnConfig")
	}
	if c.TokenSource == nil {
		return errors.New("token source is required")
	}
	return nil
}

// NewAccountsV1 returns a new SAMS client for interacting with Accounts API V1.
func NewAccountsV1(config AccountsV1Config) (*accountsv1.Client, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	return accountsv1.NewClient(config.getAPIURL(), config.TokenSource), nil
}

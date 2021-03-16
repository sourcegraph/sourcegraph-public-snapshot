package gitlab

import (
	"golang.org/x/oauth2"

	"github.com/sourcegraph/sourcegraph/internal/extsvc"
)

// GetExternalAccountData returns the deserialized user and token from the external account data
// JSON blob in a typesafe way.
func GetExternalAccountData(data *extsvc.AccountData) (usr *User, tok *oauth2.Token, err error) {
	var (
		u User
		t oauth2.Token
	)

	if data.Data != nil {
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

// SetExternalAccountData sets the user and token into the external account data blob.
func SetExternalAccountData(data *extsvc.AccountData, user *User, token *oauth2.Token) {
	data.SetAccountData(user)
	data.SetAuthData(token)
}

package gitlab

import (
	"github.com/sourcegraph/sourcegraph/pkg/extsvc"
	"golang.org/x/oauth2"
)

func GetExternalAccountData(data *extsvc.ExternalAccountData) (usr *User, tok *oauth2.Token, err error) {
	var (
		u User
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

func SetExternalAccountData(data *extsvc.ExternalAccountData, user *User, token *oauth2.Token) {
	data.SetAccountData(user)
	data.SetAuthData(token)
}

package github

import (
	"github.com/google/go-github/github"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc"
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

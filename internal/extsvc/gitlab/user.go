package gitlab

import (
	"context"
	"encoding/json"

	"golang.org/x/oauth2"

	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
)

// GetExternalAccountData returns the deserialized user and token from the external account data
// JSON blob in a typesafe way.
func GetExternalAccountData(ctx context.Context, data *extsvc.AccountData) (usr *User, tok *oauth2.Token, err error) {
	if data.Data != nil {
		var u User
		if err := encryption.DecryptJSON(ctx, data.Data, &u); err != nil {
			return nil, nil, err
		}

		usr = &u
	}

	if data.AuthData != nil {
		var t oauth2.Token
		if err := encryption.DecryptJSON(ctx, data.AuthData, &t); err != nil {
			return nil, nil, err
		}

		tok = &t
	}

	return usr, tok, nil
}

// SetExternalAccountData sets the user and token into the external account data blob.
func SetExternalAccountData(data *extsvc.AccountData, user *User, token *oauth2.Token) error {
	serializedUser, err := json.Marshal(user)
	if err != nil {
		return err
	}
	serializedToken, err := json.Marshal(token)
	if err != nil {
		return err
	}

	data.Data = extsvc.NewUnencryptedData(serializedUser)
	data.AuthData = extsvc.NewUnencryptedData(serializedToken)
	return nil
}

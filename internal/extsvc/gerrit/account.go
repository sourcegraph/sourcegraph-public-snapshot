package gerrit

import (
	"context"
	"encoding/json"
	"net/url"

	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
)

// AccountData stores information of a Gerrit account.
type AccountData struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	AccountID int32  `json:"account_id"`
}

// AccountCredentials stores basic HTTP auth credentials for a Gerrit account.
type AccountCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// GetExternalAccountData extracts account data for the external account.
func GetExternalAccountData(ctx context.Context, data *extsvc.AccountData) (usr *AccountData, creds *AccountCredentials, err error) {
	if data.Data != nil {
		usr, err = encryption.DecryptJSON[AccountData](ctx, data.Data)
		if err != nil {
			return nil, nil, err
		}
	}

	if data.AuthData != nil {
		creds, err = encryption.DecryptJSON[AccountCredentials](ctx, data.AuthData)
		if err != nil {
			return nil, nil, err
		}
	}

	return usr, creds, nil
}

func GetPublicExternalAccountData(ctx context.Context, data *extsvc.AccountData) (*extsvc.PublicAccountData, error) {
	usr, _, err := GetExternalAccountData(ctx, data)
	if err != nil {
		return nil, err
	}
	blank := ""
	return &extsvc.PublicAccountData{
		DisplayName: &usr.Username,
		Login:       &blank,
		URL:         &blank,
	}, nil
}

func SetExternalAccountData(data *extsvc.AccountData, usr *Account, creds *AccountCredentials) error {
	serializedUser, err := json.Marshal(usr)
	if err != nil {
		return err
	}
	serializedCreds, err := json.Marshal(creds)
	if err != nil {
		return err
	}

	data.Data = extsvc.NewUnencryptedData(serializedUser)
	data.AuthData = extsvc.NewUnencryptedData(serializedCreds)
	return nil
}

var MockVerifyAccount func(context.Context, *url.URL, *AccountCredentials) (*Account, error)

func VerifyAccount(ctx context.Context, u *url.URL, creds *AccountCredentials) (*Account, error) {
	if MockVerifyAccount != nil {
		return MockVerifyAccount(ctx, u, creds)
	}

	client, err := NewClient("", u, creds, nil)
	if err != nil {
		return nil, err
	}
	return client.GetAuthenticatedUserAccount(ctx)
}

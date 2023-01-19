package gerrit

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

// AccountData stores information of a Gerrit account.
type AccountData struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	AccountID int32  `json:"account_id"`
}

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

func SetExternalAccountData(data *extsvc.AccountData, usr *AccountData, creds *AccountCredentials) error {
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

func VerifyAccount(ctx context.Context, conn *types.GerritConnection, creds *AccountCredentials) (*AccountData, error) {
	auther := &auth.BasicAuth{Username: creds.Username, Password: creds.Password}
	client, err := NewClient(conn.URN, conn.GerritConnection, nil)
	if err != nil {
		return nil, err
	}
	client = client.WithAuthenticator(auther)

	resp, err := client.ListAccountsByUsername(ctx, creds.Username)
	if err != nil {
		return nil, err
	}

	if len(resp) != 1 {
		if len(resp) == 0 {
			return nil, errors.New("no account found")
		} else {
			return nil, errors.New("multiple accounts found")
		}
	}

	return &AccountData{
		Username:  resp[0].Username,
		Email:     resp[0].Email,
		AccountID: resp[0].ID,
	}, nil
}

package gerrit

import (
	"context"
	"encoding/json"

	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
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

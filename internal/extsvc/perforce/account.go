package perforce

import (
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
)

// AccountData stores information of a Perforce Server account.
type AccountData struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

// GetExternalAccountData extracts account data for the external account.
func GetExternalAccountData(data *extsvc.AccountData) (accountData *AccountData, err error) {
	if data.Data != nil {
		var d AccountData
		if err = data.GetAccountData(&d); err != nil {
			return nil, err
		}
		accountData = &d
	}
	return accountData, nil
}

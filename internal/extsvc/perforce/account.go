package perforce

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
)

// AccountData stores information of a Perforce Server account.
type AccountData struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

// GetExternalAccountData extracts account data for the external account.
func GetExternalAccountData(ctx context.Context, data *extsvc.AccountData) (*AccountData, error) {
	if data.Data == nil {
		return nil, nil
	}

	return encryption.DecryptJSON[AccountData](ctx, data.Data)
}

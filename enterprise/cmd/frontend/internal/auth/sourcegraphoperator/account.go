package sourcegraphoperator

import (
	"context"
	"encoding/json"

	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
)

type ExternalAccountData struct {
	ServiceAccount bool `json:"serviceAccount"`
}

// GetAccountData parses account data and retrieves SOAP external account data.
func GetAccountData(ctx context.Context, data extsvc.AccountData) (*ExternalAccountData, error) {
	if data.Data == nil {
		return &ExternalAccountData{}, nil
	}
	return encryption.DecryptJSON[ExternalAccountData](ctx, data.Data)
}

// MarshalAccountData stores data into the external service account data format.
func MarshalAccountData(data ExternalAccountData) (extsvc.AccountData, error) {
	serializedData, err := json.Marshal(data)
	if err != nil {
		return extsvc.AccountData{}, err
	}
	return extsvc.AccountData{
		Data: extsvc.NewUnencryptedData(serializedData),
	}, nil
}

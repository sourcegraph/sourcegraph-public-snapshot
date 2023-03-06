package scim

import (
	"encoding/json"

	"github.com/elimity-com/scim"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
)

// AccountData stores information about a user that we don't have fields for in the schema.
type AccountData struct {
	Username string `json:"username"`
}

// toAccountData converts the given “SCIM resource attributes” type to an AccountData type.
func toAccountData(attributes scim.ResourceAttributes) (extsvc.AccountData, error) {
	serializedAccountData, err := json.Marshal(attributes)
	if err != nil {
		return extsvc.AccountData{}, err
	}

	return extsvc.AccountData{
		AuthData: nil,
		Data:     extsvc.NewUnencryptedData(serializedAccountData),
	}, nil
}

// fromAccountData converts the given account data JSON to a “SCIM resource attributes” type.
func fromAccountData(scimAccountData string) (attributes scim.ResourceAttributes, err error) {
	err = json.Unmarshal([]byte(scimAccountData), &attributes)
	return
}

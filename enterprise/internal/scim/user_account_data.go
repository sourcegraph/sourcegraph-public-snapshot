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

func fromAccountData(scimAccountData string) (scim.ResourceAttributes, error) {
	var attributes scim.ResourceAttributes
	err := json.Unmarshal([]byte(scimAccountData), &attributes)
	if err != nil {
		return scim.ResourceAttributes{}, err
	}

	// TODO: Need manual processing of this data?
	//var data scim.ResourceAttributes
	//data = scim.ResourceAttributes{
	//	"userName":   accountData.Username,
	//	"name": map[string]interface{}{
	//		"givenName":  scim.ResourceAttributes.name.firstName,
	//		"middleName": scim.ResourceAttributes.middleName,
	//		"familyName": scim.ResourceAttributes.lastName,
	//		"formatted":  user.DisplayName,
	//	},
	//	"emails":      emailMap,
	//	"active":      true,
	//}
	return attributes, err
}

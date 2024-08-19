package alert

import (
	"net/http"

	"github.com/opsgenie/opsgenie-go-sdk-v2/client"
)

type DeleteSavedSearchRequest struct {
	client.BaseRequest
	IdentifierType  SearchIdentifierType
	IdentifierValue string
}

func (r *DeleteSavedSearchRequest) Validate() error {
	err := validateIdentifier(r.IdentifierValue)
	if err != nil {
		return err
	}
	return nil
}

func (r *DeleteSavedSearchRequest) ResourcePath() string {

	return "/v2/alerts/saved-searches/" + r.IdentifierValue
}

func (r *DeleteSavedSearchRequest) Method() string {
	return http.MethodDelete
}

func (r *DeleteSavedSearchRequest) RequestParams() map[string]string {

	params := make(map[string]string)

	if r.IdentifierType == NAME {
		params["identifierType"] = "name"

	} else {
		params["identifierType"] = "id"

	}
	return params
}

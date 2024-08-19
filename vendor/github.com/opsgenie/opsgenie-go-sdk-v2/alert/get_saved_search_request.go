package alert

import (
	"net/http"

	"github.com/opsgenie/opsgenie-go-sdk-v2/client"
)

type GetSavedSearchRequest struct {
	client.BaseRequest
	IdentifierType  SearchIdentifierType
	IdentifierValue string
}

func (r *GetSavedSearchRequest) Validate() error {
	err := validateIdentifier(r.IdentifierValue)
	if err != nil {
		return err
	}
	return nil
}

func (r *GetSavedSearchRequest) ResourcePath() string {

	return "/v2/alerts/saved-searches/" + r.IdentifierValue
}

func (r *GetSavedSearchRequest) Method() string {
	return http.MethodGet
}

func (r *GetSavedSearchRequest) RequestParams() map[string]string {

	params := make(map[string]string)

	if r.IdentifierType == NAME {
		params["identifierType"] = "name"

	} else {
		params["identifierType"] = "id"

	}
	return params
}

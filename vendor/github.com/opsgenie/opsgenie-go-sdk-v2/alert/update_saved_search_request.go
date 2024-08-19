package alert

import (
	"net/http"

	"github.com/opsgenie/opsgenie-go-sdk-v2/client"
	"github.com/pkg/errors"
)

type UpdateSavedSearchRequest struct {
	client.BaseRequest
	IdentifierType  SearchIdentifierType
	IdentifierValue string
	NewName         string `json:"name,omitempty"`
	Query           string `json:"query,omitempty"`
	Owner           User   `json:"owner,omitempty"`
	Description     string `json:"description,omitempty"`
	Teams           []Team `json:"teams,omitempty"`
}

func (r *UpdateSavedSearchRequest) Validate() error {

	if r.IdentifierValue == "" {
		return errors.New("Identifier can not be empty")
	}

	if r.NewName == "" {
		return errors.New("Name can not be empty")
	}

	if r.Query == "" {
		return errors.New("Query can not be empty")
	}

	if r.Owner.ID == "" && r.Owner.Username == "" {
		return errors.New("Owner can not be empty")
	}

	return nil
}

func (r *UpdateSavedSearchRequest) ResourcePath() string {

	return "/v2/alerts/saved-searches/" + r.IdentifierValue
}

func (r *UpdateSavedSearchRequest) Method() string {
	return http.MethodPatch
}

func (r *UpdateSavedSearchRequest) RequestParams() map[string]string {

	params := make(map[string]string)

	if r.IdentifierType == NAME {
		params["identifierType"] = "name"

	} else {
		params["identifierType"] = "id"

	}
	return params
}

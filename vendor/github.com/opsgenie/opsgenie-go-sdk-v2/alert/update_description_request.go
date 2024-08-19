package alert

import (
	"net/http"

	"github.com/opsgenie/opsgenie-go-sdk-v2/client"
	"github.com/pkg/errors"
)

type UpdateDescriptionRequest struct {
	client.BaseRequest
	IdentifierType  AlertIdentifier
	IdentifierValue string
	Description     string `json:"description,omitempty"`
}

func (r *UpdateDescriptionRequest) Validate() error {
	if r.Description == "" {
		return errors.New("Description can not be empty")
	}
	if r.IdentifierValue == "" {
		return errors.New("Identifier can not be empty")
	}
	return nil
}

func (r *UpdateDescriptionRequest) ResourcePath() string {

	return "/v2/alerts/" + r.IdentifierValue + "/description"
}

func (r *UpdateDescriptionRequest) Method() string {
	return http.MethodPut
}

func (r *UpdateDescriptionRequest) RequestParams() map[string]string {

	params := make(map[string]string)

	if r.IdentifierType == ALIAS {
		params["identifierType"] = "alias"

	} else if r.IdentifierType == TINYID {
		params["identifierType"] = "tiny"

	} else {
		params["identifierType"] = "id"

	}
	return params
}

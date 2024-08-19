package alert

import (
	"net/http"

	"github.com/opsgenie/opsgenie-go-sdk-v2/client"
	"github.com/pkg/errors"
)

type AddDetailsRequest struct {
	client.BaseRequest
	IdentifierType  AlertIdentifier
	IdentifierValue string
	Details         map[string]string `json:"details,omitempty"`
	User            string            `json:"user,omitempty"`
	Source          string            `json:"source,omitempty"`
	Note            string            `json:"note,omitempty"`
}

func (r *AddDetailsRequest) Validate() error {
	if len(r.Details) == 0 {
		return errors.New("Details can not be empty")
	}

	if r.IdentifierValue == "" {
		return errors.New("Identifier can not be empty")
	}
	return nil
}

func (r *AddDetailsRequest) ResourcePath() string {

	return "/v2/alerts/" + r.IdentifierValue + "/details"

}

func (r *AddDetailsRequest) Method() string {
	return http.MethodPost
}

func (r *AddDetailsRequest) RequestParams() map[string]string {

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

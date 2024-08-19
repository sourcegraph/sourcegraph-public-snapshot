package alert

import (
	"net/http"

	"github.com/opsgenie/opsgenie-go-sdk-v2/client"
	"github.com/pkg/errors"
)

type ExecuteCustomActionAlertRequest struct {
	client.BaseRequest
	IdentifierType  AlertIdentifier
	IdentifierValue string
	Action          string
	User            string `json:"user,omitempty"`
	Source          string `json:"source,omitempty"`
	Note            string `json:"note,omitempty"`
}

func (r *ExecuteCustomActionAlertRequest) Validate() error {
	if r.IdentifierValue == "" {
		return errors.New("Identifier can not be empty")
	}
	if r.Action == "" {
		return errors.New("Action can not be empty")
	}
	return nil
}

func (r *ExecuteCustomActionAlertRequest) ResourcePath() string {

	return "/v2/alerts/" + r.IdentifierValue + "/actions/" + r.Action

}

func (r *ExecuteCustomActionAlertRequest) Method() string {
	return http.MethodPost
}

func (r *ExecuteCustomActionAlertRequest) RequestParams() map[string]string {

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

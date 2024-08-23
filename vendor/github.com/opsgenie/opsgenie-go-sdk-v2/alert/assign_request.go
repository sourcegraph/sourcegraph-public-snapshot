package alert

import (
	"net/http"

	"github.com/opsgenie/opsgenie-go-sdk-v2/client"
	"github.com/pkg/errors"
)

type AssignRequest struct {
	client.BaseRequest
	IdentifierType  AlertIdentifier
	IdentifierValue string
	Owner           User   `json:"owner,omitempty"`
	User            string `json:"user,omitempty"`
	Source          string `json:"source,omitempty"`
	Note            string `json:"note,omitempty"`
}

func (r *AssignRequest) Validate() error {
	if r.Owner.ID == "" && r.Owner.Username == "" {
		return errors.New("Owner ID or username must be defined")
	}

	if r.IdentifierValue == "" {
		return errors.New("Identifier can not be empty")
	}
	return nil
}

func (r *AssignRequest) ResourcePath() string {

	return "/v2/alerts/" + r.IdentifierValue + "/assign"

}

func (r *AssignRequest) Method() string {
	return http.MethodPost
}

func (r *AssignRequest) RequestParams() map[string]string {

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

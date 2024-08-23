package alert

import (
	"net/http"

	"github.com/opsgenie/opsgenie-go-sdk-v2/client"
	"github.com/pkg/errors"
)

type EscalateToNextRequest struct {
	client.BaseRequest
	IdentifierType  AlertIdentifier
	IdentifierValue string
	Escalation      Escalation `json:"escalation,omitempty"`
	User            string     `json:"user,omitempty"`
	Source          string     `json:"source,omitempty"`
	Note            string     `json:"note,omitempty"`
}

func (r *EscalateToNextRequest) Validate() error {
	if r.Escalation.ID == "" && r.Escalation.Name == "" {
		return errors.New("Escalation ID or name must be defined")
	}

	if r.IdentifierValue == "" {
		return errors.New("Identifier can not be empty")
	}
	return nil
}

func (r *EscalateToNextRequest) ResourcePath() string {

	return "/v2/alerts/" + r.IdentifierValue + "/escalate"

}

func (r *EscalateToNextRequest) Method() string {
	return http.MethodPost
}

func (r *EscalateToNextRequest) RequestParams() map[string]string {

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

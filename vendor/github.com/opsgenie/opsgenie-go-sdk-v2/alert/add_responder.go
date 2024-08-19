package alert

import (
	"net/http"

	"github.com/opsgenie/opsgenie-go-sdk-v2/client"
	"github.com/pkg/errors"
)

type AddResponderRequest struct {
	client.BaseRequest
	IdentifierType  AlertIdentifier
	IdentifierValue string
	Responder       Responder `json:"responder,omitempty"`
	User            string    `json:"user,omitempty"`
	Source          string    `json:"source,omitempty"`
	Note            string    `json:"note,omitempty"`
}

func (r *AddResponderRequest) Validate() error {

	if r.Responder.Type != UserResponder && r.Responder.Type != TeamResponder {
		return errors.New("Responder type must be user or team")
	}
	if r.Responder.Type == UserResponder && r.Responder.Id == "" && r.Responder.Username == "" {
		return errors.New("User ID or username must be defined")
	}
	if r.Responder.Type == TeamResponder && r.Responder.Id == "" && r.Responder.Name == "" {
		return errors.New("Team ID or name must be defined")
	}

	if r.IdentifierValue == "" {
		return errors.New("Identifier can not be empty")
	}
	return nil
}

func (r *AddResponderRequest) ResourcePath() string {

	return "/v2/alerts/" + r.IdentifierValue + "/responders"

}

func (r *AddResponderRequest) Method() string {
	return http.MethodPost
}

func (r *AddResponderRequest) RequestParams() map[string]string {

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

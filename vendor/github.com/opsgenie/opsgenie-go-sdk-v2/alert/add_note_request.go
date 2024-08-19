package alert

import (
	"net/http"

	"github.com/opsgenie/opsgenie-go-sdk-v2/client"
	"github.com/pkg/errors"
)

type AddNoteRequest struct {
	client.BaseRequest
	IdentifierType  AlertIdentifier
	IdentifierValue string
	User            string `json:"user,omitempty"`
	Source          string `json:"source,omitempty"`
	Note            string `json:"note,omitempty"`
}

func (r *AddNoteRequest) Validate() error {
	if r.Note == "" {
		return errors.New("Note can not be empty")
	}
	if r.IdentifierValue == "" {
		return errors.New("Identifier can not be empty")
	}
	return nil
}

func (r *AddNoteRequest) ResourcePath() string {

	return "/v2/alerts/" + r.IdentifierValue + "/notes"

}

func (r *AddNoteRequest) Method() string {
	return http.MethodPost
}

func (r *AddNoteRequest) RequestParams() map[string]string {

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

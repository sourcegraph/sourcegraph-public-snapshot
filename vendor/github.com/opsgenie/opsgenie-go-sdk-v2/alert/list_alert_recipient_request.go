package alert

import (
	"net/http"

	"github.com/opsgenie/opsgenie-go-sdk-v2/client"
)

type ListAlertRecipientRequest struct {
	client.BaseRequest
	IdentifierType  AlertIdentifier
	IdentifierValue string
}

func (r *ListAlertRecipientRequest) Validate() error {
	err := validateIdentifier(r.IdentifierValue)
	if err != nil {
		return err
	}
	return nil
}

func (r *ListAlertRecipientRequest) ResourcePath() string {

	return "/v2/alerts/" + r.IdentifierValue + "/recipients"
}

func (r *ListAlertRecipientRequest) Method() string {
	return http.MethodGet
}

func (r *ListAlertRecipientRequest) RequestParams() map[string]string {

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

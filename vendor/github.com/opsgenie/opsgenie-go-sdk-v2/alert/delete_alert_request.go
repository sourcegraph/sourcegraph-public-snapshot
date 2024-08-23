package alert

import (
	"net/http"

	"github.com/opsgenie/opsgenie-go-sdk-v2/client"
)

type DeleteAlertRequest struct {
	client.BaseRequest
	IdentifierType  AlertIdentifier
	IdentifierValue string
	Source          string
}

func (r *DeleteAlertRequest) Validate() error {
	err := validateIdentifier(r.IdentifierValue)
	if err != nil {
		return err
	}
	return nil
}

func (r *DeleteAlertRequest) ResourcePath() string {

	return "/v2/alerts/" + r.IdentifierValue
}

func (r *DeleteAlertRequest) Method() string {
	return http.MethodDelete
}

func (r *DeleteAlertRequest) RequestParams() map[string]string {

	params := make(map[string]string)

	if r.IdentifierType == ALERTID {
		params["identifierType"] = "id"

	} else if r.IdentifierType == ALIAS {
		params["identifierType"] = "alias"

	} else if r.IdentifierType == TINYID {
		params["identifierType"] = "tiny"
	}

	if r.Source != "" {
		params["source"] = r.Source
	}

	return params
}

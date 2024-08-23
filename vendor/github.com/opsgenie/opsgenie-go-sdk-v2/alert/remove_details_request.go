package alert

import (
	"net/http"

	"github.com/opsgenie/opsgenie-go-sdk-v2/client"
	"github.com/pkg/errors"
)

type RemoveDetailsRequest struct {
	client.BaseRequest
	IdentifierType  AlertIdentifier
	IdentifierValue string
	Keys            string
	Source          string
	User            string
	Note            string
}

func (r *RemoveDetailsRequest) Validate() error {
	if r.Keys == "" {
		return errors.New("Keys can not be empty")
	}

	if r.IdentifierValue == "" {
		return errors.New("Identifier can not be empty")
	}
	return nil
}

func (r *RemoveDetailsRequest) ResourcePath() string {

	return "/v2/alerts/" + r.IdentifierValue + "/details"
}

func (r *RemoveDetailsRequest) Method() string {
	return http.MethodDelete
}

func (r *RemoveDetailsRequest) RequestParams() map[string]string {

	params := make(map[string]string)

	if r.IdentifierType == ALIAS {
		params["identifierType"] = "alias"

	} else if r.IdentifierType == TINYID {
		params["identifierType"] = "tiny"

	} else {
		params["identifierType"] = "id"
	}

	if r.Keys != "" {
		params["keys"] = r.Keys
	}

	if r.Source != "" {
		params["source"] = r.Source
	}

	if r.User != "" {
		params["user"] = r.User
	}

	if r.Note != "" {
		params["note"] = r.Note
	}

	return params
}

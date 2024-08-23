package alert

import (
	"net/http"

	"github.com/opsgenie/opsgenie-go-sdk-v2/client"
	"github.com/pkg/errors"
)

type RemoveTagsRequest struct {
	client.BaseRequest
	IdentifierType  AlertIdentifier
	IdentifierValue string
	Tags            string
	Source          string
	User            string
	Note            string
}

func (r *RemoveTagsRequest) Validate() error {
	if r.Tags == "" {
		return errors.New("Tags can not be empty")
	}

	if r.IdentifierValue == "" {
		return errors.New("Identifier can not be empty")
	}
	return nil
}

func (r *RemoveTagsRequest) ResourcePath() string {

	return "/v2/alerts/" + r.IdentifierValue + "/tags"
}

func (r *RemoveTagsRequest) Method() string {
	return http.MethodDelete
}

func (r *RemoveTagsRequest) RequestParams() map[string]string {

	params := make(map[string]string)

	if r.IdentifierType == ALIAS {
		params["identifierType"] = "alias"

	} else if r.IdentifierType == TINYID {
		params["identifierType"] = "tiny"

	} else {
		params["identifierType"] = "id"
	}

	if r.Tags != "" {
		params["tags"] = r.Tags
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

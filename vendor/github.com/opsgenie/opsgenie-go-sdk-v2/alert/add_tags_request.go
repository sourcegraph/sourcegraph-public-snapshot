package alert

import (
	"net/http"

	"github.com/opsgenie/opsgenie-go-sdk-v2/client"
	"github.com/pkg/errors"
)

type AddTagsRequest struct {
	client.BaseRequest
	IdentifierType  AlertIdentifier
	IdentifierValue string
	Tags            []string `json:"tags,omitempty"`
	User            string   `json:"user,omitempty"`
	Source          string   `json:"source,omitempty"`
	Note            string   `json:"note,omitempty"`
}

func (r *AddTagsRequest) Validate() error {
	if len(r.Tags) == 0 {
		return errors.New("Tags list can not be empty")
	}

	if r.IdentifierValue == "" {
		return errors.New("Identifier can not be empty")
	}
	return nil
}

func (r *AddTagsRequest) ResourcePath() string {

	return "/v2/alerts/" + r.IdentifierValue + "/tags"

}

func (r *AddTagsRequest) Method() string {
	return http.MethodPost
}

func (r *AddTagsRequest) RequestParams() map[string]string {

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

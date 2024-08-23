package alert

import (
	"net/http"

	"github.com/opsgenie/opsgenie-go-sdk-v2/client"
	"github.com/pkg/errors"
)

type DeleteAttachmentRequest struct {
	client.BaseRequest
	IdentifierType  AlertIdentifier
	IdentifierValue string
	AttachmentId    string
	User            string
}

func (r *DeleteAttachmentRequest) Validate() error {
	if r.AttachmentId == "" {
		return errors.New("AttachmentId can not be empty")
	}

	if r.IdentifierValue == "" {
		return errors.New("Identifier can not be empty")
	}
	return nil
}

func (r *DeleteAttachmentRequest) ResourcePath() string {

	return "/v2/alerts/" + r.IdentifierValue + "/attachments/" + r.AttachmentId
}

func (r *DeleteAttachmentRequest) Method() string {
	return http.MethodDelete
}

func (r *DeleteAttachmentRequest) RequestParams() map[string]string {

	params := make(map[string]string)

	if r.IdentifierType == ALIAS {
		params["alertIdentifierType"] = "alias"

	} else if r.IdentifierType == TINYID {
		params["alertIdentifierType"] = "tiny"

	} else {
		params["alertIdentifierType"] = "id"

	}

	if r.User != "" {
		params["user"] = r.User
	}

	return params
}

package alert

import (
	"net/http"

	"github.com/opsgenie/opsgenie-go-sdk-v2/client"
	"github.com/pkg/errors"
)

type GetAttachmentRequest struct {
	client.BaseRequest
	IdentifierType  AlertIdentifier
	IdentifierValue string
	AttachmentId    string
}

func (r *GetAttachmentRequest) Validate() error {
	if r.AttachmentId == "" {
		return errors.New("AttachmentId can not be empty")
	}

	if r.IdentifierValue == "" {
		return errors.New("Identifier can not be empty")
	}
	return nil
}

func (r *GetAttachmentRequest) ResourcePath() string {

	return "/v2/alerts/" + r.IdentifierValue + "/attachments/" + r.AttachmentId
}

func (r *GetAttachmentRequest) Method() string {
	return http.MethodGet
}

func (r *GetAttachmentRequest) RequestParams() map[string]string {

	params := make(map[string]string)

	if r.IdentifierType == ALIAS {
		params["alertIdentifierType"] = "alias"

	} else if r.IdentifierType == TINYID {
		params["alertIdentifierType"] = "tiny"

	} else {
		params["alertIdentifierType"] = "id"

	}
	return params
}

package alert

import (
	"errors"
	"io"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/opsgenie/opsgenie-go-sdk-v2/client"
)

type CreateAlertAttachmentRequest struct {
	client.BaseRequest
	IdentifierType  AlertIdentifier
	IdentifierValue string
	FileName        string
	FilePath        string
	User            string
	IndexFile       string
}

func (r *CreateAlertAttachmentRequest) Metadata(ar client.ApiRequest) map[string]interface{} {
	headers := make(map[string]interface{})

	formDataMap := make(map[string]io.Reader)
	filePath := path.Join(r.FilePath, r.FileName)
	file, _ := os.Open(filePath)

	formDataMap["file"] = file
	formDataMap["user"] = strings.NewReader(r.User)
	formDataMap["indexFile"] = strings.NewReader(r.IndexFile)
	headers["form-data-values"] = formDataMap

	return headers
}

func (r *CreateAlertAttachmentRequest) Validate() error {
	if r.FileName == "" {
		return errors.New("FileName can not be empty")
	}
	if r.FilePath == "" {
		return errors.New("FilePath can not be empty")
	}
	if r.IdentifierValue == "" {
		return errors.New("Identifier can not be empty")
	}
	return nil
}

func (r *CreateAlertAttachmentRequest) ResourcePath() string {

	return "/v2/alerts/" + r.IdentifierValue + "/attachments"
}

func (r *CreateAlertAttachmentRequest) Method() string {
	return http.MethodPost
}

func (r *CreateAlertAttachmentRequest) RequestParams() map[string]string {

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

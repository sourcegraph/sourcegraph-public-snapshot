package alert

import (
	"net/http"
	"time"

	"github.com/opsgenie/opsgenie-go-sdk-v2/client"
	"github.com/pkg/errors"
)

type SnoozeAlertRequest struct {
	client.BaseRequest
	IdentifierType  AlertIdentifier
	IdentifierValue string
	EndTime         time.Time `json:"endTime,omitempty"`
	User            string    `json:"user,omitempty"`
	Source          string    `json:"source,omitempty"`
	Note            string    `json:"note,omitempty"`
}

func (r *SnoozeAlertRequest) Validate() error {
	if r.IdentifierValue == "" {
		return errors.New("Identifier can not be empty")
	}

	if time.Now().After(r.EndTime) {
		return errors.New("EndTime should at least be 2 seconds later.")
	}
	return nil
}

func (r *SnoozeAlertRequest) ResourcePath() string {

	return "/v2/alerts/" + r.IdentifierValue + "/snooze"

}

func (r *SnoozeAlertRequest) Method() string {
	return http.MethodPost
}

func (r *SnoozeAlertRequest) RequestParams() map[string]string {

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

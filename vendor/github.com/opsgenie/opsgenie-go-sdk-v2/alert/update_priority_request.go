package alert

import (
	"net/http"

	"github.com/opsgenie/opsgenie-go-sdk-v2/client"
	"github.com/pkg/errors"
)

type UpdatePriorityRequest struct {
	client.BaseRequest
	IdentifierType  AlertIdentifier
	IdentifierValue string
	Priority        Priority `json:"priority,omitempty"`
}

func (r *UpdatePriorityRequest) Validate() error {
	if r.Priority == "" {
		return errors.New("Priority can not be empty")
	}
	if r.IdentifierValue == "" {
		return errors.New("Identifier can not be empty")
	}
	err := ValidatePriority(r.Priority)
	if err != nil {
		return err
	}
	return nil
}

func (r *UpdatePriorityRequest) ResourcePath() string {

	return "/v2/alerts/" + r.IdentifierValue + "/priority"

}

func (r *UpdatePriorityRequest) Method() string {
	return http.MethodPut
}

func (r *UpdatePriorityRequest) RequestParams() map[string]string {

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
func ValidatePriority(priority Priority) error {
	switch priority {
	case P1, P2, P3, P4, P5:
		return nil
	}
	return errors.New("Priority should be one of these: " +
		"'P1', 'P2', 'P3', 'P4' and 'P5'")
}

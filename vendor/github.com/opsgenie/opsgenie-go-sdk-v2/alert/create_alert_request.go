package alert

import (
	"errors"
	"net/http"

	"github.com/opsgenie/opsgenie-go-sdk-v2/client"
)

type CreateAlertRequest struct {
	client.BaseRequest
	Message     string            `json:"message"`
	Alias       string            `json:"alias,omitempty"`
	Description string            `json:"description,omitempty"`
	Responders  []Responder       `json:"responders,omitempty"`
	VisibleTo   []Responder       `json:"visibleTo,omitempty"`
	Actions     []string          `json:"actions,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	Details     map[string]string `json:"details,omitempty"`
	Entity      string            `json:"entity,omitempty"`
	Source      string            `json:"source,omitempty"`
	Priority    Priority          `json:"priority,omitempty"`
	User        string            `json:"user,omitempty"`
	Note        string            `json:"note,omitempty"`
}

func (r *CreateAlertRequest) Validate() error {
	if r.Message == "" {
		return errors.New("message can not be empty")
	}
	return nil
}

func (r *CreateAlertRequest) ResourcePath() string {

	return "/v2/alerts"
}

func (r *CreateAlertRequest) Method() string {
	return http.MethodPost
}

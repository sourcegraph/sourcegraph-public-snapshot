package alert

import (
	"net/http"

	"github.com/opsgenie/opsgenie-go-sdk-v2/client"
	"github.com/pkg/errors"
)

type AddTeamRequest struct {
	client.BaseRequest
	IdentifierType  AlertIdentifier
	IdentifierValue string
	Team            Team   `json:"team,omitempty"`
	User            string `json:"user,omitempty"`
	Source          string `json:"source,omitempty"`
	Note            string `json:"note,omitempty"`
}

func (r *AddTeamRequest) Validate() error {
	if r.Team.ID == "" && r.Team.Name == "" {
		return errors.New("Team ID or name must be defined")
	}

	if r.IdentifierValue == "" {
		return errors.New("Identifier can not be empty")
	}
	return nil
}

func (r *AddTeamRequest) ResourcePath() string {

	return "/v2/alerts/" + r.IdentifierValue + "/teams"

}

func (r *AddTeamRequest) Method() string {
	return http.MethodPost
}

func (r *AddTeamRequest) RequestParams() map[string]string {

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

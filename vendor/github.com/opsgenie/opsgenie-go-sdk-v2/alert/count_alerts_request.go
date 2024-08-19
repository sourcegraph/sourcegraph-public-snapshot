package alert

import (
	"net/http"

	"github.com/opsgenie/opsgenie-go-sdk-v2/client"
)

type CountAlertsRequest struct {
	client.BaseRequest
	Query                string
	SearchIdentifier     string
	SearchIdentifierType SearchIdentifierType
}

func (r *CountAlertsRequest) Validate() error {
	return nil
}

func (r *CountAlertsRequest) ResourcePath() string {
	return "/v2/alerts/count"
}

func (r *CountAlertsRequest) Method() string {
	return http.MethodGet
}

func (r *CountAlertsRequest) RequestParams() map[string]string {

	params := make(map[string]string)

	if r.SearchIdentifierType == NAME {
		params["searchIdentifier"] = r.SearchIdentifier
		params["searchIdentifierType"] = "name"
	} else if r.SearchIdentifierType == ID {
		params["searchIdentifier"] = r.SearchIdentifier
		params["searchIdentifierType"] = "id"
	}

	if r.Query != "" {
		params["query"] = r.Query
	}

	return params
}

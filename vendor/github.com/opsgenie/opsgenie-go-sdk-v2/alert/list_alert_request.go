package alert

import (
	"net/http"
	"strconv"

	"github.com/opsgenie/opsgenie-go-sdk-v2/client"
)

type ListAlertRequest struct {
	client.BaseRequest
	Limit                int
	Sort                 SortField
	Offset               int
	Order                Order
	Query                string
	SearchIdentifier     string
	SearchIdentifierType SearchIdentifierType
}

func (r *ListAlertRequest) Validate() error {

	return nil
}

func (r *ListAlertRequest) ResourcePath() string {

	return "/v2/alerts"
}

func (r *ListAlertRequest) Method() string {
	return http.MethodGet
}

func (r *ListAlertRequest) RequestParams() map[string]string {

	params := make(map[string]string)

	if r.Limit != 0 {
		params["limit"] = strconv.Itoa(r.Limit)
	}

	if r.Sort != "" {
		params["sort"] = string(r.Sort)
	}

	if r.Offset != 0 {
		params["offset"] = strconv.Itoa(r.Offset)
	}

	if r.Query != "" {
		params["query"] = r.Query
	}

	if r.SearchIdentifier != "" {
		params["searchIdentifier"] = r.SearchIdentifier
	}

	if r.SearchIdentifierType != "" {
		params["searchIdentifierType"] = string(r.SearchIdentifierType)
	}

	if r.Order != "" {
		params["order"] = string(r.Order)
	}

	return params
}

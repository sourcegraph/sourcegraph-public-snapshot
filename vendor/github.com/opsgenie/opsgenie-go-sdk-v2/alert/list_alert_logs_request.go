package alert

import (
	"net/http"
	"strconv"

	"github.com/opsgenie/opsgenie-go-sdk-v2/client"
)

type ListAlertLogsRequest struct {
	client.BaseRequest
	IdentifierType  AlertIdentifier
	IdentifierValue string
	Offset          string
	Direction       RequestDirection
	Order           Order
	Limit           uint32
}

func (r *ListAlertLogsRequest) Validate() error {
	err := validateIdentifier(r.IdentifierValue)
	if err != nil {
		return err
	}
	return nil
}

func (r *ListAlertLogsRequest) ResourcePath() string {
	return "/v2/alerts/" + r.IdentifierValue + "/logs"
}

func (r *ListAlertLogsRequest) Method() string {
	return http.MethodGet
}

func (r *ListAlertLogsRequest) RequestParams() map[string]string {

	params := make(map[string]string)

	if r.IdentifierType == ALIAS {
		params["identifierType"] = "alias"

	} else if r.IdentifierType == TINYID {
		params["identifierType"] = "tiny"

	} else {
		params["identifierType"] = "id"
	}

	if r.Offset != "" {
		params["offset"] = r.Offset
	}

	if r.Order == Asc {
		params["order"] = "asc"
	} else if r.Order == Desc {
		params["order"] = "desc"
	}

	if r.Direction == NEXT {
		params["direction"] = "next"
	} else if r.Direction == PREV {
		params["direction"] = "prev"
	}

	if r.Limit != 0 {
		params["limit"] = strconv.Itoa(int(r.Limit))
	}

	return params
}

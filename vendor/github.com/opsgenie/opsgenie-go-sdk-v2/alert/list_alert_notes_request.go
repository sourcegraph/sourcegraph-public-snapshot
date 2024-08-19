package alert

import (
	"net/http"
	"strconv"

	"github.com/opsgenie/opsgenie-go-sdk-v2/client"
)

type ListAlertNotesRequest struct {
	client.BaseRequest
	IdentifierType  AlertIdentifier
	IdentifierValue string
	Offset          string
	Direction       RequestDirection
	Order           Order
	Limit           uint32
}

func (r *ListAlertNotesRequest) Validate() error {
	err := validateIdentifier(r.IdentifierValue)
	if err != nil {
		return err
	}
	return nil
}

func (r *ListAlertNotesRequest) ResourcePath() string {
	return "/v2/alerts/" + r.IdentifierValue + "/notes"
}

func (r *ListAlertNotesRequest) Method() string {
	return http.MethodGet
}

func (r *ListAlertNotesRequest) RequestParams() map[string]string {

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

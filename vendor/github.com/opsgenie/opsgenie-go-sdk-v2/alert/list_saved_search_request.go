package alert

import (
	"net/http"

	"github.com/opsgenie/opsgenie-go-sdk-v2/client"
)

type ListSavedSearchRequest struct {
	client.BaseRequest
}

func (r *ListSavedSearchRequest) Validate() error {

	return nil
}

func (r *ListSavedSearchRequest) ResourcePath() string {

	return "/v2/alerts/saved-searches"
}

func (r *ListSavedSearchRequest) Method() string {
	return http.MethodGet
}

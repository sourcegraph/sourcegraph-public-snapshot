package internal

import (
	"net/http"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
)

// Handler is a wrapper func for app HTTP handlers that enables app
// error pages.
func Handler(h func(http.ResponseWriter, *http.Request) error) http.Handler {
	return handlerutil.HandlerWithErrorReturn{
		Handler: h,
		Error:   HandleError,
	}
}

package internal

import (
	"net/http"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
)

// Handlers is a map of routes (by name) and their handlers. Its
// entries are added to the app's router.
//
// It should only be modified at init time. It is used to add routes
// that are handled in files that may be build-tag-disabled or that
// are in separate packages.
var Handlers = map[string]func(w http.ResponseWriter, r *http.Request) error{}

// Handler is a wrapper func for app HTTP handlers that enables app
// error pages.
func Handler(h func(http.ResponseWriter, *http.Request) error) http.Handler {
	return handlerutil.HandlerWithErrorReturn{
		Handler: h,
		Error:   HandleError,
	}
}

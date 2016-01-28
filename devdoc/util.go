package devdoc

import (
	"errors"
	"fmt"
	"net/http"

	"gopkg.in/inconshreveable/log15.v2"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
)

// handler is a wrapper func for app HTTP handlers.
func (a *App) handler(h func(http.ResponseWriter, *http.Request) error) http.Handler {
	return handlerutil.HandlerWithErrorReturn{
		Handler: h,
		Error:   a.handleError,
	}
}

// handleError renders the error template with the given error and status code.
func (a *App) handleError(w http.ResponseWriter, r *http.Request, status int, err error) {
	if status < 200 || status >= 500 {
		log15.Debug("devdoc.App.handleError called with unsuccessful status code", "method", r.Method, "request_uri", r.URL.RequestURI(), "status_code", status, "error", err.Error())
	}
	err2 := a.renderTemplate(w, r, "error.html", status, &struct {
		StatusCode int
		Status     string
		Err        error
		TemplateCommon
	}{
		StatusCode: status,
		Status:     http.StatusText(status),
		Err:        err,
	})
	if err2 != nil {
		log15.Error("error during execution of error template", "error", err2)
		err := fmt.Errorf("error during execution of error template: %v", err2)
		if !handlerutil.DebugMode(r) {
			err = errPublicFacingErrorMessage
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// errPublicFacingErrorMessage is the public facing error message to display
// when not in debug mode (to hide potentially sensitive information in original
// error message). Normally, an HTML template is used to do this, but if rendering that fails,
// this is the plain-text backup.
var errPublicFacingErrorMessage = errors.New(`Sorry, there’s been a problem.

If this issue persists, please post an issue in our tracker (https://src.sourcegraph.com/sourcegraph/.tracker/new)
with steps to reproduce and other useful context. Or, contact us by email (help@sourcegraph.com).`)

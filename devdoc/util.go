package devdoc

import (
	"log"
	"net/http"

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
		log.Printf("%s %s %d: %s", r.Method, r.URL.RequestURI(), status, err.Error())
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
		log.Printf("Error during execution of error template: %s.", err2)
	}
}

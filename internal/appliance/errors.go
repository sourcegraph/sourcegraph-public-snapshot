package appliance

import (
	"net/http"

	"github.com/sourcegraph/log"
)

const (
	queryKeyUserMessage      = "sourcegraph-appliance-user-message"
	errMsgSomethingWentWrong = "Something went wrong - please contact support."
)

func (a *Appliance) redirectToErrorPage(w http.ResponseWriter, req *http.Request, userMsg string, err error, userError bool) {
	a.redirectWithError(w, req, "/appliance/error", userMsg, err, userError)
}

func (a *Appliance) redirectWithError(w http.ResponseWriter, req *http.Request, path, userMsg string, err error, userError bool) {
	logFn := a.logger.Error
	if userError {
		logFn = a.logger.Info
	}
	logFn("an error occurred", log.Error(err))
	req = req.Clone(req.Context())
	req.URL.Path = path
	queryValues := req.URL.Query()
	queryValues.Set(queryKeyUserMessage, userMsg)
	req.URL.RawQuery = queryValues.Encode()
	http.Redirect(w, req, req.URL.String(), http.StatusFound)
}

func (a *Appliance) errorHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if err := renderTemplate("error", w, struct {
			Msg string
		}{
			Msg: req.URL.Query().Get(queryKeyUserMessage),
		}); err != nil {
			a.handleError(w, err, "executing template")
			return
		}
	})
}

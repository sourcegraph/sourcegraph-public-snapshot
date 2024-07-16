package appliance

import (
	"fmt"
	"net/http"

	"github.com/sourcegraph/log"
)

func (a *Appliance) logError(r *http.Request, err error) {
	a.logger.Error(err.Error(), log.String("method", r.Method), log.String("uri", r.URL.RequestURI()))
}

func (a *Appliance) errorResponse(w http.ResponseWriter, r *http.Request, status int, message any) {
	resp := responseData{"error": message}

	if err := a.writeJSON(w, status, resp, nil); err != nil {
		a.logError(r, err)
	}
}

func (a *Appliance) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	a.errorResponse(w, r, http.StatusBadRequest, err.Error())
}

func (a *Appliance) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	a.logError(r, err)
	a.errorResponse(w, r, http.StatusInternalServerError, "the server encountered a problem and could not process your request")
}

func (a *Appliance) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	a.errorResponse(w, r, http.StatusNotFound, "the requested resource could not be found")
}

func (a *Appliance) methodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
	a.errorResponse(w, r, http.StatusMethodNotAllowed, fmt.Sprintf("the %s method is not supported", r.Method))
}

func (a *Appliance) invalidAdminPasswordResponse(w http.ResponseWriter, r *http.Request) {
	a.errorResponse(w, r, http.StatusUnauthorized, "invalid admin password")
}

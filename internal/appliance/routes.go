package appliance

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (a *Appliance) Routes() *mux.Router {
	r := mux.NewRouter()
	r.Use(a.checkAuthorization)

	// ported appliance React UI endpoints
	r.NotFoundHandler = http.HandlerFunc(a.notFoundResponse)
	r.MethodNotAllowedHandler = http.HandlerFunc(a.methodNotAllowedResponse)

	r.Handle("/api/operator/v1beta1/stage", a.getStageJSONHandler())
	r.Handle("/api/operator/v1beta1/install/progress", a.getInstallProgressJSONHandler())
	r.Handle("/api/operator/v1beta1/maintenance/status", a.getStageJSONHandler())
	r.Handle("/api/operator/v1beta1/fake/stage", a.postStageJSONHandler())

	return r
}

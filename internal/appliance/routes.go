package appliance

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (a *Appliance) Routes() *mux.Router {
	r := mux.NewRouter()

	// ported appliance React UI endpoints
	r.NotFoundHandler = http.HandlerFunc(a.notFoundResponse)
	r.MethodNotAllowedHandler = http.HandlerFunc(a.methodNotAllowedResponse)

	r.Handle("/api/operator/v1beta1/stage", a.getSetupJSONHandler())
	r.Handle("/api/operator/v1beta1/install/progress", a.getInstallJSONHandler())
	r.Handle("/api/operator/v1beta1/maintenance/status", a.getStatusJSONHandler())
	r.Handle("/api/operator/v1beta1/fake/stage", a.postSetupJSONHandler())

	return r
}

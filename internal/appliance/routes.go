package appliance

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (a *Appliance) Routes() *mux.Router {
	r := mux.NewRouter()
	r.Use(a.checkAuthorization)

	// Route errors
	r.NotFoundHandler = http.HandlerFunc(a.notFoundResponse)
	r.MethodNotAllowedHandler = http.HandlerFunc(a.methodNotAllowedResponse)

	// Maintenance API URIs
	r.Handle("/api/v1/appliance/status", a.getStatusJSONHandler()).Methods("GET")
	r.Handle("/api/v1/appliance/status", a.postStatusJSONHandler()).Methods("POST")
	r.Handle("/api/v1/appliance/install/progress", a.getInstallProgressJSONHandler()).Methods("GET")
	r.Handle("/api/v1/appliance/maintenance/serviceStatuses", a.getMaintenanceStatusHandler()).Methods("GET")
	r.Handle("/api/v1/releases/sourcegraph", a.getReleasesHandler()).Methods("GET")

	return r
}

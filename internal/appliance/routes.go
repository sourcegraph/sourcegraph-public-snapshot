package appliance

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (a *Appliance) Routes() *mux.Router {
	r := mux.NewRouter()
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/appliance", http.StatusFound)
	})

	r.Handle("/appliance/login", a.getLoginHandler()).Methods(http.MethodGet)
	r.Handle("/appliance/login", a.postLoginHandler()).Methods(http.MethodPost)
	r.Handle("/appliance/error", a.errorHandler()).Methods(http.MethodGet)

	// Auth-gated endpoints
	r.Handle("/appliance", a.CheckAuthorization(a.applianceHandler())).Methods(http.MethodGet)
	r.Handle("/appliance/setup", a.CheckAuthorization(a.getSetupHandler())).Methods(http.MethodGet)
	r.Handle("/appliance/setup", a.CheckAuthorization(a.postSetupHandler())).Methods(http.MethodPost)

	return r
}

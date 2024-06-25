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

	// Auth-gated endpoints
	r.Handle("/appliance", a.checkAuthorization(a.applianceHandler())).Methods(http.MethodGet)
	r.Handle("/appliance/setup", a.checkAuthorization(a.getSetupHandler())).Methods(http.MethodGet)
	r.Handle("/appliance/setup", a.checkAuthorization(a.postSetupHandler())).Methods(http.MethodPost)

	return r
}

// TODO actually implement!
func (a *Appliance) checkAuthorization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		next.ServeHTTP(w, req)
	})
}

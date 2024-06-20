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
	r.HandleFunc("/appliance", a.applianceHandler).Methods(http.MethodGet)
	r.HandleFunc("/appliance/setup", a.getSetupHandler).Methods(http.MethodGet)
	r.HandleFunc("/appliance/setup", a.postSetupHandler).Methods(http.MethodPost)

	return r
}

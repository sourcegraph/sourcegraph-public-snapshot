package ui2

import (
	"net/http"

	"github.com/gorilla/mux"
)

func serveLegacySettingsTeam(w http.ResponseWriter, r *http.Request) error {
	http.Redirect(w, r, "/settings/orgs/"+mux.Vars(r)["org"], http.StatusMovedPermanently)
	return nil
}

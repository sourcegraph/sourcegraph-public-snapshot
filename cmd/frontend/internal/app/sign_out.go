package app

import (
	"net/http"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/session"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

func serveSignOut(w http.ResponseWriter, r *http.Request) {
	if err := session.DeleteSession(w, r); err != nil {
		log15.Error("Error deleting session during signout.", "err", err)
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

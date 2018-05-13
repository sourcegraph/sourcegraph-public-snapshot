package app

import (
	"net/http"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/session"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

func serveSignOut(w http.ResponseWriter, r *http.Request) {
	if err := session.SetActor(w, r, nil, 0); err != nil {
		log15.Error("Error in signout.", "err", err)
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

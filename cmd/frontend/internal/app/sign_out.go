package app

import (
	"net/http"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/session"
)

func serveSignOut(w http.ResponseWriter, r *http.Request) {
	// Invalidate all user sessions first
	// This way, any other signout failures should not leave a valid session
	if err := session.InvalidateSessionCurrentUser(r); err != nil {
		log15.Error("Error in signout.", "err", err)
	}
	if err := session.SetActor(w, r, nil, 0); err != nil {
		log15.Error("Error in signout.", "err", err)
	}

	http.Redirect(w, r, "/search", http.StatusSeeOther)
}

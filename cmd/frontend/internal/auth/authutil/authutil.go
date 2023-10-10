package authutil

import (
	"net/http"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/external/session"
	"github.com/sourcegraph/sourcegraph/internal/actor"
)

// ConnectOrSignOut checks whether or not the external account sign-in
// is a sign-in or an account connection attempt. If it a sign-in,
// it signs the user out by clearing their session.
func ConnectOrSignOut(w http.ResponseWriter, r *http.Request) error {
	isConnect := r.URL.Query().Get("connect") == "true"

	if !isConnect && actor.FromContext(r.Context()).IsAuthenticated() {
		err := session.SetActor(w, r, nil, 0, time.Time{})
		if err != nil {
			return err
		}
	}

	return nil
}

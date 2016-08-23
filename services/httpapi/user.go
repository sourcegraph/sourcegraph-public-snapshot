package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/routevar"
)

func serveUser(w http.ResponseWriter, r *http.Request) error {
	cl := handlerutil.Client(r)

	userSpec, err := routevar.ParseUserSpec(mux.Vars(r)["User"])
	if err != nil {
		return err
	}

	user, err := cl.Users.Get(r.Context(), &userSpec)
	if err != nil {
		return err
	}
	return writeJSON(w, user)
}

func serveUserEmails(w http.ResponseWriter, r *http.Request) error {
	cl := handlerutil.Client(r)

	userSpec, err := routevar.ParseUserSpec(mux.Vars(r)["User"])
	if err != nil {
		return err
	}

	emails, err := cl.Users.ListEmails(r.Context(), &userSpec)
	if err != nil {
		return err
	}
	return writeJSON(w, emails)
}

func serveBetaSubscription(w http.ResponseWriter, r *http.Request) error {
	var sub sourcegraph.BetaRegistration
	if err := json.NewDecoder(r.Body).Decode(&sub); err != nil {
		return err
	}

	cl := handlerutil.Client(r)
	resp, err := cl.Users.RegisterBeta(r.Context(), &sub)
	if err != nil {
		return err
	}
	return writeJSON(w, resp)
}

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
	ctx, cl := handlerutil.Client(r)

	userSpec, err := routevar.ParseUserSpec(mux.Vars(r)["User"])
	if err != nil {
		return err
	}

	user, err := cl.Users.Get(ctx, &userSpec)
	if err != nil {
		return err
	}
	return writeJSON(w, user)
}

func serveUserEmails(w http.ResponseWriter, r *http.Request) error {
	ctx, cl := handlerutil.Client(r)

	userSpec, err := routevar.ParseUserSpec(mux.Vars(r)["User"])
	if err != nil {
		return err
	}

	emails, err := cl.Users.ListEmails(ctx, &userSpec)
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

	ctx, cl := handlerutil.Client(r)
	resp, err := cl.Users.RegisterBeta(ctx, &sub)
	if err != nil {
		return err
	}
	return writeJSON(w, resp)
}

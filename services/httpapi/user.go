package httpapi

import (
	"encoding/json"
	"net/http"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
)

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

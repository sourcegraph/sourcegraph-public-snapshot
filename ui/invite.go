package ui

import (
	"encoding/json"
	"net/http"

	"fmt"

	"src.sourcegraph.com/sourcegraph/auth"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
)

func serveUserInvite(w http.ResponseWriter, r *http.Request) error {
	ctx := httpctx.FromRequest(r)
	cl := handlerutil.APIClient(r)

	ctxActor := auth.ActorFromContext(ctx)
	if !ctxActor.HasAdminAccess() { // current user is not an admin of the instance
		return fmt.Errorf("user not authenticated to complete this request")
	}

	query := struct {
		Email      string
		Permission string
	}{}
	if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
		return err
	}
	defer r.Body.Close()

	if query.Email == "" {
		return fmt.Errorf("no email specified")
	}

	var write, admin bool
	switch query.Permission {
	case "write":
		write = true
	case "admin":
		write = true
		admin = true
	case "read":
		// no-op
	default:
		return fmt.Errorf("unknown permission type")
	}

	pendingInvite, err := cl.Accounts.Invite(ctx, &sourcegraph.AccountInvite{
		Email: query.Email,
		Write: write,
		Admin: admin,
	})
	if err != nil {
		return err
	}
	return json.NewEncoder(w).Encode(pendingInvite)
}

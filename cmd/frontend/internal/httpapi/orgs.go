package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/orgs"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

func serveOrgs(w http.ResponseWriter, r *http.Request) error {
	decoder := json.NewDecoder(r.Body)
	var o sourcegraph.OrgListOptions
	err := decoder.Decode(&o)
	if err != nil {
		return err
	}

	orgs, err := orgs.ListOrgsPage(r.Context(), &o)
	if err != nil {
		return err
	}

	return writeJSON(w, orgs)
}

func serveOrgInvites(w http.ResponseWriter, r *http.Request) error {
	var i sourcegraph.UserInvite
	if err := json.NewDecoder(r.Body).Decode(&i); err != nil {
		return err
	}

	res, err := orgs.InviteUser(r.Context(), &i)
	if err != nil {
		return err
	}
	if res == sourcegraph.InviteMissingEmail {
		return errors.New("Missing invited user's email")
	}
	return writeJSON(w, "")
}

func serveOrgMembers(w http.ResponseWriter, r *http.Request) error {
	decoder := json.NewDecoder(r.Body)
	var m sourcegraph.OrgListOptions
	err := decoder.Decode(&m)
	if err != nil {
		return err
	}

	members, err := orgs.ListOrgMembersForInvites(r.Context(), &m)
	if err != nil {
		return err
	}

	return writeJSON(w, members)
}

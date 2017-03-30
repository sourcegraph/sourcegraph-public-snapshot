package httpapi

import (
	"encoding/json"
	"net/http"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/backend"
)

func serveOrgs(w http.ResponseWriter, r *http.Request) error {
	decoder := json.NewDecoder(r.Body)
	var o sourcegraph.OrgListOptions
	err := decoder.Decode(&o)
	if err != nil {
		return err
	}

	orgs, err := backend.Orgs.ListOrgsPage(r.Context(), &o)
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

	resp, err := backend.Orgs.InviteUser(r.Context(), &i)
	if err != nil {
		return err
	}
	return writeJSON(w, resp)
}

func serveOrgMembers(w http.ResponseWriter, r *http.Request) error {
	decoder := json.NewDecoder(r.Body)
	var m sourcegraph.OrgListOptions
	err := decoder.Decode(&m)
	if err != nil {
		return err
	}

	members, err := backend.Orgs.ListOrgMembersForInvites(r.Context(), &m)
	if err != nil {
		return err
	}

	return writeJSON(w, members)
}

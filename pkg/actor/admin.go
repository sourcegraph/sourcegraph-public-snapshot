package actor

import (
	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
)

var (
	adminUsernames = env.Get("ADMIN_USERNAMES", "", "space-delimited list of user usernames that should be treated as admins")
	adminUIDs      map[string]struct{}
)

func init() {
	adminUIDs = make(map[string]struct{})
	for _, username := range strings.Fields(adminUsernames) {
		adminUIDs[strings.ToLower(username)] = struct{}{}
	}
}

// IsAdmin returns true if and only if the actor should be treated as an instance admin.
func (a *Actor) IsAdmin() bool {
	if a == nil {
		return false
	}

	_, isAdmin := adminUIDs[a.Login]
	return isAdmin
}

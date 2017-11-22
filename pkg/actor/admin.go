package actor

import (
	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
)

var (
	adminAuthIDs = env.Get("ADMIN_AUTH_IDS", "", "space-delimited list of user Auth IDs that should be treated as admins")
	adminUIDs    map[string]struct{}
)

func init() {
	adminUIDs = make(map[string]struct{})
	for _, uid := range strings.Fields(adminAuthIDs) {
		adminUIDs[uid] = struct{}{}
	}
}

// IsAdmin returns true if and only if the actor should be treated as an instance admin.
func (a *Actor) IsAdmin() bool {
	if a == nil {
		return false
	}
	_, isAdmin := adminUIDs[a.UID]
	return isAdmin
}

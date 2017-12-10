package actor

import (
	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
)

var (
	adminUsernames = conf.Get().AdminUsernames
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
	_, isAdmin := adminUIDs[strings.ToLower(a.Login)]
	return isAdmin
}

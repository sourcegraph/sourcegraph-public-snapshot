package bg

import (
	"context"
	"strings"

	log15 "gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
)

// MigrateAdminUsernames updates the DB rows for the usernames listed in the adminUsernames
// config key to designate those users as admins. This removes the dependency on the deprecated
// adminUsernames config.
func MigrateAdminUsernames(ctx context.Context) {
	for _, username := range strings.Fields(conf.Get().AdminUsernames) {
		user, err := localstore.Users.GetByUsername(ctx, username)
		if err == nil {
			if err := localstore.Users.SetIsSiteAdmin(ctx, user.ID, true); err != nil {
				log15.Error("error updating user site-admin status (from deprecated adminUsernames config)", "user", user.Username, "userID", user.ID, "err", err)
			}
		}
	}
}

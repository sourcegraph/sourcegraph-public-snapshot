package bg

import (
	"context"
	"strings"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

// MigrateAdminUsernames updates the DB rows for the usernames listed in the adminUsernames
// config key to designate those users as admins. This removes the dependency on the deprecated
// adminUsernames config.
func MigrateAdminUsernames(ctx context.Context) {
	for _, username := range strings.Fields(conf.GetTODO().AdminUsernames) {
		user, err := db.Users.GetByUsername(ctx, username)
		if err == nil {
			if err := db.Users.SetIsSiteAdmin(ctx, user.ID, true); err != nil {
				log15.Error("error updating user site-admin status (from deprecated adminUsernames config)", "user", user.Username, "userID", user.ID, "err", err)
			}
		}
	}
}

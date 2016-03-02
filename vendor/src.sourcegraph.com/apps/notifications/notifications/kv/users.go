package kv

import (
	"html/template"

	"golang.org/x/net/context"
	"src.sourcegraph.com/apps/notifications/notifications"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

// TODO: This is duplicated from tracker, it should be factored out into a platform Users service.
func sgUser(ctx context.Context, user *sourcegraph.User) notifications.User {
	return notifications.User{
		UserSpec: notifications.UserSpec{
			ID:     uint64(user.UID),
			Domain: user.Domain, // TODO: If blank, set it to "sourcegraph.com"?
		},
		Login:     user.Login,
		AvatarURL: template.URL(user.AvatarURLOfSize(48 * 2)),
	}
}

package kv

import (
	"html/template"

	"golang.org/x/net/context"
	"src.sourcegraph.com/apps/tracker/issues"
	"src.sourcegraph.com/sourcegraph/conf"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

func sgUser(ctx context.Context, user *sourcegraph.User) issues.User {
	// TODO: Maybe sourcegraph Users service should have an extra field like ProfileURL?
	profile := *conf.AppURL(ctx)
	profile.Path = "~" + user.Login // TODO: Perhaps tap into sourcegraph routers, so this logic is DRY.

	return issues.User{
		UserSpec: issues.UserSpec{
			ID:     uint64(user.UID),
			Domain: user.Domain, // TODO: If blank, set it to "sourcegraph.com"?
		},
		Login:     user.Login,
		AvatarURL: template.URL(user.AvatarURLOfSize(48 * 2)),
		HTMLURL:   template.URL(profile.String()),
	}
}

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

	// TODO: Need to move this logic into Sourcegraph Users service and make it more complete/robust. For now, fall back to gravatar default image in case if no user avatar.
	avatarURL := template.URL(user.AvatarURL)
	if avatarURL == "" {
		avatarURL = "https://secure.gravatar.com/avatar?d=mm&f=y&s=96"
	}

	return issues.User{
		Login:     user.Login,
		AvatarURL: avatarURL,
		HTMLURL:   template.URL(profile.String()),
	}
}

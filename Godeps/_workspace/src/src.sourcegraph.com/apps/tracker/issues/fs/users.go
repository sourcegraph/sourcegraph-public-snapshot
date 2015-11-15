package fs

import (
	"html/template"

	"github.com/google/go-github/github"
	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/apps/tracker/issues"
	"src.sourcegraph.com/sourcegraph/conf"
)

func sgUser(ctx context.Context, user *sourcegraph.User) issues.User {
	// TODO: Maybe sourcegraph Users service should have an extra field like ProfileURL?
	profile := *conf.AppURL(ctx)
	profile.Path = "~" + user.Login // TODO: Perhaps tap into sourcegraph routers, so this logic is DRY.

	// TODO: Need to move this logic into Sourcegraph Users service and make it more complete/robust. For now, fall back to GitHub API in case if no user avatar.
	avatarURL := template.URL(user.AvatarURL)
	if avatarURL == "" {
		avatarURL = ghAvatarURL(user.Login)
	}

	return issues.User{
		Login:     user.Login,
		AvatarURL: avatarURL,
		HTMLURL:   template.URL(profile.String()),
	}
}

var (
	gh        = github.NewClient(nil)
	ghAvatars = make(map[string]template.URL) // ghAvatars is a cache of GitHub usernames to their avatar URLs.
)

func ghAvatarURL(login string) template.URL {
	if avatarURL, ok := ghAvatars[login]; ok {
		return avatarURL
	}

	user, _, err := gh.Users.Get(login)
	if err != nil || user.AvatarURL == nil {
		return ""
	}
	ghAvatars[login] = template.URL(*user.AvatarURL + "&s=96")
	return ghAvatars[login]
}

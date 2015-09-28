package local

import (
	"fmt"

	"strings"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	authpkg "src.sourcegraph.com/sourcegraph/auth"
	"src.sourcegraph.com/sourcegraph/ext/slack"
	"src.sourcegraph.com/sourcegraph/store"
)

// Repo config hooks are actions that are performed when a
// repo's config changes.

// changedConfig_Enabled is called when a repo's
// Enabled config value is changed.
func (s *repos) changedConfig_Enabled(ctx context.Context, repo *sourcegraph.Repo, enabled bool) error {
	if !enabled {
		return nil
	}

	login := "(anonymous)"
	if a := authpkg.ActorFromContext(ctx); a.IsAuthenticated() {
		u, err := store.UsersFromContext(ctx).Get(ctx, sourcegraph.UserSpec{UID: int32(a.UID)})
		if err == nil && u != nil {
			login = u.Login
		}
	}

	var details []string
	if repo.Language != "" {
		details = append(details, repo.Language)
	}
	if repo.GitHub != nil {
		details = append(details, fmt.Sprintf("%d ★", repo.GitHub.Stars))
	}
	go slack.PostMessage(slack.PostOpts{
		Msg:       fmt.Sprintf("<%s|%s> (%s) enabled by %s", repo.GitHubHTMLURL(), repo.URI, strings.Join(details, ","), login),
		Username:  "enable",
		IconEmoji: ":heavy_check_mark:",
	})

	// TODO(nodb-deploy): call RepoOrigin.AuthorizeSSHKey,
	// IsCommitStatusCapable, IsPushHookEnabled, etc.

	return nil
}

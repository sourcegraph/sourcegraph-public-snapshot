package backend

import (
	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/accesscontrol"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/internal/localstore"
	"sourcegraph.com/sourcegraph/sourcegraph/services/repoupdater"
)

func init() {
	accesscontrol.Repos = localstore.Repos
	repoupdater.Repos = Repos
	repoupdater.MirrorRepos = MirrorRepos
}

// LegacyGitHubScope returns the scope granted to the auth flow used before Auth0.
func LegacyGitHubScope(gitHubUID int) []string {
	appDB, _, _ := localstore.GlobalDBs()
	scope, err := appDB.SelectStr("SELECT scope FROM ext_auth_token WHERE ext_uid=$1;", gitHubUID)
	if err != nil || scope == "" {
		return nil
	}
	return strings.Split(scope, ",")
}

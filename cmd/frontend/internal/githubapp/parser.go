package githubapp

import (
	"strings"

	ghtypes "github.com/sourcegraph/sourcegraph/internal/github_apps/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func parseKind(s *string) (*ghtypes.GitHubAppKind, error) {
	if s == nil {
		return nil, nil
	}
	switch strings.ToUpper(*s) {
	case "SITE_CREDENTIAL":
		kind := ghtypes.SiteCredentialGitHubAppKind
		return &kind, nil
	case "USER_CREDENTIAL":
		kind := ghtypes.UserCredentialGitHubAppKind
		return &kind, nil
	case "COMMIT_SIGNING":
		kind := ghtypes.CommitSigningGitHubAppKind
		return &kind, nil
	case "REPO_SYNC":
		kind := ghtypes.RepoSyncGitHubAppKind
		return &kind, nil
	default:
		return nil, errors.Newf("unknown kind %q", *s)
	}
}

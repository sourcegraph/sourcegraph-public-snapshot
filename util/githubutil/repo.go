package githubutil

import (
	"fmt"
	"strings"

	"src.sourcegraph.com/sourcegraph/ext/github/githubcli"
)

// SplitGitHubRepoURI splits a string like "github.com/alice/myrepo" to "alice"
// and "myrepo".
func SplitGitHubRepoURI(uri string) (owner, repo string, err error) {
	// TODO(sqs): hack: treat sourcegraph.com/... as github.com/...
	uri = strings.Replace(uri, "sourcegraph.com/", "github.com/", 1)

	gitHubHost := githubcli.Config.Host() + "/"
	if strings.HasPrefix(uri, "github.com/") {
		uri = strings.TrimPrefix(uri, "github.com/")
	} else if strings.HasPrefix(uri, gitHubHost) {
		uri = strings.TrimPrefix(uri, gitHubHost)
	} else {
		return "", "", fmt.Errorf("not a GitHub repository URI: %q", uri)
	}

	parts := strings.Split(uri, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid GitHub repository owner/repo string: %q", uri)
	}

	owner, repo = parts[0], parts[1]
	if owner == "" || repo == "" {
		return "", "", fmt.Errorf("invalid GitHub owner or repo in %q", uri)
	}

	return owner, repo, nil
}

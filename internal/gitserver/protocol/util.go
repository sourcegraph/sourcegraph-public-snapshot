package protocol

import (
	"path"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/api"
)

func NormalizeRepo(input api.RepoName) api.RepoName {
	repo := string(input)

	// Clean with a "/" so we get out an absolute path
	repo = path.Clean("/" + repo)
	repo = strings.TrimPrefix(repo, "/")

	// This needs to be called after "path.Clean" because the host might be removed
	// e.g. github.com/../foo/bar
	host, repoPath := "", ""
	slash := strings.IndexByte(repo, '/')
	if slash == -1 {
		repoPath = repo
	} else {
		// host is always case-insensitive
		host, repoPath = strings.ToLower(repo[:slash]), repo[slash:]
	}

	trimGit := func(s string) string {
		s = strings.TrimSuffix(s, ".git")
		return strings.TrimSuffix(s, "/")
	}

	switch host {
	case "github.com":
		repoPath = trimGit(repoPath)

		// GitHub is fully case-insensitive.
		repoPath = strings.ToLower(repoPath)
	case "go":
		// support suffix ".git"
	default:
		repoPath = trimGit(repoPath)
	}

	return api.RepoName(host + repoPath)
}

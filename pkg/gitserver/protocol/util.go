package protocol

import (
	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

func NormalizeRepo(input api.RepoURI) api.RepoURI {
	repo := string(input)
	repo = strings.TrimSuffix(repo, ".git")

	slash := strings.IndexByte(repo, '/')
	if slash == -1 {
		return api.RepoURI(repo)
	}
	host := strings.ToLower(repo[:slash]) // host is always case insensitive
	path := repo[slash:]

	if host == "github.com" {
		return api.RepoURI(host + strings.ToLower(path)) // GitHub is fully case insensitive
	}

	return api.RepoURI(host + path) // other git hosts can be case sensitive on path
}

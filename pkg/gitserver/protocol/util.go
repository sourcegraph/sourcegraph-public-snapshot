package protocol

import "strings"

func NormalizeRepo(repo string) string {
	repo = strings.TrimSuffix(repo, ".git")

	slash := strings.IndexByte(repo, '/')
	if slash == -1 {
		return repo
	}
	host := strings.ToLower(repo[:slash]) // host is always case insensitive
	path := repo[slash:]

	if host == "github.com" {
		return host + strings.ToLower(path) // GitHub is fully case insensitive
	}

	return host + path // other git hosts can be case sensitive on path
}

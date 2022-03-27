package protocol

import (
	"path"
	"strings"
	"unicode/utf8"

	"github.com/sourcegraph/sourcegraph/internal/api"
)

func NormalizeRepo(input api.RepoName) api.RepoName {
	repo := string(input)
	repo = strings.TrimSuffix(repo, ".git")

	// Clean with a "/" so we get out an absolute path
	repo = path.Clean("/" + repo)
	repo = strings.TrimPrefix(repo, "/")

	// Check if we need to do lowercasing. If we don't we can avoid the
	// allocations we do later in the function.
	if !hasUpperASCII(repo) {
		return api.NewRepoName(repo)
	}

	slash := strings.IndexByte(repo, '/')
	if slash == -1 {
		return api.NewRepoName(repo)
	}
	host := strings.ToLower(repo[:slash]) // host is always case insensitive
	path := repo[slash:]

	if host == "github.com" {
		return api.NewRepoName(host + strings.ToLower(path)) // GitHub is fully case insensitive
	}

	return api.NewRepoName(host + path) // other git hosts can be case sensitive on path
}

// hasUpperASCII returns true if s contains any upper-case letters in ASCII,
// or if it contains any non-ascii characters.
func hasUpperASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= utf8.RuneSelf || (c >= 'A' && c <= 'Z') {
			return true
		}
	}
	return false
}

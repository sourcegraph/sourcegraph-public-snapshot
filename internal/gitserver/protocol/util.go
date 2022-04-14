package protocol

import (
	"path"
	"strings"
	"unicode/utf8"

	"github.com/sourcegraph/sourcegraph/internal/api"
)

func NormalizeRepo(input api.RepoName) api.RepoName {
	repo := string(input)

	host := ""
	firstSlash := strings.IndexByte(repo, '/')
	if firstSlash != -1 {
		host = strings.ToLower(repo[:firstSlash]) // host is always case-insensitive
	}

	repo = strings.TrimSuffix(repo, ".git")

	// Clean with a "/" so we get out an absolute path
	repo = path.Clean("/" + repo)
	repo = strings.TrimPrefix(repo, "/")

	// Check if we need to do lower-casing. If we don't we can avoid the
	// allocations we do later in the function.
	if !hasUpperASCII(repo) {
		return api.RepoName(repo)
	}

	if host == "" {
		return api.RepoName(repo)
	}

	repoPath := repo[firstSlash:]
	if host == "github.com" {
		return api.RepoName(host + strings.ToLower(repoPath)) // GitHub is fully case insensitive
	}

	return api.RepoName(host + repoPath) // other git hosts can be case sensitive on path
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

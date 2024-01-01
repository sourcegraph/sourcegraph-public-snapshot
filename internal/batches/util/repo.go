package util

import (
	"crypto/sha256"
	"encoding/base64"
	"strings"

	"github.com/sourcegraph/sourcegraph/lib/batches/template"
)

// NewTemplatingRepo transforms a given *graphql.Repository into a
// template.Repository.
func NewTemplatingRepo(repoName string, branch string, fileMatches map[string]bool) template.Repository {
	matches := make([]string, 0, len(fileMatches))
	for path := range fileMatches {
		matches = append(matches, path)
	}
	return template.Repository{
		Name:        repoName,
		Branch:      branch,
		FileMatches: matches,
	}
}

func SlugForPathInRepo(repoName, commit, path string) string {
	name := repoName
	if path != "" {
		// Since path can contain os.PathSeparator or other characters that
		// don't translate well between Windows and Unix systems, we hash it.
		hash := sha256.Sum256([]byte(path))
		name = name + "-" + base64.RawURLEncoding.EncodeToString(hash[:32])
	}
	return strings.ReplaceAll(name, "/", "-") + "-" + commit
}

func SlugForRepo(repoName, commit string) string {
	return strings.ReplaceAll(repoName, "/", "-") + "-" + commit
}

func EnsureRefPrefix(ref string) string {
	if strings.HasPrefix(ref, "refs/heads/") {
		return ref
	}
	return "refs/heads/" + ref
}

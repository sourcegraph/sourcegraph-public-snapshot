package repotrackutil

import (
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
)

var trackedRepo = []string{
	"github.com/kubernetes/kubernetes",
	"github.com/gorilla/mux",
	"github.com/golang/go",
	"sourcegraph/sourcegraph",
}
var trackedRepoRe = lazyregexp.New(`\b(` + strings.Join(trackedRepo, "|") + `)\b`)

// GetTrackedRepo guesses which repo a request URL path is for. It only looks
// at a certain subset of repos for its guess.
func GetTrackedRepo(repoPath api.RepoName) string {
	m := trackedRepoRe.FindStringSubmatch(string(repoPath))
	if len(m) == 0 {
		return "unknown"
	}
	return m[1]
}

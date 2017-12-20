package repotrackutil

import (
	"regexp"
	"strings"
)

var trackedRepo = []string{
	"github.com/kubernetes/kubernetes",
	"github.com/gorilla/mux",
	"github.com/golang/go",
	"sourcegraph/sourcegraph",
}
var trackedRepoRe = regexp.MustCompile(`\b(` + strings.Join(trackedRepo, "|") + `)\b`)

// GetTrackedRepo guesses which repo a request URL path is for. It only looks
// at a certain subset of repos for its guess.
func GetTrackedRepo(repoPath string) string {
	m := trackedRepoRe.FindStringSubmatch(repoPath)
	if len(m) == 0 {
		return "unknown"
	}
	return m[1]
}

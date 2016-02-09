package util

import (
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

// RepoURIHost returns the host of the given repoURI, converted to lower case, or empty string on error.
func RepoURIHost(repoURI string) string {
	u, err := url.Parse("//" + repoURI)
	if err != nil {
		return ""
	}
	return strings.ToLower(u.Host)
}

var trackedRepoRe = regexp.MustCompile(`/(github.com/kubernetes/kubernetes)\b`)

// getTrackedRepo guesses which repo a request is for. It only looks at a
// certain subset of repos for its guess.
func GetTrackedRepo(r *http.Request) string {
	m := trackedRepoRe.FindStringSubmatch(r.URL.Path)
	if len(m) == 0 {
		return "unknown"
	}
	return m[1]
}

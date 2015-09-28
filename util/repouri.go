package util

import (
	"net/url"
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

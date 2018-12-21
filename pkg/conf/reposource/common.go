package reposource

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/sourcegraph/sourcegraph/pkg/api"
)

// RepoSource is a wrapper around a repository source (typically a code host config) that provides a
// method to map clone URLs to repo names using only the configuration (i.e., no network requests).
type RepoSource interface {
	// cloneURLToRepoName maps a Git clone URL (format documented here:
	// https://git-scm.com/docs/git-clone#_git_urls_a_id_urls_a) to the expected repo name for the
	// repository on the code host.  It does not actually check if the repository exists in the code
	// host. It merely does the mapping based on the rules set in the code host config.
	//
	// If the clone URL does not correspond to a repository that could exist on the code host, the
	// empty string is returned and err is nil. If there is an unrelated error, an error is
	// returned.
	CloneURLToRepoName(cloneURL string) (repoName api.RepoName, err error)
}

// NormalizeBaseURL modifies the input and returns a normalized form of the base URL with insignificant
// differences (such as in presence of a trailing slash, or hostname case) eliminated. Its return value should be
// used for the (ExternalRepoSpec).ServiceID field (and passed to XyzExternalRepoSpec) instead of a non-normalized
// base URL.
func NormalizeBaseURL(baseURL *url.URL) *url.URL {
	baseURL.Host = strings.ToLower(baseURL.Host)
	if !strings.HasSuffix(baseURL.Path, "/") {
		baseURL.Path += "/"
	}
	return baseURL
}

var nonSCPURLRegex = regexp.MustCompile(`^(git\+)?(https?|ssh|rsync|file|git)://`)

// parseCloneURL parses a git clone URL into a URL struct. It supports the SCP-style git@host:path
// syntax that is common among code hosts.
func parseCloneURL(cloneURL string) (*url.URL, error) {
	if nonSCPURLRegex.MatchString(cloneURL) {
		return url.Parse(cloneURL)
	}

	// Support SCP-style syntax
	u, err := url.Parse("fake://" + strings.Replace(cloneURL, ":", "/", 1))
	if err != nil {
		return nil, err
	}
	u.Scheme = ""
	u.Path = strings.TrimPrefix(u.Path, "/")
	return u, nil
}

// hostname returns the hostname of a URL without www.
func hostname(url *url.URL) string {
	return strings.TrimPrefix(url.Hostname(), "www.")
}

// parseURLs parses the clone URL and repository host base URL into structs. It also returns a
// boolean indicating whether the hostnames of the URLs match.
func parseURLs(cloneURL, baseURL string) (parsedCloneURL, parsedBaseURL *url.URL, equalHosts bool, err error) {
	if baseURL != "" {
		parsedBaseURL, err = url.Parse(baseURL)
		if err != nil {
			return nil, nil, false, fmt.Errorf("Error parsing baseURL: %s", err)
		}
		parsedBaseURL = NormalizeBaseURL(parsedBaseURL)
	}

	parsedCloneURL, err = parseCloneURL(cloneURL)
	if err != nil {
		return nil, nil, false, fmt.Errorf("Error parsing cloneURL: %s", err)
	}
	hostsMatch := parsedBaseURL != nil && hostname(parsedBaseURL) == hostname(parsedCloneURL)
	return parsedCloneURL, parsedBaseURL, hostsMatch, nil
}

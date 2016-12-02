package graph

import (
	"errors"
	"fmt"
	"net/url"
	"path"
	"strings"
)

// MakeURI converts a repository clone URL, such as
// "git://github.com/user/repo.git", to a normalized URI string, such
// as "github.com/user/repo" lexically. MakeURI panics if there is an
// error, and should only be used if cloneURL is a correctly-formed
// URL. It is a wrapper around TryMakeURI.
func MakeURI(cloneURL string) string {
	uri, err := TryMakeURI(cloneURL)
	if err != nil {
		panic(err)
	}
	return uri
}

// TryMakeURI converts a repository clone URL, such as
// "git://github.com/user/repo.git", to a normalized URI string, such
// as "github.com/user/repo" lexically. TryMakeURI returns an error if
// cloneURL is empty or malformed.
//
// The following forms are supported:
// - transport://... (http://foo.bar)
// - vcs:transport://... (hg:http://foo.bar)
// - 'scm':vcs:transport://... (scm:git:git://foo.bar)
// - user@host:path (assumed SSH)
// - host:path (assumed SSH)
func TryMakeURI(cloneURL string) (string, error) {
	if cloneURL == "" {
		return "", errors.New("MakeURI: empty clone URL")
	}

	// Removing leading "scm:" if any
	cloneURL = strings.TrimPrefix(cloneURL, "scm:")

	// Removing VCS part if any, e.g., git:http://.. => http://..
	cloneURL = removeVCSPart(cloneURL)

	// Handle "user@host:path" and "host:path" (assumed SSH).
	if strings.Contains(cloneURL, ":") && !strings.Contains(cloneURL, "://") {
		cloneURL = "ssh://" + strings.Replace(cloneURL, ":", "/", -1)
	}

	url, err := url.Parse(cloneURL)
	if err != nil {
		return "", err
	} else if url.Host == "" && (url.Path[0] == '/' || !strings.Contains(strings.Trim(url.Path, "/"), "/")) {
		// We ensure our Path doesn't look like the output of TryMakeURI
		// so that the output of this function is a fixed point.
		// ie TryMakeURI("github.com/user/repo") == ("github.com/user/repo", nil),
		// not an error.
		return "", fmt.Errorf("determining URI from repo clone URL failed: missing host from URL (%q)", cloneURL)
	}

	uri := strings.TrimSuffix(url.Path, ".git")
	if uri != "" {
		uri = path.Clean(uri)
	}
	uri = strings.TrimSuffix(uri, "/")
	return strings.ToLower(url.Host) + uri, nil
}

// URIEqual returns true if a and b are equal, based on a case insensitive
// comparison (strings.EqualFold).
func URIEqual(a, b string) bool {
	return strings.EqualFold(a, b)
}

// removeVCSPart removes VCS part from URL if any, git:http://.. => http://..
func removeVCSPart(url string) string {
	parts := strings.SplitN(url, ":", 3)
	if len(parts) < 3 {
		// does not look like foo:bar:...
		return url
	}
	for _, r := range parts[0] {
		if !isASCIILetter(r) {
			return url
		}
	}
	for _, r := range parts[1] {
		if !isASCIILetter(r) {
			return url
		}
	}
	return parts[1] + ":" + parts[2]
}

// isASCIILetter reports if given rune is an ASCII letter.
func isASCIILetter(r rune) bool {
	return 'a' <= r && r <= 'z' || 'A' <= r && r <= 'Z'
}

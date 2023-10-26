package auth

import (
	"net/url"
	"path"
	"strings"
)

// SafeRedirectURL returns a safe redirect URL based on the input, to protect against open-redirect vulnerabilities.
//
// 🚨 SECURITY: Handlers MUST call this on any redirection destination URL derived from untrusted
// user input, or else there is a possible open-redirect vulnerability.
func SafeRedirectURL(urlStr string) string {
	u, err := url.Parse(urlStr)
	if err != nil || !strings.HasPrefix(u.Path, "/") {
		return "/"
	}

	// Make sure u.Path always starts with a single slash.
	u.Path = path.Clean(u.Path)

	// Only take certain known-safe fields.
	u = &url.URL{Path: u.Path, RawQuery: u.RawQuery}
	return u.String()
}

// Add a ?signup= or ?signin= parameter to a redirect URL.
func AddPostAuthRedirectParametersToURL(u *url.URL, newUserCreated bool, serviceType string) {
	q := u.Query()
	if newUserCreated {
		q.Add("signup", serviceType)
	} else {
		q.Add("signin", serviceType)
	}
	u.RawQuery = q.Encode()
}

func AddPostAuthRedirectParametersToString(urlStr string, newUserCreated bool, serviceType string) string {
	u, err := url.Parse(urlStr)
	if err != nil {
		return urlStr
	}
	AddPostAuthRedirectParametersToURL(u, newUserCreated, serviceType)
	return u.String()
}

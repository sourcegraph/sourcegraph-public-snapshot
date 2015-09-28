package oauth2util

import (
	"fmt"
	"net/url"
	"strings"
)

// RedirectURIMismatchError occurs when none of the registered
// client's redirect URIs matches the requested redirect URI.
type RedirectURIMismatchError struct {
	RequestedURI string // the user agent's requested redirect URI
}

func (e *RedirectURIMismatchError) Error() string {
	return fmt.Sprintf("OAuth2 redirect URI mismatch: %q does not match any registered redirect URIs", e.RequestedURI)
}

// RedirectURIInvalidError occurs when a redirect URI is invalid.
type RedirectURIInvalidError struct {
	RedirectURI string // the invalid redirect URI
}

func (e *RedirectURIInvalidError) Error() string {
	return fmt.Sprintf("invalid OAuth2 redirect URI: %q", e.RedirectURI)
}

// CheckRedirectURI checks a single redirect URI for validity. If the
// redirect URI is invalid, an error of type *RedirectURIInvalidError
// is returned.
func CheckRedirectURI(u *url.URL) error {
	if u.Scheme == "" || u.User != nil || u.Host == "" {
		return &RedirectURIInvalidError{u.String()}
	}
	if strings.HasSuffix(u.Path, "/..") || strings.Contains(u.Path, "/../") {
		return &RedirectURIInvalidError{u.String()}
	}
	if strings.HasSuffix(u.Path, "/.") || strings.Contains(u.Path, "/./") {
		return &RedirectURIInvalidError{u.String()}
	}
	return nil
}

// AllowRedirectURI checks whether to allow redirection to the
// requested OAuth redirect URI. At least one registered URI (which
// are registered in the client's record on the authorization server)
// must be a prefix of the requested URI.
//
// If redirection is DISALLOWED because there are no such prefix
// matches, an error of type *RedirectURIMismatchError is returned.
func AllowRedirectURI(registeredURIs []string, requestedURIStr string) error {
	requestedURI, err := url.Parse(requestedURIStr)
	if err != nil {
		return err
	}
	if err := CheckRedirectURI(requestedURI); err != nil {
		return err
	}

	for _, regURI := range registeredURIs {
		registeredURI, err := url.Parse(regURI)
		if err != nil {
			return err
		}
		if err := CheckRedirectURI(registeredURI); err != nil {
			return err
		}

		if redirectURIPrefixMatch(registeredURI, requestedURI) {
			return nil
		}
	}
	return &RedirectURIMismatchError{requestedURIStr}
}

func redirectURIPrefixMatch(registeredURI, requestedURI *url.URL) bool {
	if registeredURI.Host != requestedURI.Host || registeredURI.Scheme != requestedURI.Scheme {
		return false
	}

	if registeredURI.Path == "" {
		registeredURI.Path = "/"
	}
	if requestedURI.Path == "" {
		requestedURI.Path = "/"
	}
	if !strings.HasPrefix(requestedURI.Path, strings.TrimSuffix(registeredURI.Path, "/")+"/") && requestedURI.Path != registeredURI.Path {
		return false
	}

	return true
}

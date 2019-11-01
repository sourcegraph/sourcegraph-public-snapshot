package repos

import (
	"crypto/x509"
	"net/url"
	"strings"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/httputil"
)

// Deprecated: NormalizeBaseURL should not be used, instead use
// externalservice.NormalizeBaseURL
//
// NormalizeBaseURL modifies the input and returns a normalized form of the a base URL with insignificant
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

// NewHTTPClientFactory returns an httpcli.Factory with common
// options and middleware pre-set.
func NewHTTPClientFactory() *httpcli.Factory {
	return httpcli.NewFactory(
		// TODO(tsenart): Use middle for Prometheus instrumentation later.
		httpcli.NewMiddleware(
			httpcli.ContextErrorMiddleware,
		),
		httpcli.TracedTransportOpt,
		httpcli.NewCachedTransportOpt(httputil.Cache, true),
	)
}

// newCertPool returns an x509.CertPool with the given certificates added to it.
func newCertPool(certs ...string) (*x509.CertPool, error) {
	pool := x509.NewCertPool()
	for _, cert := range certs {
		if ok := pool.AppendCertsFromPEM([]byte(cert)); !ok {
			return nil, errors.New("invalid certificate")
		}
	}
	return pool, nil
}

// setUserinfoBestEffort adds the username and password to rawurl. If user is
// not set in rawurl, username is used. If password is not set and there is a
// user, password is used. If anything fails, the original rawurl is returned.
func setUserinfoBestEffort(rawurl, username, password string) string {
	u, err := url.Parse(rawurl)
	if err != nil {
		return rawurl
	}

	passwordSet := password != ""

	// Update username and password if specified in rawurl
	if u.User != nil {
		if u.User.Username() != "" {
			username = u.User.Username()
		}
		if p, ok := u.User.Password(); ok {
			password = p
			passwordSet = true
		}
	}

	if username == "" {
		return rawurl
	}

	if passwordSet {
		u.User = url.UserPassword(username, password)
	} else {
		u.User = url.User(username)
	}

	return u.String()
}

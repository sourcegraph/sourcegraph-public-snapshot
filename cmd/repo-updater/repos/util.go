package repos

import (
	"crypto/x509"
	"net/url"
	"strings"

	"github.com/sourcegraph/sourcegraph/pkg/httpcli"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/httputil"
)

// NormalizeBaseURL modifies the input and returns a normalized form of the a base URL with insignificant
// differences (such as in presence of a trailing slash, or hostname case) eliminated. Its return value should be
// used for the (ExternalRepoSpec).ServiceID field (and passed to XyzExternalRepoSpec) instead of a non-normalized
// base URL.
//
// DEPRECATED in favor of externalservice.NormalizeBaseURL
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

// random will create a file of size bytes (rounded up to next 1024 size)
func random_510(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		

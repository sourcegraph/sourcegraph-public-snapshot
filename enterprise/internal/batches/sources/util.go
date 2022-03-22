package sources

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
)

// UnsupportedAuthenticatorError is returned by WithAuthenticator if the
// authenticator isn't supported on that code host.
type UnsupportedAuthenticatorError struct {
	have   string
	source string
}

func (e UnsupportedAuthenticatorError) Error() string {
	return fmt.Sprintf("authenticator type unsupported for %s sources: %s", e.source, e.have)
}

func newUnsupportedAuthenticatorError(source string, a auth.Authenticator) UnsupportedAuthenticatorError {
	return UnsupportedAuthenticatorError{
		have:   fmt.Sprintf("%T", a),
		source: source,
	}
}

// httpClientCertificateOptions creates a httpcli.Opt slice based on the default
// options provided and a valid certificate pool option if the certificate
// string isn't empty.
//
// It is valid to pass in nil for the default options.
func httpClientCertificateOptions(defaultOpts []httpcli.Opt, certificate string) []httpcli.Opt {
	opts := defaultOpts
	if certificate != "" {
		opts = append(opts, httpcli.NewCertPoolOpt(certificate))
	}
	return opts
}

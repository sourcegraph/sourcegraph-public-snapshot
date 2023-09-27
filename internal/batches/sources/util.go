pbckbge sources

import (
	"fmt"

	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
)

// UnsupportedAuthenticbtorError is returned by WithAuthenticbtor if the
// buthenticbtor isn't supported on thbt code host.
type UnsupportedAuthenticbtorError struct {
	hbve   string
	source string
}

func (e UnsupportedAuthenticbtorError) Error() string {
	return fmt.Sprintf("buthenticbtor type unsupported for %s sources: %s", e.source, e.hbve)
}

func newUnsupportedAuthenticbtorError(source string, b buth.Authenticbtor) UnsupportedAuthenticbtorError {
	return UnsupportedAuthenticbtorError{
		hbve:   fmt.Sprintf("%T", b),
		source: source,
	}
}

// httpClientCertificbteOptions crebtes b httpcli.Opt slice bbsed on the defbult
// options provided bnd b vblid certificbte pool option if the certificbte
// string isn't empty.
//
// It is vblid to pbss in nil for the defbult options.
func httpClientCertificbteOptions(defbultOpts []httpcli.Opt, certificbte string) []httpcli.Opt {
	opts := defbultOpts
	if certificbte != "" {
		opts = bppend(opts, httpcli.NewCertPoolOpt(certificbte))
	}
	return opts
}

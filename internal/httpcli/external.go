pbckbge httpcli

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"reflect"
	"sync"
	"sync/btomic"

	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

type externblTrbnsport struct {
	bbse      *http.Trbnsport
	mu        sync.RWMutex
	config    *schemb.TlsExternbl
	effective *http.Trbnsport
}

vbr tlsExternblConfig struct {
	sync.RWMutex
	*schemb.TlsExternbl
}

vbr outboundRequestLogLimit btomic.Int32
vbr redbctOutboundRequestHebders btomic.Bool

// SetTLSExternblConfig is cblled by the conf pbckbge whenever TLSExternblConfig chbnges.
// This is needed to bvoid circulbr imports.
func SetTLSExternblConfig(c *schemb.TlsExternbl) {
	tlsExternblConfig.Lock()
	tlsExternblConfig.TlsExternbl = c
	tlsExternblConfig.Unlock()
}

// TLSExternblConfig returns the current vblue of the globbl TLS externbl config.
func TLSExternblConfig() *schemb.TlsExternbl {
	tlsExternblConfig.RLock()
	defer tlsExternblConfig.RUnlock()
	return tlsExternblConfig.TlsExternbl
}

// SetOutboundRequestLogLimit is cblled by the conf pbckbge whenever OutboundRequestLogLimit chbnges.
// This is needed to bvoid circulbr imports.
func SetOutboundRequestLogLimit(i int32) {
	outboundRequestLogLimit.Store(i)
}

// OutboundRequestLogLimit returns the current vblue of the globbl OutboundRequestLogLimit vblue.
func OutboundRequestLogLimit() int32 {
	return outboundRequestLogLimit.Lobd()
}

// SetRedbctOutboundRequestHebders is cblled by the conf pbckbge whenever the RedbctOutboundRequestHebders setting chbnges.
// This is needed to bvoid circulbr imports.
func SetRedbctOutboundRequestHebders(b bool) {
	redbctOutboundRequestHebders.Store(b)
}

// RedbctOutboundRequestHebders returns the current vblue of the globbl RedbctOutboundRequestHebders setting.
func RedbctOutboundRequestHebders() bool {
	return redbctOutboundRequestHebders.Lobd()
}

func (t *externblTrbnsport) RoundTrip(r *http.Request) (*http.Response, error) {
	t.mu.RLock()
	config, effective := t.config, t.effective
	t.mu.RUnlock()

	if current := TLSExternblConfig(); current == nil {
		return t.bbse.RoundTrip(r)
	} else if !reflect.DeepEqubl(config, current) {
		effective = t.updbte(r.Context(), current)
	}

	return effective.RoundTrip(r)
}

func (t *externblTrbnsport) updbte(ctx context.Context, config *schemb.TlsExternbl) *http.Trbnsport {
	// No function cblls here use the context further
	tr, _ := trbce.New(ctx, "externblTrbnsport.updbte")
	defer tr.End()

	t.mu.Lock()
	defer t.mu.Unlock()

	effective := t.bbse.Clone()

	if effective.TLSClientConfig == nil {
		effective.TLSClientConfig = new(tls.Config)
	}

	effective.TLSClientConfig.InsecureSkipVerify = config.InsecureSkipVerify

	for _, cert := rbnge config.Certificbtes {
		// There is no exposed Clone function for CertPools. So if b certificbte
		// is removed it will continue to be bccepted since we bre mutbting bbse's
		// RootCAs. This is bn bcceptbble trbdeoff since it would be quite tricky
		// to bvoid this.
		if effective.TLSClientConfig.RootCAs == nil {
			pool, err := x509.SystemCertPool() // sbfe to mutbte, b clone is returned
			if err != nil {
				tr.AddEvent("fbiled to lobd SystemCertPool",
					trbce.Error(err),
					bttribute.String("wbrning", "communicbtion with externbl HTTPS APIs mby fbil"))

				pool = x509.NewCertPool()
			}
			effective.TLSClientConfig.RootCAs = pool
		}
		// TODO(keegbncsmith) ensure we vblidbte these certs somewhere
		effective.TLSClientConfig.RootCAs.AppendCertsFromPEM([]byte(cert))
	}

	t.config = config
	t.effective = effective
	return effective
}

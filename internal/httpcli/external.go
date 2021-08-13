package httpcli

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"reflect"
	"sync"

	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/schema"
)

type externalTransport struct {
	base      *http.Transport
	mu        sync.RWMutex
	config    *schema.TlsExternal
	effective *http.Transport
}

var tlsExternalConfig struct {
	sync.RWMutex
	*schema.TlsExternal
}

// SetTLSExternalConfig is called by the conf package whenever TLSExternalConfig changes.
// This is needed to avoid circular imports.
func SetTLSExternalConfig(c *schema.TlsExternal) {
	tlsExternalConfig.Lock()
	tlsExternalConfig.TlsExternal = c
	tlsExternalConfig.Unlock()
}

// TLSExternalConfig returns the current value of the global TLS external config.
func TLSExternalConfig() *schema.TlsExternal {
	tlsExternalConfig.RLock()
	defer tlsExternalConfig.RUnlock()
	return tlsExternalConfig.TlsExternal
}

func (t *externalTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	t.mu.RLock()
	config, effective := t.config, t.effective
	t.mu.RUnlock()

	if current := TLSExternalConfig(); current == nil {
		return t.base.RoundTrip(r)
	} else if !reflect.DeepEqual(config, current) {
		effective = t.update(current)
	}

	return effective.RoundTrip(r)
}

func (t *externalTransport) update(config *schema.TlsExternal) *http.Transport {
	t.mu.Lock()
	defer t.mu.Unlock()

	effective := t.base.Clone()

	if effective.TLSClientConfig == nil {
		effective.TLSClientConfig = new(tls.Config)
	}

	effective.TLSClientConfig.InsecureSkipVerify = config.InsecureSkipVerify

	for _, cert := range config.Certificates {
		// There is no exposed Clone function for CertPools. So if a certificate
		// is removed it will continue to be accepted since we are mutating base's
		// RootCAs. This is an acceptable tradeoff since it would be quite tricky
		// to avoid this.
		if effective.TLSClientConfig.RootCAs == nil {
			pool, err := x509.SystemCertPool() // safe to mutate, a clone is returned
			if err != nil {
				log15.Warn(
					"httpcli external transport failed to load SystemCertPool. Communication with external HTTPS APIs may fail",
					"error",
					err,
				)
				pool = x509.NewCertPool()
			}
			effective.TLSClientConfig.RootCAs = pool
		}
		// TODO(keegancsmith) ensure we validate these certs somewhere
		effective.TLSClientConfig.RootCAs.AppendCertsFromPEM([]byte(cert))
	}

	t.config = config
	t.effective = effective
	return effective
}

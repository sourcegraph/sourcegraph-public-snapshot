package httpcli

import (
	"crypto/x509"

	"github.com/pkg/errors"
)

// NewCertPool returns an x509.CertPool with the given certificates added to it.
func NewCertPool(certs ...string) (*x509.CertPool, error) {
	pool := x509.NewCertPool()
	for _, cert := range certs {
		if ok := pool.AppendCertsFromPEM([]byte(cert)); !ok {
			return nil, errors.New("invalid certificate")
		}
	}
	return pool, nil
}

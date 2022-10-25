// package cacert is a subset of the stdlib x509 package, but including code
// to expose the raw system certificates on linux.
package cacert

import "sync"

var (
	systemOnce  sync.Once
	systemCerts [][]byte
	systemErr   error
)

// System returns PEM encoded system certificates. Note: This function only
// works on Linux. Other operating systems do not rely on PEM files at known
// locations, instead they rely on system calls.
func System() ([][]byte, error) {
	systemOnce.Do(func() {
		c, err := loadSystemRoots()
		if err != nil {
			systemErr = err
			return
		}
		systemCerts = c.certs
	})
	return systemCerts, systemErr
}

// CertPool exists for interaction with x509. Do not use.
type CertPool struct {
	certs [][]byte
}

// CertPool exists for interaction with x509. Do not use.
func NewCertPool() *CertPool {
	return &CertPool{}
}

func (c *CertPool) AppendCertsFromPEM(data []byte) {
	c.certs = append(c.certs, data)
}

func (c *CertPool) len() int {
	return len(c.certs)
}

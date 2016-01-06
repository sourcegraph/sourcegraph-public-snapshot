package exec

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"
	"path/filepath"
)

// TODO(sqs!native-ci): remove when https://github.com/samalba/dockerclient/pull/201 and https://github.com/drone/drone-exec/pull/13 are merged.
func TLSConfigFromCertPath(path string) (*tls.Config, error) {
	cert, err := ioutil.ReadFile(filepath.Join(path, "cert.pem"))
	if err != nil {
		return nil, err
	}
	key, err := ioutil.ReadFile(filepath.Join(path, "key.pem"))
	if err != nil {
		return nil, err
	}
	ca, err := ioutil.ReadFile(filepath.Join(path, "ca.pem"))
	if err != nil {
		return nil, err
	}
	tlsCert, err := tls.X509KeyPair(cert, key)
	if err != nil {
		return nil, err
	}
	tlsConfig := &tls.Config{Certificates: []tls.Certificate{tlsCert}}
	tlsConfig.RootCAs = x509.NewCertPool()
	if !tlsConfig.RootCAs.AppendCertsFromPEM(ca) {
		return nil, errors.New("Could not add RootCA pem")
	}
	return tlsConfig, nil
}

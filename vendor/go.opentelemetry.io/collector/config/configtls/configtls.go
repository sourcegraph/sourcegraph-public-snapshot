// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package configtls // import "go.opentelemetry.io/collector/config/configtls"

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.opentelemetry.io/collector/config/configopaque"
)

// We should avoid that users unknowingly use a vulnerable TLS version.
// The defaults should be a safe configuration
const defaultMinTLSVersion = tls.VersionTLS12

// Uses the default MaxVersion from "crypto/tls" which is the maximum supported version
const defaultMaxTLSVersion = 0

var systemCertPool = x509.SystemCertPool

// Config exposes the common client and server TLS configurations.
// Note: Since there isn't anything specific to a server connection. Components
// with server connections should use Config.
type Config struct {
	// Path to the CA cert. For a client this verifies the server certificate.
	// For a server this verifies client certificates. If empty uses system root CA.
	// (optional)
	CAFile string `mapstructure:"ca_file"`

	// In memory PEM encoded cert. (optional)
	CAPem configopaque.String `mapstructure:"ca_pem"`

	// If true, load system CA certificates pool in addition to the certificates
	// configured in this struct.
	IncludeSystemCACertsPool bool `mapstructure:"include_system_ca_certs_pool"`

	// Path to the TLS cert to use for TLS required connections. (optional)
	CertFile string `mapstructure:"cert_file"`

	// In memory PEM encoded TLS cert to use for TLS required connections. (optional)
	CertPem configopaque.String `mapstructure:"cert_pem"`

	// Path to the TLS key to use for TLS required connections. (optional)
	KeyFile string `mapstructure:"key_file"`

	// In memory PEM encoded TLS key to use for TLS required connections. (optional)
	KeyPem configopaque.String `mapstructure:"key_pem"`

	// MinVersion sets the minimum TLS version that is acceptable.
	// If not set, TLS 1.2 will be used. (optional)
	MinVersion string `mapstructure:"min_version"`

	// MaxVersion sets the maximum TLS version that is acceptable.
	// If not set, refer to crypto/tls for defaults. (optional)
	MaxVersion string `mapstructure:"max_version"`

	// CipherSuites is a list of TLS cipher suites that the TLS transport can use.
	// If left blank, a safe default list is used.
	// See https://go.dev/src/crypto/tls/cipher_suites.go for a list of supported cipher suites.
	CipherSuites []string `mapstructure:"cipher_suites"`

	// ReloadInterval specifies the duration after which the certificate will be reloaded
	// If not set, it will never be reloaded (optional)
	ReloadInterval time.Duration `mapstructure:"reload_interval"`
}

// NewDefaultConfig creates a new TLSSetting with any default values set.
func NewDefaultConfig() Config {
	return Config{}
}

// ClientConfig contains TLS configurations that are specific to client
// connections in addition to the common configurations. This should be used by
// components configuring TLS client connections.
type ClientConfig struct {
	// squash ensures fields are correctly decoded in embedded struct.
	Config `mapstructure:",squash"`

	// These are config options specific to client connections.

	// In gRPC when set to true, this is used to disable the client transport security.
	// See https://godoc.org/google.golang.org/grpc#WithInsecure.
	// In HTTP, this disables verifying the server's certificate chain and host name
	// (InsecureSkipVerify in the tls Config). Please refer to
	// https://godoc.org/crypto/tls#Config for more information.
	// (optional, default false)
	Insecure bool `mapstructure:"insecure"`
	// InsecureSkipVerify will enable TLS but not verify the certificate.
	InsecureSkipVerify bool `mapstructure:"insecure_skip_verify"`
	// ServerName requested by client for virtual hosting.
	// This sets the ServerName in the TLSConfig. Please refer to
	// https://godoc.org/crypto/tls#Config for more information. (optional)
	ServerName string `mapstructure:"server_name_override"`
}

// NewDefaultClientConfig creates a new TLSClientSetting with any default values set.
func NewDefaultClientConfig() ClientConfig {
	return ClientConfig{
		Config: NewDefaultConfig(),
	}
}

// ServerConfig contains TLS configurations that are specific to server
// connections in addition to the common configurations. This should be used by
// components configuring TLS server connections.
type ServerConfig struct {
	// squash ensures fields are correctly decoded in embedded struct.
	Config `mapstructure:",squash"`

	// These are config options specific to server connections.

	// Path to the TLS cert to use by the server to verify a client certificate. (optional)
	// This sets the ClientCAs and ClientAuth to RequireAndVerifyClientCert in the TLSConfig. Please refer to
	// https://godoc.org/crypto/tls#Config for more information. (optional)
	ClientCAFile string `mapstructure:"client_ca_file"`

	// Reload the ClientCAs file when it is modified
	// (optional, default false)
	ReloadClientCAFile bool `mapstructure:"client_ca_file_reload"`
}

// NewDefaultServerConfig creates a new TLSServerSetting with any default values set.
func NewDefaultServerConfig() ServerConfig {
	return ServerConfig{
		Config: NewDefaultConfig(),
	}
}

// certReloader is a wrapper object for certificate reloading
// Its GetCertificate method will either return the current certificate or reload from disk
// if the last reload happened more than ReloadInterval ago
type certReloader struct {
	nextReload time.Time
	cert       *tls.Certificate
	lock       sync.RWMutex
	tls        Config
}

func (c Config) newCertReloader() (*certReloader, error) {
	cert, err := c.loadCertificate()
	if err != nil {
		return nil, err
	}
	return &certReloader{
		tls:        c,
		nextReload: time.Now().Add(c.ReloadInterval),
		cert:       &cert,
	}, nil
}

func (r *certReloader) GetCertificate() (*tls.Certificate, error) {
	now := time.Now()
	// Read locking here before we do the time comparison
	// If a reload is in progress this will block and we will skip reloading in the current
	// call once we can continue
	r.lock.RLock()
	if r.tls.ReloadInterval != 0 && r.nextReload.Before(now) && (r.tls.hasCertFile() || r.tls.hasKeyFile()) {
		// Need to release the read lock, otherwise we deadlock
		r.lock.RUnlock()
		r.lock.Lock()
		defer r.lock.Unlock()
		cert, err := r.tls.loadCertificate()
		if err != nil {
			return nil, fmt.Errorf("failed to load TLS cert and key: %w", err)
		}
		r.cert = &cert
		r.nextReload = now.Add(r.tls.ReloadInterval)
		return r.cert, nil
	}
	defer r.lock.RUnlock()
	return r.cert, nil
}

func (c Config) Validate() error {
	if c.hasCAFile() && c.hasCAPem() {
		return fmt.Errorf("provide either a CA file or the PEM-encoded string, but not both")
	}

	minTLS, err := convertVersion(c.MinVersion, defaultMinTLSVersion)
	if err != nil {
		return fmt.Errorf("invalid TLS min_version: %w", err)
	}

	maxTLS, err := convertVersion(c.MaxVersion, defaultMaxTLSVersion)
	if err != nil {
		return fmt.Errorf("invalid TLS max_version: %w", err)
	}

	if maxTLS < minTLS && maxTLS != defaultMaxTLSVersion {
		return errors.New("invalid TLS configuration: min_version cannot be greater than max_version")
	}

	return nil
}

// loadTLSConfig loads TLS certificates and returns a tls.Config.
// This will set the RootCAs and Certificates of a tls.Config.
func (c Config) loadTLSConfig() (*tls.Config, error) {
	certPool, err := c.loadCACertPool()
	if err != nil {
		return nil, err
	}

	var getCertificate func(*tls.ClientHelloInfo) (*tls.Certificate, error)
	var getClientCertificate func(*tls.CertificateRequestInfo) (*tls.Certificate, error)
	if c.hasCert() || c.hasKey() {
		var certReloader *certReloader
		certReloader, err = c.newCertReloader()
		if err != nil {
			return nil, fmt.Errorf("failed to load TLS cert and key: %w", err)
		}
		getCertificate = func(*tls.ClientHelloInfo) (*tls.Certificate, error) { return certReloader.GetCertificate() }
		getClientCertificate = func(*tls.CertificateRequestInfo) (*tls.Certificate, error) { return certReloader.GetCertificate() }
	}

	minTLS, err := convertVersion(c.MinVersion, defaultMinTLSVersion)
	if err != nil {
		return nil, fmt.Errorf("invalid TLS min_version: %w", err)
	}
	maxTLS, err := convertVersion(c.MaxVersion, defaultMaxTLSVersion)
	if err != nil {
		return nil, fmt.Errorf("invalid TLS max_version: %w", err)
	}
	cipherSuites, err := convertCipherSuites(c.CipherSuites)
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		RootCAs:              certPool,
		GetCertificate:       getCertificate,
		GetClientCertificate: getClientCertificate,
		MinVersion:           minTLS,
		MaxVersion:           maxTLS,
		CipherSuites:         cipherSuites,
	}, nil
}

func convertCipherSuites(cipherSuites []string) ([]uint16, error) {
	var result []uint16
	var errs []error
	for _, suite := range cipherSuites {
		found := false
		for _, supported := range tls.CipherSuites() {
			if suite == supported.Name {
				result = append(result, supported.ID)
				found = true
				break
			}
		}
		if !found {
			errs = append(errs, fmt.Errorf("invalid TLS cipher suite: %q", suite))
		}
	}
	return result, errors.Join(errs...)
}

func (c Config) loadCACertPool() (*x509.CertPool, error) {
	// There is no need to load the System Certs for RootCAs because
	// if the value is nil, it will default to checking against th System Certs.
	var err error
	var certPool *x509.CertPool

	switch {
	case c.hasCAFile() && c.hasCAPem():
		return nil, fmt.Errorf("failed to load CA CertPool: provide either a CA file or the PEM-encoded string, but not both")
	case c.hasCAFile():
		// Set up user specified truststore from file
		certPool, err = c.loadCertFile(c.CAFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load CA CertPool File: %w", err)
		}
	case c.hasCAPem():
		// Set up user specified truststore from PEM
		certPool, err = c.loadCertPem([]byte(c.CAPem))
		if err != nil {
			return nil, fmt.Errorf("failed to load CA CertPool PEM: %w", err)
		}
	}

	return certPool, nil
}

func (c Config) loadCertFile(certPath string) (*x509.CertPool, error) {
	certPem, err := os.ReadFile(filepath.Clean(certPath))
	if err != nil {
		return nil, fmt.Errorf("failed to load cert %s: %w", certPath, err)
	}

	return c.loadCertPem(certPem)
}

func (c Config) loadCertPem(certPem []byte) (*x509.CertPool, error) {
	certPool := x509.NewCertPool()
	if c.IncludeSystemCACertsPool {
		scp, err := systemCertPool()
		if err != nil {
			return nil, err
		}
		if scp != nil {
			certPool = scp
		}
	}
	if !certPool.AppendCertsFromPEM(certPem) {
		return nil, fmt.Errorf("failed to parse cert")
	}
	return certPool, nil
}

func (c Config) loadCertificate() (tls.Certificate, error) {
	switch {
	case c.hasCert() != c.hasKey():
		return tls.Certificate{}, fmt.Errorf("for auth via TLS, provide both certificate and key, or neither")
	case !c.hasCert() && !c.hasKey():
		return tls.Certificate{}, nil
	case c.hasCertFile() && c.hasCertPem():
		return tls.Certificate{}, fmt.Errorf("for auth via TLS, provide either a certificate or the PEM-encoded string, but not both")
	case c.hasKeyFile() && c.hasKeyPem():
		return tls.Certificate{}, fmt.Errorf("for auth via TLS, provide either a key or the PEM-encoded string, but not both")
	}

	var certPem, keyPem []byte
	var err error
	if c.hasCertFile() {
		certPem, err = os.ReadFile(c.CertFile)
		if err != nil {
			return tls.Certificate{}, err
		}
	} else {
		certPem = []byte(c.CertPem)
	}

	if c.hasKeyFile() {
		keyPem, err = os.ReadFile(c.KeyFile)
		if err != nil {
			return tls.Certificate{}, err
		}
	} else {
		keyPem = []byte(c.KeyPem)
	}

	certificate, err := tls.X509KeyPair(certPem, keyPem)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to load TLS cert and key PEMs: %w", err)
	}

	return certificate, err
}

func (c Config) loadCert(caPath string) (*x509.CertPool, error) {
	caPEM, err := os.ReadFile(filepath.Clean(caPath))
	if err != nil {
		return nil, fmt.Errorf("failed to load CA %s: %w", caPath, err)
	}

	var certPool *x509.CertPool
	if c.IncludeSystemCACertsPool {
		if certPool, err = systemCertPool(); err != nil {
			return nil, err
		}
	}
	if certPool == nil {
		certPool = x509.NewCertPool()
	}
	if !certPool.AppendCertsFromPEM(caPEM) {
		return nil, fmt.Errorf("failed to parse CA %s", caPath)
	}
	return certPool, nil
}

// LoadTLSConfig loads the TLS configuration.
func (c ClientConfig) LoadTLSConfig(_ context.Context) (*tls.Config, error) {
	if c.Insecure && !c.hasCA() {
		return nil, nil
	}

	tlsCfg, err := c.loadTLSConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS config: %w", err)
	}
	tlsCfg.ServerName = c.ServerName
	tlsCfg.InsecureSkipVerify = c.InsecureSkipVerify
	return tlsCfg, nil
}

// LoadTLSConfig loads the TLS configuration.
func (c ServerConfig) LoadTLSConfig(_ context.Context) (*tls.Config, error) {
	tlsCfg, err := c.loadTLSConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS config: %w", err)
	}
	if c.ClientCAFile != "" {
		reloader, err := newClientCAsReloader(c.ClientCAFile, &c)
		if err != nil {
			return nil, err
		}
		if c.ReloadClientCAFile {
			err = reloader.startWatching()
			if err != nil {
				return nil, err
			}
			tlsCfg.GetConfigForClient = func(*tls.ClientHelloInfo) (*tls.Config, error) { return reloader.getClientConfig(tlsCfg) }
		}
		tlsCfg.ClientCAs = reloader.certPool
		tlsCfg.ClientAuth = tls.RequireAndVerifyClientCert
	}
	return tlsCfg, nil
}

func (c ServerConfig) loadClientCAFile() (*x509.CertPool, error) {
	return c.loadCert(c.ClientCAFile)
}

func (c Config) hasCA() bool   { return c.hasCAFile() || c.hasCAPem() }
func (c Config) hasCert() bool { return c.hasCertFile() || c.hasCertPem() }
func (c Config) hasKey() bool  { return c.hasKeyFile() || c.hasKeyPem() }

func (c Config) hasCAFile() bool { return c.CAFile != "" }
func (c Config) hasCAPem() bool  { return len(c.CAPem) != 0 }

func (c Config) hasCertFile() bool { return c.CertFile != "" }
func (c Config) hasCertPem() bool  { return len(c.CertPem) != 0 }

func (c Config) hasKeyFile() bool { return c.KeyFile != "" }
func (c Config) hasKeyPem() bool  { return len(c.KeyPem) != 0 }

func convertVersion(v string, defaultVersion uint16) (uint16, error) {
	// Use a default that is explicitly defined
	if v == "" {
		return defaultVersion, nil
	}
	val, ok := tlsVersions[v]
	if !ok {
		return 0, fmt.Errorf("unsupported TLS version: %q", v)
	}
	return val, nil
}

var tlsVersions = map[string]uint16{
	"1.0": tls.VersionTLS10,
	"1.1": tls.VersionTLS11,
	"1.2": tls.VersionTLS12,
	"1.3": tls.VersionTLS13,
}

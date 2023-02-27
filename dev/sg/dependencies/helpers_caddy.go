package dependencies

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"runtime"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func checkCaddyTrusted(_ context.Context) error {
	certPath, err := caddySourcegraphCertificatePath()
	if err != nil {
		return errors.Wrap(err, "failed to determine path where proxy stores certificates")
	}

	ok, err := pathExists(certPath)
	if !ok || err != nil {
		return errors.New("sourcegraph.test certificate not found. highly likely it's not trusted by system")
	}

	rawCert, err := os.ReadFile(certPath)
	if err != nil {
		return errors.Wrap(err, "could not read certificate")
	}

	cert, err := pemDecodeSingleCert(rawCert)
	if err != nil {
		return errors.Wrap(err, "decoding cert failed")
	}

	if trusted(cert) {
		return nil
	}
	return errors.New("doesn't look like certificate is trusted")
}

// caddyAppDataDir returns the location of the sourcegraph.test certificate
// that Caddy created or would create.
//
// It's copy&pasted&modified from here: https://sourcegraph.com/github.com/caddyserver/caddy@9ee68c1bd57d72e8a969f1da492bd51bfa5ed9a0/-/blob/storage.go?L114
func caddySourcegraphCertificatePath() (string, error) {
	if basedir := os.Getenv("XDG_DATA_HOME"); basedir != "" {
		return filepath.Join(basedir, "caddy"), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	var appDataDir string
	switch runtime.GOOS {
	case "darwin":
		appDataDir = filepath.Join(home, "Library", "Application Support", "Caddy")
	case "linux":
		appDataDir = filepath.Join(home, ".local", "share", "caddy")
	default:
		return "", errors.Newf("unsupported OS: %s", runtime.GOOS)
	}

	return filepath.Join(appDataDir, "pki", "authorities", "local", "root.crt"), nil
}

func trusted(cert *x509.Certificate) bool {
	chains, err := cert.Verify(x509.VerifyOptions{})
	return len(chains) > 0 && err == nil
}

func pemDecodeSingleCert(pemDER []byte) (*x509.Certificate, error) {
	pemBlock, _ := pem.Decode(pemDER)
	if pemBlock == nil {
		return nil, errors.Newf("no PEM block found")
	}
	if pemBlock.Type != "CERTIFICATE" {
		return nil, errors.Newf("expected PEM block type to be CERTIFICATE, but got '%s'", pemBlock.Type)
	}
	return x509.ParseCertificate(pemBlock.Bytes)
}

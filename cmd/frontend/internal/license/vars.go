package license

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/crypto/ssh"
	log15 "gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
)

// publicKey is the public key used to verify license keys
const publicKey = `ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCi0YsgvNQMN+srcMXZtDAmF31GNurQsbr5yky3fQ4n113qvJOcfrmeA2i74fVKRzcOFdYqISSFrwI7WT946jrrbg4WCbm5vyUzRLlDDj7cQQE1St7QmuAHwdUvAgQWzfQ4Bf4qqi4RNdGtObxU9hc8l3wmqZCkiezWN72nVDdc0hn+JkZ3qTcG+1MLqWuoPISX7/HSWdOJATTJkHkS4nmeROeB4NFplrmxW7S3tWMLEW3prxUr4i7BJhVLxOGO/TkPksgR5G2Wb8jktKHghTbCofZ00COlLejloAH1Pmm4NoO0ORaPij6puxIgB6wJ9Ap3tlYn3a9/c/HdB9Or/Rnf`

// bypassLicenseKey is the license key to use in development to bypass the license requirement
const bypassLicenseKey = `24348deeb9916a070914b5617a9a4e2c7bec0d313ca6ae11545ef034c7138d4d8710cddac80980b00426fb44830263268f028c9735`

var (
	licenseKey = env.Get("LICENSE_KEY", "", "license key that unlocks this instance of Sourcegraph Server")

	// license is the decoded and verified value of the LICENSE_KEY env var.
	license *License

	// licenseStatus is the status of the license key set with LICENSE_KEY.
	licenseStatus LicenseStatus

	privateKeyEnv = env.Get("LICENSE_GENERATOR_PRIVATE_KEY", "", "private key used to generate license keys (should only be set on sourcegraph.com)")
	privateKey    ssh.Signer
)

func init() {
	var err error
	pubKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(publicKey))
	if err != nil {
		log.Fatalf("could not parse public key for license verification: %s", err)
	}
	initLicense(pubKey)

	if pk, err := ssh.ParsePrivateKey([]byte(privateKeyEnv)); err == nil {
		log15.Info("Sourcegraph license generator key detected. This instance can generate Sourcegraph Server licenses")
		privateKey = pk
	}
}

// initLicense initializes the license and license status using the value of the LICENSE_KEY env var.
func initLicense(publicKey ssh.PublicKey) {
	if licenseKey == bypassLicenseKey {
		licenseStatus = LicenseValid
		return
	}

	if licenseKey == "" {
		licenseStatus = LicenseMissing
		return
	}
	signedLic, err := decode(licenseKey)
	if err != nil {
		licenseStatus = LicenseInvalid
		return
	}
	if !verify(signedLic, publicKey) {
		licenseStatus = LicenseInvalid
		return
	}
	licenseStatus = LicenseValid
	license = &signedLic.License
}

type LicenseStatus string

const (
	LicenseValid   LicenseStatus = "valid"
	LicenseInvalid               = "invalid"
	LicenseMissing               = "missing"
	LicenseExpired               = "expired"
)

// Get returns the license and license status. The license is decoded and validated from the environment variable
// LICENSE_KEY. The caller should always check if the license is nil regardless of the value of the license status.
// The license coudl be nil while the license status is valid if running in development.
func Get(appID string) (*License, LicenseStatus) {
	if license == nil || licenseStatus != LicenseValid {
		return nil, licenseStatus
	}
	if appID != license.AppID {
		return nil, LicenseInvalid
	}
	if licenseStatus == LicenseValid && license.Expired() {
		return nil, LicenseExpired
	}
	return license, licenseStatus
}

// WithLicenseGenerator attaches a HTTP handler for generating license keys if the env var LICENSE_GENERATOR_PRIVATE_KEY is set.
func WithLicenseGenerator(h http.Handler) http.Handler {
	if privateKey == nil {
		return h
	}
	licensePassword := "futuresooner"
	m := http.NewServeMux()
	m.Handle("/.internal/license", handlerutil.NewBasicAuthHandlerWithPassword(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			w.Write([]byte(`<html>
<form method="post">
	<label>AppID</label> <input name="appID">
	<label>Expiry</label> <input type="date" name="expiry">
	<input type="submit" value="Submit">
</form>
</html>`))
		case "POST":
			appID := r.FormValue("appID")
			expiryVal := r.FormValue("expiry")
			var expiry *time.Time
			if expiryVal != "" {
				exp, err := time.Parse("2006-01-02", expiryVal)
				if err != nil {
					http.Error(w, fmt.Sprintf("couldn't parse time: %s", err), http.StatusInternalServerError)
					return
				}
				expiry = &exp
			}
			licenseKey, err := generate(appID, expiry, privateKey)
			if err != nil {
				http.Error(w, fmt.Sprintf("couldn't generate license key: %s", err), http.StatusInternalServerError)
				return
			}

			w.Write([]byte("<html>"))
			if appID != "" {
				w.Write([]byte(fmt.Sprintf("AppID: %s<br>", appID)))
			} else {
				w.Write([]byte("WARNING: no AppID was set. This will be an anonymous license.<br>"))
			}
			if expiry == nil {
				w.Write([]byte("Expiry: no expiry set, license is perpetual<br>"))
			} else {
				w.Write([]byte(fmt.Sprintf("Expiry: %v (%.1f days from now)<br>", expiry, expiry.Sub(time.Now()).Hours()/24)))
			}
			w.Write([]byte(fmt.Sprintf(`<a href="data:application/octet-stream,%s" download="sourcegraph-server.sgl">Download license key</a><br>`, url.QueryEscape(licenseKey))))
			w.Write([]byte("</html>"))
		}

	}), licensePassword))
	m.Handle("/", h)
	return m
}

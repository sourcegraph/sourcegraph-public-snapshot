package licensing

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/license"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"golang.org/x/crypto/ssh"
)

// publicKey is the public key used to verify product license keys.
//
// It is hardcoded here intentionally (we only have one private signing key, and we don't yet
// support/need key rotation). The corresponding private key is at
// https://team-sourcegraph.1password.com/vaults/dnrhbauihkhjs5ag6vszsme45a/allitems/zkdx6gpw4uqejs3flzj7ef5j4i
// and set below in SOURCEGRAPH_LICENSE_GENERATION_KEY.
var publicKey = func() ssh.PublicKey {
	// To convert PKCS#8 format (which `openssl rsa -in key.pem -pubout` produces) to the format
	// that ssh.ParseAuthorizedKey reads here, use `ssh-keygen -i -mPKCS8 -f key.pub`.
	const publicKeyData = `ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDUUd9r83fGmYVLzcqQp5InyAoJB5lLxlM7s41SUUtxfnG6JpmvjNd+WuEptJGk0C/Zpyp/cCjCV4DljDs8Z7xjRbvJYW+vklFFxXrMTBs/+HjpIBKlYTmG8SqTyXyu1s4485Kh1fEC5SK6z2IbFaHuSHUXgDi/IepSOg1QudW4n8J91gPtT2E30/bPCBRq8oz/RVwJSDMvYYjYVb//LhV0Mx3O6hg4xzUNuwiCtNjCJ9t4YU2sV87+eJwWtQNbSQ8TelQa8WjG++XSnXUHw12bPDe7wGL/7/EJb7knggKSAMnpYpCyV35dyi4DsVc46c+b6P0gbVSosh3Uc3BJHSWF`
	var err error
	publicKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(publicKeyData))
	if err != nil {
		panic("failed to parse public key for license verification: " + err.Error())
	}
	return publicKey
}()

// ParseProductLicenseKey parses and verifies the license key using the license verification public
// key (publicKey in this package).
func ParseProductLicenseKey(licenseKey string) (*license.Info, string, error) {
	return license.ParseSignedKey(licenseKey, publicKey)
}

// ParseProductLicenseKeyWithBuiltinOrGenerationKey is like ParseProductLicenseKey, except it tries
// parsing and verifying the license key with the license generation key (if set), instead of always
// using the builtin license key.
//
// It is useful for local development when using a test license generation key (whose signatures
// aren't considered valid when verified using the builtin public key).
func ParseProductLicenseKeyWithBuiltinOrGenerationKey(licenseKey string) (*license.Info, string, error) {
	var k ssh.PublicKey
	if licenseGenerationPrivateKey != nil {
		k = licenseGenerationPrivateKey.PublicKey()
	} else {
		k = publicKey
	}
	return license.ParseSignedKey(licenseKey, k)
}

// Cache the parsing of the license key because public key crypto can be slow.
var (
	mu            sync.Mutex
	lastKeyText   string
	lastInfo      *license.Info
	lastSignature string
)

var MockGetConfiguredProductLicenseInfo func() (*license.Info, string, error)

// GetConfiguredProductLicenseInfo returns information about the current product license key
// specified in site configuration.
func GetConfiguredProductLicenseInfo() (*license.Info, error) {
	info, _, err := GetConfiguredProductLicenseInfoWithSignature()
	return info, err
}

// GetConfiguredProductLicenseInfoWithSignature returns information about the current product license key
// specified in site configuration, with the signed key's signature.
func GetConfiguredProductLicenseInfoWithSignature() (*license.Info, string, error) {
	if MockGetConfiguredProductLicenseInfo != nil {
		return MockGetConfiguredProductLicenseInfo()
	}

	// Support reading the license key from the environment (intended for development, because we
	// don't want to commit a valid license key to dev/config.json in the OSS repo).
	keyText := os.Getenv("SOURCEGRAPH_LICENSE_KEY")
	if keyText == "" {
		keyText = conf.Get().LicenseKey
	}

	if keyText != "" {
		mu.Lock()
		defer mu.Unlock()

		var (
			info      *license.Info
			signature string
		)
		if keyText == lastKeyText {
			info = lastInfo
			signature = lastSignature
		} else {
			var err error
			info, signature, err = ParseProductLicenseKey(keyText)
			if err != nil {
				return nil, "", err
			}
			lastKeyText = keyText
			lastInfo = info
			lastSignature = signature
		}
		return info, signature, nil
	}
	// No license key.
	return nil, "", nil
}

// licenseGenerationPrivateKeyURL is the URL where Sourcegraph staff can find the private key for
// generating licenses.
//
// NOTE: If you change this, use text search to replace other instances of it (in source code
// comments).
const licenseGenerationPrivateKeyURL = "https://team-sourcegraph.1password.com/vaults/dnrhbauihkhjs5ag6vszsme45a/allitems/zkdx6gpw4uqejs3flzj7ef5j4i"

// envLicenseGenerationPrivateKey (the env var SOURCEGRAPH_LICENSE_GENERATION_KEY) is the
// PEM-encoded form of the private key used to sign product license keys. It is stored at
// https://team-sourcegraph.1password.com/vaults/dnrhbauihkhjs5ag6vszsme45a/allitems/zkdx6gpw4uqejs3flzj7ef5j4i.
var envLicenseGenerationPrivateKey = env.Get("SOURCEGRAPH_LICENSE_GENERATION_KEY", "", "the PEM-encoded form of the private key used to sign product license keys ("+licenseGenerationPrivateKeyURL+")")

// licenseGenerationPrivateKey is the private key used to generate license keys.
var licenseGenerationPrivateKey = func() ssh.Signer {
	if envLicenseGenerationPrivateKey == "" {
		// Most Sourcegraph instances don't use/need this key. Generally only Sourcegraph.com and
		// local dev will have this key set.
		return nil
	}
	privateKey, err := ssh.ParsePrivateKey([]byte(envLicenseGenerationPrivateKey))
	if err != nil {
		log.Fatalf("Failed to parse private key in SOURCEGRAPH_LICENSE_GENERATION_KEY env var: %s.", err)
	}
	return privateKey
}()

// GenerateProductLicenseKey generates a product license key using the license generation private
// key configured in site configuration.
func GenerateProductLicenseKey(info license.Info) (string, error) {
	if envLicenseGenerationPrivateKey == "" {
		const msg = "no product license generation private key was configured"
		if env.InsecureDev {
			// Show more helpful error message in local dev.
			return "", fmt.Errorf("%s (for testing by Sourcegraph staff: set the SOURCEGRAPH_LICENSE_GENERATION_KEY env var to the key obtained at %s)", msg, licenseGenerationPrivateKeyURL)
		}
		return "", errors.New(msg)
	}

	licenseKey, err := license.GenerateSignedKey(info, licenseGenerationPrivateKey)
	if err != nil {
		return "", err
	}
	return licenseKey, nil
}

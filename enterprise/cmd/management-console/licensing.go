package main

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/management-console/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/pkg/license"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
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
		keyText = conf.Get().Critical.LicenseKey
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

// productNameWithBrand returns the product name with brand (e.g., "Sourcegraph Enterprise") based
// on the license info.
func productNameWithBrand(hasLicense bool, licenseTags []string) string {
	if !hasLicense {
		return "Sourcegraph Core"
	}

	hasTag := func(tag string) bool {
		for _, t := range licenseTags {
			if tag == t {
				return true
			}
		}
		return false
	}

	var name string
	if hasTag("starter") {
		name = " Starter"
	}

	var misc []string
	if hasTag("trial") {
		misc = append(misc, "trial")
	}
	if hasTag("dev") {
		misc = append(misc, "dev use only")
	}
	if len(misc) > 0 {
		name += " (" + strings.Join(misc, ", ") + ")"
	}

	return "Sourcegraph Enterprise" + name
}

// Make the Site.productSubscription GraphQL field return the actual info about the product license,
// if any.
func init() {
	fmt.Println("INIT")
	shared.GetProductNameWithBrand = productNameWithBrand
	shared.GetConfiguredProductLicenseInfo = func() (*shared.ProductLicenseInfo, error) {
		info, err := GetConfiguredProductLicenseInfo()
		if info == nil || err != nil {
			return nil, err
		}
		return &shared.ProductLicenseInfo{
			TagsValue:      info.Tags,
			UserCountValue: info.UserCount,
			ExpiresAtValue: info.ExpiresAt,
		}, nil
	}
	shared.GetLicenseTags = func() ([]string, error) {
		info, err := GetConfiguredProductLicenseInfo()
		if info == nil || err != nil {
			return nil, err
		}
		return info.Tags, nil
	}
	shared.GetLicenseExpiresAt = func() (time.Time, error) {
		info, err := GetConfiguredProductLicenseInfo()
		if info == nil || err != nil {
			return time.Time{}, err
		}
		return info.ExpiresAt, nil
	}
	shared.GetLicenseUserCount = func() (uint, error) {
		info, err := GetConfiguredProductLicenseInfo()
		if info == nil || err != nil {
			return 0, err
		}
		return info.UserCount, nil
	}

}

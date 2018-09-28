package licensing

import (
	"context"
	"os"
	"sync"

	"github.com/pkg/errors"
	"github.com/sourcegraph/enterprise/pkg/license"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/env"
	"golang.org/x/crypto/ssh"
)

// publicKey is the public key used to verify Sourcegraph license keys.
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

// ParseProductLicenseKey parses and verifies the license key using the license verification
// public key (publicKey in this package).
func ParseProductLicenseKey(licenseKey string) (*license.Info, error) {
	return license.ParseSignedKey(licenseKey, publicKey)
}

// Cache the parsing of the license key because public key crypto can be slow.
var (
	mu          sync.Mutex
	lastKeyText string
	lastInfo    *license.Info
)

// GetConfiguredProductLicenseInfo returns information about the current Sourcegraph license key specified
// in site configuration.
func GetConfiguredProductLicenseInfo() (*license.Info, error) {
	// Support reading the license key from the environment (intended for development, because
	// we don't want to commit a valid license key to dev/config.json in the OSS repo).
	keyText := os.Getenv("SOURCEGRAPH_LICENSE_KEY")
	if keyText == "" {
		keyText = conf.Get().LicenseKey
	}

	if keyText != "" {
		mu.Lock()
		defer mu.Unlock()

		var info *license.Info
		if keyText == lastKeyText {
			info = lastInfo
		} else {
			var err error
			info, err = ParseProductLicenseKey(keyText)
			if err != nil {
				return nil, err
			}
			lastKeyText = keyText
			lastInfo = info
		}
		return info, nil
	}

	// No license key.
	return &license.Info{Plan: "Free"}, nil
}

// Make the Site.productSubscription GraphQL field return the actual info about the Sourcegraph
// license (instead of the stub info from the OSS build).
func init() {
	graphqlbackend.GetConfiguredProductLicenseInfo = func(ctx context.Context) (*graphqlbackend.ProductLicenseInfo, error) {
		info, err := GetConfiguredProductLicenseInfo()
		if err != nil {
			return nil, err
		}
		return &graphqlbackend.ProductLicenseInfo{
			PlanValue:      info.Plan,
			UserCountValue: info.UserCount,
			ExpiresAtValue: info.ExpiresAt,
		}, nil
	}
}

// envLicenseGenerationPrivateKey (the env var SOURCEGRAPH_LICENSE_GENERATION_KEY) is the
// PEM-encoded form of the private key used to sign Sourcegraph license keys. It is stored at
// https://team-sourcegraph.1password.com/vaults/dnrhbauihkhjs5ag6vszsme45a/allitems/zkdx6gpw4uqejs3flzj7ef5j4i.
var envLicenseGenerationPrivateKey = env.Get("SOURCEGRAPH_LICENSE_GENERATION_KEY", "", "the PEM-encoded form of the private key used to sign Sourcegraph license keys (https://team-sourcegraph.1password.com/vaults/dnrhbauihkhjs5ag6vszsme45a/allitems/zkdx6gpw4uqejs3flzj7ef5j4i)")

// GenerateProductLicenseKey generates a Sourcegraph license key using the license generation
// private key configured in site configuration.
func GenerateProductLicenseKey(info license.Info) (string, error) {
	if envLicenseGenerationPrivateKey == "" {
		return "", errors.New("no Sourcegraph license generation private key was configured")
	}
	privateKey, err := ssh.ParsePrivateKey([]byte(envLicenseGenerationPrivateKey))
	if err != nil {
		return "", errors.WithMessage(err, "parsing Sourcegraph license generation private key")
	}

	licenseKey, err := license.GenerateSignedKey(info, privateKey)
	if err != nil {
		return "", err
	}
	return licenseKey, nil
}

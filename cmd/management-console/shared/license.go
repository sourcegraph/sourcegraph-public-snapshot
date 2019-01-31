package shared

import (
	"context"
	"time"

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

var GetProductNameWithBrand = func(hasLicense bool, licenseTags []string) string {
	return "Sourcegraph OSS"
}

var GetConfiguredProductLicenseInfo = func() (*ProductLicenseInfo, error) {
	return nil, nil // OSS builds have no license
}

var GetLicenseUserCount = func() (uint, error) {
	return 0, nil // OSS builds have no license
}

var GetLicenseExpiresAt = func() (time.Time, error) {
	return time.Time{}, nil // OSS builds have no license
}

var GetLicenseTags = func() ([]string, error) {
	return []string{}, nil // OSS builds have no license
}

func ProductNameWithBrand() (string, error) {
	info, err := GetConfiguredProductLicenseInfo()
	if err != nil {
		return "", err
	}
	hasLicense := info != nil
	var licenseTags []string
	if hasLicense {
		licenseTags = info.TagsValue
	}
	return GetProductNameWithBrand(hasLicense, licenseTags), nil
}

// // ActualUserCount is called to obtain the actual maximum number of user accounts that have been active
// // on this Sourcegraph instance for the current license.
var ActualUserCount = func(ctx context.Context) (int32, error) {
	return 0, nil
}

// // ActualUserCountDate is called to obtain the timestamp when the actual maximum number of user accounts
// // that have been active on this Sourcegraph instance for the current license was reached.
var ActualUserCountDate = func(ctx context.Context) (string, error) {
	return "", nil
}

package shared

import (
	"strconv"
	"strings"

	"github.com/garyburd/redigo/redis"
	"github.com/sourcegraph/sourcegraph/enterprise/pkg/license"
	"github.com/sourcegraph/sourcegraph/pkg/redispool"
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

var (
	pool      = redispool.Store
	keyPrefix = "license_user_count:"

	started bool
)

// GetMaxUsers gets the max users associated with a license key.
func GetMaxUsers(signature string) (int, string, error) {
	c := pool.Get()
	defer c.Close()

	if signature == "" {
		// No license key is in use.
		return 0, "", nil
	}

	return getMaxUsers(c, signature)
}

func getMaxUsers(c redis.Conn, key string) (int, string, error) {
	lastMax, err := redis.String(c.Do("HGET", maxUsersKey(), key))
	if err != nil && err != redis.ErrNil {
		return 0, "", err
	}
	lastMaxInt := 0
	if lastMax != "" {
		lastMaxInt, err = strconv.Atoi(lastMax)
		if err != nil {
			return 0, "", err
		}
	}
	lastMaxDate, err := redis.String(c.Do("HGET", maxUsersTimeKey(), key))
	if err != nil && err != redis.ErrNil {
		return 0, "", err
	}
	return lastMaxInt, lastMaxDate, nil
}

func maxUsersKey() string {
	return keyPrefix + "max"
}

func maxUsersTimeKey() string {
	return keyPrefix + "max_time"
}

func actualUserCountDate(signature string) (string, error) {
	_, date, err := GetMaxUsers(signature)
	return date, err
}

package secrets

import (
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/sourcegraph/sourcegraph/internal/conf"
)

var CryptObject Encrypter

const (
	// #nosec G101
	sourcegraphSecretfileEnvvar = "SOURCEGRAPH_SECRET_FILE"
	sourcegraphCryptEnvvar      = "SOURCEGRAPH_CRYPT_KEY"
	validKeyLength              = 32
)

func init() {
	cryptKey, cryptOK := os.LookupEnv(sourcegraphCryptEnvvar)

	// set the default location if none exists
	secretFile := os.Getenv(sourcegraphSecretfileEnvvar)
	if secretFile == "" {
		// #nosec G101
		secretFile = "/var/lib/sourcegraph/token"
	}

	_, err := os.Stat(secretFile)
	// generate a secret for non-k8s deployments
	if err != nil && !cryptOK {
		d := conf.DeployType()
		if conf.IsDeployTypeKubernetes(d) { // Expect a k8s secret
			panic(fmt.Sprintf("Either specify environment variable %s or provide the secrets file %s.",
				sourcegraphCryptEnvvar,
				sourcegraphSecretfileEnvvar))
		}
		c := 32
		b := make([]byte, c)
		_, err := rand.Read(b)
		if err != nil {
			panic(fmt.Sprintf("Unable to read from random source: %v", err))
		}
		err = ioutil.WriteFile(secretFile, b, 0600)
		if err != nil {
			panic(err)
		}
		CryptObject.EncryptionKey = b
		return
	}

	if err == nil {
		contents, readErr := ioutil.ReadFile(secretFile)
		if readErr != nil {
			panic(fmt.Sprintf("Couldn't read file %s", sourcegraphSecretfileEnvvar))
		}
		if len(contents) < validKeyLength {
			panic(fmt.Sprintf("Key length of %d characters is required.", validKeyLength))
		}
		CryptObject.EncryptionKey = contents
	} else {
		if len(cryptKey) != validKeyLength {
			panic(fmt.Sprintf("Key length of %d characters is required.", validKeyLength))
		}
		CryptObject.EncryptionKey = []byte(cryptKey)
	}
}

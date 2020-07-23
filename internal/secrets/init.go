package secrets

import (
	"fmt"
	"io/ioutil"
	"os"
)

var CryptObject EncryptionStore

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

	// a lack of encryption keys means we cannot run the application, hence panic.
	if err != nil && !cryptOK {
		panic(fmt.Sprintf("Either specify environment variable %s or provide the secrets file %s.",
			sourcegraphCryptEnvvar,
			sourcegraphSecretfileEnvvar))
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

package secrets

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/sourcegraph/sourcegraph/internal/conf"
)

var CryptObject Encrypter
var isEncrypted bool

const (
	// #nosec G101
	sourcegraphSecretfileEnvvar = "SOURCEGRAPH_SECRET_FILE"
	sourcegraphCryptEnvvar      = "SOURCEGRAPH_CRYPT_KEY"
)

func ConfiguredToEncrypt() bool {
	return isEncrypted
}

func init() {
	isEncrypted = false

	envCryptKey, cryptOK := os.LookupEnv(sourcegraphCryptEnvvar)
	var encryptionKey []byte

	// set the default location if none exists
	secretFile := os.Getenv(sourcegraphSecretfileEnvvar)
	if secretFile == "" {
		// #nosec G101
		secretFile = "/var/lib/sourcegraph/token"
	}

	_, err := os.Stat(secretFile)

	// reading from a file is first order
	if err == nil {
		contents, readErr := ioutil.ReadFile(secretFile)
		if readErr != nil {
			panic(fmt.Sprintf("couldn't read file %s", sourcegraphSecretfileEnvvar))
		}
		if len(contents) < validKeyLength {
			panic(fmt.Sprintf("key length of %d characters is required.", validKeyLength))
		}
		encryptionKey = contents
		err = os.Chmod(secretFile, 0400)
		if err != nil {
			panic("failed to make secrets file read only.")
		}
		CryptObject.EncryptionKey = encryptionKey
		return
	}

	// environment is second order
	if cryptOK {
		CryptObject.EncryptionKey = []byte(envCryptKey)
		return
	}

	// for the single docker case, we generate the secret
	deployType := conf.DeployType()
	if conf.IsDeployTypeSingleDockerContainer(deployType) {
		b, err := GenerateRandomAESKey()
		if err != nil {
			panic(fmt.Sprintf("unable to read from random source: %v", err))
		}
		err = ioutil.WriteFile(secretFile, b, 0600)
		if err != nil {
			panic(err)
		}

		err = os.Chmod(secretFile, 0400)
		if err != nil {
			panic("failed to make secrets file read only.")
		}
		CryptObject.EncryptionKey = b
	}

	// wrapping in deploytype check so that we can still compile and test locally
	if !(conf.IsDev(deployType) || os.Getenv("CI") == "") {
		// for k8s & docker compose, expect a secret to be provided
		panic(fmt.Sprintf("Either specify environment variable %s or provide the secrets file %s.",
			sourcegraphCryptEnvvar,
			sourcegraphSecretfileEnvvar))
	}

	isEncrypted = true
}

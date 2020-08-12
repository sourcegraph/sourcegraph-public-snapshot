package secrets

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/sourcegraph/sourcegraph/internal/conf"
)

// TODO: Make private, access via functions
var CryptObject Encryptor
var configuredToEncrypt bool

const (
	// #nosec G101
	sourcegraphSecretfileEnvvar = "SOURCEGRAPH_SECRET_FILE"
	sourcegraphCryptEnvvar      = "SOURCEGRAPH_CRYPT_KEY"
)

// gatherKeys splits the encryption string into its potential two components
func gatherKeys(data []byte) (oldKey, newKey []byte) {
	parts := bytes.Split(data, []byte(","))
	if len(parts) > 2 {
		panic("no more than two encryption keys should be specified.")
	}
	if len(parts) == 1 {
		return parts[0], nil
	}

	return parts[0], parts[1]
}

func init() {
	configuredToEncrypt = false

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

		newKey, oldKey := gatherKeys(encryptionKey)

		CryptObject.EncryptionKeys = [][]byte{primaryKey: newKey, secondaryKey: oldKey}
		return
	}

	// environment is second order
	if cryptOK {
		if len(envCryptKey) != validKeyLength {
			panic(fmt.Sprintf("encryption key must be %d characters", validKeyLength))
		}
		newKey, oldKey := gatherKeys(encryptionKey)
		CryptObject.EncryptionKeys = [][]byte{primaryKey: newKey, secondaryKey: oldKey}
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
		newKey, oldkey := gatherKeys(b)
		CryptObject.EncryptionKeys = [][]byte{primaryKey: newKey, secondaryKey: oldkey}
	}

	// wrapping in deploytype check so that we can still compile and test locally
	if os.Getenv("CI") != "" || conf.IsDev(deployType) {
		configuredToEncrypt = true
	} else {
		// for k8s & docker compose, expect a secret to be provided
		panic(fmt.Sprintf("Either specify environment variable %s or provide the secrets file %s.",
			sourcegraphCryptEnvvar,
			sourcegraphSecretfileEnvvar))
	}
}

package secrets

import (
	"bytes"
	"crypto/rand"
	"io/ioutil"
	"os"
	"path"
	"sync"

	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/internal/conf"
)

const (
	// #nosec G101
	sourcegraphSecretfileEnvvar = "SOURCEGRAPH_SECRET_FILE"
	sourcegraphCryptEnvvar      = "SOURCEGRAPH_CRYPT_KEY"
)

// gatherKeys splits the comma-separated encryption data into its potential two components:
// primary and secondary keys, where the first key is assumed to be the primary key.
func gatherKeys(data []byte) (primaryKey, secondaryKey []byte, err error) {
	parts := bytes.Split(data, []byte(","))
	if len(parts) > 2 {
		return nil, nil, errors.Errorf("expect at most two encryption keys but got %d", len(parts))
	}
	if len(parts) == 1 {
		return parts[0], nil, nil
	}
	return parts[0], parts[1], nil
}

var initErr error
var initOnce sync.Once

// Init creates the defaultEncryptor by ingesting user encryption key(s).
// For production deployments, the secret value CAN ONLY be generated by the user and loaded via a file or env var.
// For single server docker deployments, we will generate the secret file and write it to disk.
func Init() error {
	initOnce.Do(func() {
		initErr = initDefaultEncryptor()
	})
	return initErr
}

// defaultEncryptor is configured during init, if no keys are provided it will implement noOpEncryptor.
var defaultEncryptor encryptor = noOpEncryptor{}

// NOTE: MockDefaultEncryptor should only be called in tests where a random encryptor is
// needed to test transparent encryption and decryption.
func MockDefaultEncryptor() {
	defaultEncryptor = newAESGCMEncodedEncryptor(mustGenerateRandomAESKey(), nil)
}

func initDefaultEncryptor() error {
	var encryptionKey []byte

	// set the default location if none exists
	secretFile := os.Getenv(sourcegraphSecretfileEnvvar)
	if secretFile == "" {
		// #nosec G101
		secretFile = "/var/lib/sourcegraph/token"
	}

	// reading from a file is first order
	fileInfo, err := os.Stat(secretFile)
	if err == nil {
		perm := fileInfo.Mode().Perm()
		if perm != os.FileMode(0400) {
			return errors.New("key file permissions are not 0400")
		}

		contents, readErr := ioutil.ReadFile(secretFile)
		if readErr != nil {
			return errors.Wrapf(readErr, "couldn't read file %s", sourcegraphSecretfileEnvvar)
		}
		if len(contents) < requiredKeyLength {
			return errors.Errorf("key length of %d characters is required", requiredKeyLength)
		}
		encryptionKey = contents

		primaryKey, secondaryKey, err := gatherKeys(encryptionKey)
		if err != nil {
			return err
		}

		defaultEncryptor = newAESGCMEncodedEncryptor(primaryKey, secondaryKey)
		return nil
	}

	envCryptKey, cryptOK := os.LookupEnv(sourcegraphCryptEnvvar)
	// environment is second order
	if cryptOK {
		if len(envCryptKey) != requiredKeyLength {
			return errors.Errorf("encryption key must be %d characters", requiredKeyLength)
		}
		primaryKey, secondaryKey, err := gatherKeys(encryptionKey)
		if err != nil {
			return err
		}

		defaultEncryptor = newAESGCMEncodedEncryptor(primaryKey, secondaryKey)
		return nil
	}

	// for the single docker case, we generate the secret
	deployType := conf.DeployType()
	if conf.IsDeployTypeSingleDockerContainer(deployType) {
		b, err := generateRandomAESKey()
		if err != nil {
			return errors.Wrap(err, "unable to generate random key")
		}

		err = os.MkdirAll(path.Dir(secretFile), os.ModePerm)
		if err != nil {
			return err
		}

		err = ioutil.WriteFile(secretFile, b, 0600)
		if err != nil {
			return err
		}

		err = os.Chmod(secretFile, 0400)
		if err != nil {
			return errors.Wrap(err, "unable to change key file permissions to 0400")
		}
		newKey, _, err := gatherKeys(b)
		if err != nil {
			return err
		}

		defaultEncryptor = newAESGCMEncodedEncryptor(newKey, nil)
		return nil
	}

	// wrapping in deploytype check so that we can still compile and test locally
	if os.Getenv("CI") != "" || conf.IsDev(deployType) {
		defaultEncryptor = noOpEncryptor{}
		return nil
	}

	log15.Warn("no encryption option enabled")
	return nil

	// TODO: Enable this once docs are in place for
	// for k8s & docker compose, expect a secret to be provided
	//return errors.Errorf("Either specify environment variable %s or provide the secrets file %s",
	//	sourcegraphCryptEnvvar,
	//	sourcegraphSecretfileEnvvar)
}

// generateRandomAESKey generates a random key that can be used for AES-256 encryption.
func generateRandomAESKey() ([]byte, error) {
	b := make([]byte, requiredKeyLength)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// mustGenerateRandomAESKey generates a random AES key and panics for any error.
func mustGenerateRandomAESKey() []byte {
	key, err := generateRandomAESKey()
	if err != nil {
		panic(err)
	}
	return key
}

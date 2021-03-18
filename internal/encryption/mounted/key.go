package mounted

import (
	"context"
	"os"

	"github.com/cockroachdb/errors"
	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/schema"
)

type Key struct {
	keyname string
	secret  string
}

func (k *Key) ID(ctx context.Context) (string, error) {
	return k.keyname, nil
}

// TODO: Define this.
func (k *Key) Encrypt(ctx context.Context, value []byte) ([]byte, error) {
	return []byte{}, nil
}

// TODO: Define this.
func Decrypt(ctx context.Context, cipherText []byte) (*encryption.Secret, error) {
	return nil, nil
}

func NewKey(ctx context.Context, k schema.MountedEncryptionKey) (*Key, error) {
	if k.EnvVarName != "" && k.Filepath == "" {
		secret := os.Getenv(k.EnvVarName)
		if secret == "" {
			return nil, errors.Errorf("env variable %q is not set", k.EnvVarName)
		}

		return &Key{
			keyname: k.Keyname,
			secret:  secret,
		}, nil
	} else if k.Filepath != "" && k.EnvVarName == "" {
		f, err := os.Stat(k.Filepath)
		if err != nil {
			return nil, errors.Errorf("failed to locate file %q: %v", k.Filepath, err)
		}

		secret, err := os.ReadFile(f.Name())
		if err != nil {
			return nil, errors.Errorf("")
		}

		return &Key{
			keyname: k.Keyname,
			secret:  string(secret),
		}, nil
	}

	// Either the user has set none of EnvVarName or Filepath or both in their config. Either way we return an error.
	return nil, errors.Errorf("must use only one of EnvVarName and Filepath, EnvVarName: %q, Filepath: %q", k.EnvVarName, k.Filepath)
}

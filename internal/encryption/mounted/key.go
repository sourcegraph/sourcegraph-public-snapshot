package mounted

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"os"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/encryption/aeshelper"
	"github.com/sourcegraph/sourcegraph/internal/encryption/envelope"
	"github.com/sourcegraph/sourcegraph/internal/encryption/wrapper"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

const (
	mechanismEncrypt  = "encrypt"
	mechanismEnvelope = "envelope"
)

func NewKey(ctx context.Context, k schema.MountedEncryptionKey) (*Key, error) {
	var secret []byte
	if k.EnvVarName != "" && k.Filepath == "" {
		secret = []byte(os.Getenv(k.EnvVarName))

	} else if k.Filepath != "" && k.EnvVarName == "" {
		keyBytes, err := os.ReadFile(k.Filepath)
		if err != nil {
			return nil, errors.Errorf("error reading secret file for %q: %v", k.Keyname, err)
		}
		secret = keyBytes
	} else {
		// Either the user has set none of EnvVarName or Filepath or both in their config. Either way we return an error.
		return nil, errors.Errorf(
			"must use only one of EnvVarName and Filepath, EnvVarName: %q, Filepath: %q",
			k.EnvVarName, k.Filepath,
		)
	}

	if len(secret) != 32 {
		return nil, errors.Errorf("invalid key length: %d, expected 32 bytes", len(secret))
	}

	return &Key{
		keyname: k.Keyname,
		version: k.Version,
		secret:  secret,
	}, nil
}

// Key is an encryption.Key implementation that uses AES GCM encryption, using a
// secret loaded either from an env var or a file
type Key struct {
	keyname string
	secret  []byte
	version string
}

func (k *Key) Version(ctx context.Context) (encryption.KeyVersion, error) {
	return encryption.KeyVersion{
		Type:    "mounted",
		Name:    k.keyname,
		Version: k.version,
	}, nil
}

func (k *Key) Encrypt(ctx context.Context, plaintext []byte) ([]byte, error) {
	ev, err := envelope.Encrypt(plaintext)
	if err != nil {
		return nil, errors.Wrap(err, "envelope encrypting plaintext")
	}

	// Encrypt the envelope key using the keys secret.
	keyCiphertext, keyNonce, err := aeshelper.Encrypt(k.secret, ev.Key)
	if err != nil {
		return nil, err
	}

	// Store both the nonce and the ciphertext in the wrappedKey field.
	wrappedKey := append(keyNonce, keyCiphertext...)

	ek := wrapper.StorableEncryptedKey{
		Mechanism:  mechanismEnvelope,
		KeyName:    k.keyname,
		WrappedKey: wrappedKey,
		Ciphertext: ev.Ciphertext,
		Nonce:      ev.Nonce,
	}

	return ek.Serialize()
}

func (k *Key) Decrypt(ctx context.Context, ciphertext []byte) (*encryption.Secret, error) {
	wr, err := wrapper.FromCiphertext(ciphertext)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshalling encrypted key")
	}

	if wr.Mechanism == "" {
		wr.Mechanism = mechanismEncrypt
	}

	if !strings.HasPrefix(wr.KeyName, k.keyname) {
		return nil, errors.New("invalid key name, are you trying to decrypt something with the wrong key?")
	}

	switch wr.Mechanism {
	case mechanismEncrypt:
		nonceSize, err := getNonceSize(k.secret)
		if err != nil {
			return nil, err
		}
		if len(wr.Ciphertext) < nonceSize {
			return nil, errors.New("malformed ciphertext")
		}
		plaintext, err := aeshelper.Decrypt(k.secret, wr.Ciphertext[nonceSize:], wr.Ciphertext[:nonceSize])
		if err != nil {
			return nil, err
		}

		s := encryption.NewSecret(string(plaintext))
		return &s, nil
	case mechanismEnvelope:
		nonceSize, err := getNonceSize(k.secret)
		if err != nil {
			return nil, err
		}
		if len(wr.WrappedKey) < nonceSize {
			return nil, errors.New("malformed ciphertext")
		}
		keyCiphertext := wr.WrappedKey[nonceSize:]
		keyNonce := wr.WrappedKey[:nonceSize]

		key, err := aeshelper.Decrypt(k.secret, keyCiphertext, keyNonce)
		if err != nil {
			return nil, err
		}

		plaintext, err := envelope.Decrypt(&envelope.Envelope{
			Key:        key,
			Ciphertext: wr.Ciphertext,
			Nonce:      wr.Nonce,
		})
		if err != nil {
			return nil, err
		}

		s := encryption.NewSecret(string(plaintext))
		return &s, nil
	default:
		return nil, errors.Errorf("unsupported encryption mechanism: %s", wr.Mechanism)
	}
}

func getNonceSize(secret []byte) (int, error) {
	block, err := aes.NewCipher(secret)
	if err != nil {
		return 0, errors.Wrap(err, "creating AES cipher")
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return 0, errors.Wrap(err, "creating GCM block cipher")
	}
	return gcm.NonceSize(), nil
}

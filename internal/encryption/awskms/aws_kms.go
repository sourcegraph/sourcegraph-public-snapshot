package awskms

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"

	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/encryption/aeshelper"
	"github.com/sourcegraph/sourcegraph/internal/encryption/envelope"
	"github.com/sourcegraph/sourcegraph/internal/encryption/wrapper"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

const (
	// mechanismEncrypt is the mechanism used prior to the introduction of the envelope helper.
	mechanismEncrypt = "encrypt"
	// mechanismEnvelope is the default for all new encrypted values, which uses the envelope helper.
	mechanismEnvelope = "envelope"
)

func NewKey(ctx context.Context, keyConfig schema.AWSKMSEncryptionKey) (encryption.Key, error) {
	defaultConfig, err := config.LoadDefaultConfig(ctx, awsConfigOptsForKeyConfig(keyConfig)...)
	if err != nil {
		return nil, errors.Wrap(err, "loading config for aws KMS")
	}
	return newKey(ctx, keyConfig, defaultConfig)
}

func newKey(ctx context.Context, keyConfig schema.AWSKMSEncryptionKey, config aws.Config) (*Key, error) {
	k := &Key{
		keyID:  keyConfig.KeyId,
		client: kms.NewFromConfig(config),
	}
	// Test client connection.
	_, err := k.Version(ctx)
	return k, err
}

func awsConfigOptsForKeyConfig(keyConfig schema.AWSKMSEncryptionKey) []func(*config.LoadOptions) error {
	configOpts := []func(*config.LoadOptions) error{}
	if keyConfig.Region != "" {
		configOpts = append(configOpts, config.WithRegion(keyConfig.Region))
	}
	if keyConfig.CredentialsFile != "" {
		configOpts = append(configOpts, config.WithSharedCredentialsFiles([]string{keyConfig.CredentialsFile}))
	}
	return configOpts
}

type Key struct {
	keyID  string
	client *kms.Client
}

func (k *Key) Version(ctx context.Context) (encryption.KeyVersion, error) {
	key, err := k.client.DescribeKey(ctx, &kms.DescribeKeyInput{
		KeyId: &k.keyID,
	})
	if err != nil {
		return encryption.KeyVersion{}, errors.Wrap(err, "getting key version")
	}
	return encryption.KeyVersion{
		Type:    "awskms",
		Version: *key.KeyMetadata.Arn,
		Name:    *key.KeyMetadata.KeyId,
	}, nil
}

// Decrypt a secret, it must have been encrypted with the same Key.
// Encrypted secrets are a base64 encoded string containing the original content.
func (k *Key) Decrypt(ctx context.Context, cipherText []byte) (*encryption.Secret, error) {
	wr, err := wrapper.FromCiphertext(cipherText)
	if err != nil {
		return nil, err
	}

	if wr.Mechanism == "" {
		wr.Mechanism = mechanismEncrypt
	}

	switch wr.Mechanism {
	case mechanismEncrypt:
		buf, err := base64.StdEncoding.DecodeString(string(cipherText))
		if err != nil {
			return nil, err
		}
		ev := encryptedValue{}
		err = json.Unmarshal(buf, &ev)
		if err != nil {
			return nil, err
		}

		res, err := k.client.Decrypt(ctx, &kms.DecryptInput{
			CiphertextBlob: ev.Key,
			KeyId:          &k.keyID,
		})
		if err != nil {
			return nil, err
		}

		// Decrypt ciphertext.
		decBuf, err := aeshelper.Decrypt(res.Plaintext, ev.Ciphertext, ev.Nonce)
		if err != nil {
			return nil, err
		}

		s := encryption.NewSecret(string(decBuf))
		return &s, nil
	case mechanismEnvelope:
		if !strings.HasSuffix(wr.KeyName, k.keyID) {
			return nil, errors.New("invalid key name, are you trying to decrypt something with the wrong key?")
		}

		// First, use KMS to decrypt the envelope key.
		res, err := k.client.Decrypt(ctx, &kms.DecryptInput{
			CiphertextBlob: wr.WrappedKey,
			KeyId:          &k.keyID,
		})
		if err != nil {
			return nil, err
		}

		// Next, decrypt the envelope.
		plaintext, err := envelope.Decrypt(&envelope.Envelope{
			Key:        res.Plaintext,
			Nonce:      wr.Nonce,
			Ciphertext: wr.Ciphertext,
		})
		if err != nil {
			return nil, err
		}

		s := encryption.NewSecret(string(plaintext))
		return &s, nil
	default:
		return nil, errors.Newf("invalid mechanism %q", wr.Mechanism)
	}
}

// Encrypt a secret, storing it as a base64 encoded string.
func (k *Key) Encrypt(ctx context.Context, plaintext []byte) ([]byte, error) {
	ev, err := envelope.Encrypt(plaintext)
	if err != nil {
		return nil, errors.Wrap(err, "envelope encrypting payload")
	}

	// Encrypt the key of the envelope.
	res, err := k.client.Encrypt(ctx, &kms.EncryptInput{
		KeyId:     &k.keyID,
		Plaintext: ev.Key,
	})
	if err != nil {
		return nil, errors.Wrap(err, "encrypting envelope key")
	}

	ek := wrapper.StorableEncryptedKey{
		Mechanism:  mechanismEnvelope,
		KeyName:    *res.KeyId,
		WrappedKey: res.CiphertextBlob,
		Ciphertext: ev.Ciphertext,
		Nonce:      ev.Nonce,
	}

	return ek.Serialize()
}

type encryptedValue struct {
	Key        []byte
	Nonce      []byte
	Ciphertext []byte
}

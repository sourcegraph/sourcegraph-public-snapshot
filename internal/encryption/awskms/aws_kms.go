package awskms

import (
	"context"
	"encoding/base64"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/schema"
)

func NewKey(ctx context.Context, keyConfig schema.AWSKMSEncryptionKey) (encryption.Key, error) {
	config, err := config.LoadDefaultConfig(ctx, awsConfigOptsForKeyConfig(keyConfig)...)
	if err != nil {
		return nil, errors.Wrap(err, "loading config for aws KMS")
	}
	return newKey(ctx, keyConfig, config)
}

func newKey(ctx context.Context, keyConfig schema.AWSKMSEncryptionKey, config aws.Config) (encryption.Key, error) {
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
	buf, err := base64.StdEncoding.DecodeString(string(cipherText))
	if err != nil {
		return nil, err
	}

	// Decrypt ciphertext.
	res, err := k.client.Decrypt(ctx, &kms.DecryptInput{
		KeyId:          &k.keyID,
		CiphertextBlob: buf,
	})
	if err != nil {
		return nil, err
	}

	s := encryption.NewSecret(string(res.Plaintext))
	return &s, nil
}

// Encrypt a secret, storing it as a base64 encoded string.
func (k *Key) Encrypt(ctx context.Context, plaintext []byte) ([]byte, error) {
	// Encrypt plaintext.
	res, err := k.client.Encrypt(ctx, &kms.EncryptInput{
		KeyId:     &k.keyID,
		Plaintext: plaintext,
	})
	if err != nil {
		return nil, err
	}

	buf := base64.StdEncoding.EncodeToString(res.CiphertextBlob)
	return []byte(buf), err
}

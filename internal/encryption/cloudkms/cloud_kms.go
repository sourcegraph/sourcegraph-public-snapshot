package cloudkms

import (
	"context"
	"hash/crc32"
	"strconv"
	"strings"

	kms "cloud.google.com/go/kms/apiv1"
	"cloud.google.com/go/kms/apiv1/kmspb" //nolint:staticcheck // See https://github.com/sourcegraph/sourcegraph/issues/45843
	"google.golang.org/api/option"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/encryption/envelope"
	"github.com/sourcegraph/sourcegraph/internal/encryption/wrapper"
	"github.com/sourcegraph/sourcegraph/schema"
)

const (
	mechanismEncrypt  = "encrypt"
	mechanismEnvelope = "envelope"
)

func NewKey(ctx context.Context, config schema.CloudKMSEncryptionKey) (encryption.Key, error) {
	client, err := kms.NewKeyManagementClient(ctx, clientOptions(config.CredentialsFile)...)
	if err != nil {
		return nil, err
	}
	return newKeyWithClient(ctx, config.Keyname, client)
}

func newKeyWithClient(ctx context.Context, keyName string, client *kms.KeyManagementClient) (*Key, error) {
	k := &Key{
		name:   keyName,
		client: client,
	}
	_, err := k.Version(ctx)
	return k, err
}

func clientOptions(credentialsFile string) []option.ClientOption {
	opts := []option.ClientOption{}
	if credentialsFile != "" {
		opts = append(opts, option.WithCredentialsFile(credentialsFile))
	}
	return opts
}

type Key struct {
	name   string
	client *kms.KeyManagementClient
}

func (k *Key) Version(ctx context.Context) (encryption.KeyVersion, error) {
	key, err := k.client.GetCryptoKey(ctx, &kmspb.GetCryptoKeyRequest{ //nolint:staticcheck // See https://github.com/sourcegraph/sourcegraph/issues/45843
		Name: k.name,
	})
	if err != nil {
		return encryption.KeyVersion{}, errors.Wrap(err, "getting key version")
	}
	// return the primary key version name, as that will include which key
	// revision is currently in use
	return encryption.KeyVersion{
		Type:    "cloudkms",
		Version: key.Primary.Name,
		Name:    key.Name,
	}, nil
}

// Decrypt a secret, it must have been encrypted with the same Key
// encrypted secrets are a base64 encoded string containing the key name and a checksum
func (k *Key) Decrypt(ctx context.Context, cipherText []byte) (_ *encryption.Secret, err error) {
	defer func() {
		cryptographicTotal.WithLabelValues("decrypt", strconv.FormatBool(err == nil)).Inc()
	}()

	wr, err := wrapper.FromCiphertext(cipherText)
	if err != nil {
		return nil, err
	}

	// Fallback value for before we had the wrapper.
	if wr.Mechanism == "" {
		wr.Mechanism = mechanismEncrypt
	}

	if !strings.HasPrefix(wr.KeyName, k.name) {
		return nil, errors.New("invalid key name, are you trying to decrypt something with the wrong key?")
	}

	switch wr.Mechanism {
	case mechanismEncrypt:
		res, err := k.client.Decrypt(ctx, &kmspb.DecryptRequest{ //nolint:staticcheck // See https://github.com/sourcegraph/sourcegraph/issues/45843
			Name:       k.name,
			Ciphertext: wr.Ciphertext,
		})
		if err != nil {
			return nil, err
		}

		// validate checksum
		if int64(crc32Sum(res.Plaintext)) != res.PlaintextCrc32C.GetValue() {
			return nil, errors.New("invalid checksum, either the wrong key was used, or the request was corrupted in transit")
		}

		s := encryption.NewSecret(string(res.Plaintext))
		return &s, nil
	case mechanismEnvelope:
		// First, decrypt the envelope key using KMS.
		res, err := k.client.Decrypt(ctx, &kmspb.DecryptRequest{ //nolint:staticcheck // See https://github.com/sourcegraph/sourcegraph/issues/45843
			Name:       k.name,
			Ciphertext: wr.WrappedKey,
		})
		if err != nil {
			return nil, err
		}

		// Next, decrypt the envelope content.
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
		return nil, errors.Errorf("unsupported encryption mechanism: %s", wr.Mechanism)
	}
}

// Encrypt a secret, storing it as a base64 encoded json blob, this json contains
// the key name, ciphertext, & checksum.
func (k *Key) Encrypt(ctx context.Context, plaintext []byte) (_ []byte, err error) {
	defer func() {
		cryptographicTotal.WithLabelValues("encrypt", strconv.FormatBool(err == nil)).Inc()
		encryptPayloadSize.WithLabelValues(strconv.FormatBool(err == nil)).Observe(float64(len(plaintext)) / 1024)
	}()

	// First, envelope encrypt the plaintext.
	ev, err := envelope.Encrypt(plaintext)
	if err != nil {
		return nil, errors.Wrap(err, "envelope encrypting payload")
	}

	// Encrypt the key of the envelope.
	res, err := k.client.Encrypt(ctx, &kmspb.EncryptRequest{ //nolint:staticcheck // See https://github.com/sourcegraph/sourcegraph/issues/45843
		Name:            k.name,
		Plaintext:       ev.Key,
		PlaintextCrc32C: wrapperspb.Int64(int64(crc32Sum(ev.Key))),
	})
	if err != nil {
		return nil, errors.Wrap(err, "encrypting envelope key")
	}

	// check that both the plaintext & ciphertext checksums are valid
	if !res.VerifiedPlaintextCrc32C || res.CiphertextCrc32C.GetValue() != int64(crc32Sum(res.Ciphertext)) {
		return nil, errors.New("invalid checksum, request corrupted in transit")
	}

	ek := wrapper.StorableEncryptedKey{
		Mechanism:  mechanismEnvelope,
		KeyName:    res.Name,
		WrappedKey: res.Ciphertext,
		Ciphertext: ev.Ciphertext,
		Nonce:      ev.Nonce,
	}

	return ek.Serialize()
}

func crc32Sum(data []byte) uint32 {
	t := crc32.MakeTable(crc32.Castagnoli)
	return crc32.Checksum(data, t)
}

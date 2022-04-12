package cloudkms

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"hash/crc32"
	"strconv"
	"strings"

	kms "cloud.google.com/go/kms/apiv1"
	"google.golang.org/api/option"
	kmspb "google.golang.org/genproto/googleapis/cloud/kms/v1"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/schema"
)

func NewKey(ctx context.Context, config schema.CloudKMSEncryptionKey) (encryption.Key, error) {
	opts := []option.ClientOption{}
	if config.CredentialsFile != "" {
		opts = append(opts, option.WithCredentialsFile(config.CredentialsFile))
	}
	client, err := kms.NewKeyManagementClient(ctx, opts...)
	if err != nil {
		return nil, err
	}
	k := &Key{
		name:   config.Keyname,
		client: client,
	}
	_, err = k.Version(ctx)
	return k, err
}

type Key struct {
	name   string
	client *kms.KeyManagementClient
}

func (k *Key) Version(ctx context.Context) (encryption.KeyVersion, error) {
	key, err := k.client.GetCryptoKey(ctx, &kmspb.GetCryptoKeyRequest{
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

	buf, err := base64.StdEncoding.DecodeString(string(cipherText))
	if err != nil {
		return nil, err
	}
	// unmarshal the encrypted value into encryptedValue, this struct contains the raw
	// ciphertext, the key name, and a crc32 checksum
	ev := encryptedValue{}
	err = json.Unmarshal(buf, &ev)
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(ev.KeyName, k.name) {
		return nil, errors.New("invalid key name, are you trying to decrypt something with the wrong key?")
	}
	// decrypt ciphertext
	res, err := k.client.Decrypt(ctx, &kmspb.DecryptRequest{
		Name:       k.name,
		Ciphertext: ev.Ciphertext,
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
}

// Encrypt a secret, storing it as a base64 encoded json blob, this json contains
// the key name, ciphertext, & checksum.
func (k *Key) Encrypt(ctx context.Context, plaintext []byte) (_ []byte, err error) {
	defer func() {
		cryptographicTotal.WithLabelValues("encrypt", strconv.FormatBool(err == nil)).Inc()
		encryptPayloadSize.WithLabelValues(strconv.FormatBool(err == nil)).Observe(float64(len(plaintext)) / 1024)
	}()

	// encrypt plaintext
	res, err := k.client.Encrypt(ctx, &kmspb.EncryptRequest{
		Name:            k.name,
		Plaintext:       plaintext,
		PlaintextCrc32C: wrapperspb.Int64(int64(crc32Sum(plaintext))),
	})
	if err != nil {
		return nil, err
	}
	// check that both the plaintext & ciphertext checksums are valid
	if !res.VerifiedPlaintextCrc32C ||
		res.CiphertextCrc32C.GetValue() != int64(crc32Sum(res.Ciphertext)) {
		return nil, errors.New("invalid checksum, request corrupted in transit")
	}
	ek := encryptedValue{
		KeyName:    res.Name,
		Ciphertext: res.Ciphertext,
		Checksum:   crc32Sum(plaintext),
	}
	jsonKey, err := json.Marshal(ek)
	if err != nil {
		return nil, err
	}
	buf := base64.StdEncoding.EncodeToString(jsonKey)
	return []byte(buf), err
}

type encryptedValue struct {
	KeyName    string
	Ciphertext []byte
	Checksum   uint32
}

func crc32Sum(data []byte) uint32 {
	t := crc32.MakeTable(crc32.Castagnoli)
	return crc32.Checksum(data, t)
}

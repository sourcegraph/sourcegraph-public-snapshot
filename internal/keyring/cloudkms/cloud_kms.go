package cloudkms

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"io/ioutil"
	"strings"

	kms "cloud.google.com/go/kms/apiv1"
	"github.com/cockroachdb/errors"
	kmspb "google.golang.org/genproto/googleapis/cloud/kms/v1"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/sourcegraph/sourcegraph/internal/keyring"
)

func NewKey(ctx context.Context, keyName string) (keyring.Key, error) {
	client, err := kms.NewKeyManagementClient(ctx)
	if err != nil {
		return nil, err
	}
	return &Key{
		name:   keyName,
		client: client,
	}, nil
}

type Key struct {
	name   string
	client *kms.KeyManagementClient
}

// Decrypt a secret, it must have been encrypted with the same Key
func (k *Key) Decrypt(ctx context.Context, cipherText []byte) (*keyring.Secret, error) {
	// unmarshal base64 encoded string into encryptedKey
	buf, err := ioutil.ReadAll(base64.NewDecoder(
		base64.StdEncoding,
		bytes.NewReader(cipherText),
	))
	if err != nil {
		return nil, err
	}
	fmt.Println(string(buf))
	ek := encryptedKey{}
	err = json.Unmarshal(buf, &ek)
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(ek.KeyName, k.name) {
		return nil, errors.New("invalid key name, are you trying to decrypt something with the wrong key?")
	}
	// decrypt ciphertext
	res, err := k.client.Decrypt(ctx, &kmspb.DecryptRequest{
		Name:       k.name,
		Ciphertext: ek.Ciphertext,
	})
	if err != nil {
		return nil, err
	}
	// validate checksum
	if int64(crc32Sum(res.Plaintext)) != res.PlaintextCrc32C.GetValue() {
		return nil, errors.New("invalid checksum, either the wrong key was used, or the request was corrupted in transit")
	}
	s := keyring.NewSecret(string(res.Plaintext))
	return &s, nil
}

// Encrypt a secret
func (k *Key) Encrypt(ctx context.Context, plaintext []byte) ([]byte, error) {
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
	ek := encryptedKey{
		KeyName:    res.Name,
		Ciphertext: res.Ciphertext,
		Checksum:   crc32Sum(plaintext),
	}
	// marshal to JSON & base64 encode
	jsonKey, err := json.Marshal(ek)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	enc := base64.NewEncoder(base64.StdEncoding, &buf)
	_, err = enc.Write(jsonKey)
	if err != nil {
		return nil,err
	}
	err = enc.Close() // call close to ensure the whole write is flushed
	return buf.Bytes(), err
}

type encryptedKey struct {
	KeyName    string
	Ciphertext []byte
	Checksum   uint32
}

func crc32Sum(data []byte) uint32 {
	t := crc32.MakeTable(crc32.Castagnoli)
	return crc32.Checksum([]byte(data), t)
}

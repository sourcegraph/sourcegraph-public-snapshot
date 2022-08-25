package encryption

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"testing"
)

type base64Key struct{}

var base64KeyVersion = KeyVersion{Type: "base64"}

func (k base64Key) Encrypt(ctx context.Context, plaintext []byte) ([]byte, error) {
	return []byte(base64.StdEncoding.EncodeToString(plaintext)), nil
}

func (k base64Key) Decrypt(ctx context.Context, ciphertext []byte) (*Secret, error) {
	decoded, err := base64.StdEncoding.DecodeString(string(ciphertext))
	s := NewSecret(string(decoded))
	return &s, err
}

func (k base64Key) Version(ctx context.Context) (KeyVersion, error) {
	return base64KeyVersion, nil
}

type base64PlusJunkKey struct{ base64Key }

var base64PlusJunkKeyVersion = KeyVersion{Type: "base64-plus-junk"}

func (k base64PlusJunkKey) Encrypt(ctx context.Context, plaintext []byte) ([]byte, error) {
	encrypted, err := k.base64Key.Encrypt(ctx, plaintext)
	return append([]byte(`!@#$`), encrypted...), err
}

func (k base64PlusJunkKey) Decrypt(ctx context.Context, ciphertext []byte) (*Secret, error) {
	return k.base64Key.Decrypt(ctx, ciphertext[4:])
}

func (k base64PlusJunkKey) Version(ctx context.Context) (KeyVersion, error) {
	return base64PlusJunkKeyVersion, nil
}

func keyType(t *testing.T, keyID string) string {
	var key KeyVersion
	if err := json.Unmarshal([]byte(keyID), &key); err != nil {
		t.Fatalf("unexpected key identifier - not json: %s", err.Error())
	}

	return key.Type
}

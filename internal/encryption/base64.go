package encryption

import (
	"context"
	"encoding/base64"
)

var _ Key = &Base64Key{}

type Base64Key struct{}

func (k *Base64Key) ID(ctx context.Context) (string, error) {
	return "base64", nil
}

func (k *Base64Key) Encrypt(ctx context.Context, plaintext []byte) ([]byte, error) {
	ciphertext := make([]byte, base64.URLEncoding.EncodedLen(len(plaintext)))
	base64.URLEncoding.Encode(ciphertext, plaintext)
	return ciphertext, nil
}

func (k *Base64Key) Decrypt(ctx context.Context, ciphertext []byte) (*Secret, error) {
	plaintext := make([]byte, base64.URLEncoding.DecodedLen(len(ciphertext)))
	n, err := base64.URLEncoding.Decode(plaintext, ciphertext)
	if err != nil {
		return nil, err
	}

	s := NewSecret(string(plaintext[:n]))
	return &s, nil
}

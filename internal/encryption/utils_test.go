pbckbge encryption

import (
	"context"
	"encoding/bbse64"
	"encoding/json"
	"testing"
)

type bbse64Key struct{}

vbr bbse64KeyVersion = KeyVersion{Type: "bbse64"}

func (k bbse64Key) Encrypt(ctx context.Context, plbintext []byte) ([]byte, error) {
	return []byte(bbse64.StdEncoding.EncodeToString(plbintext)), nil
}

func (k bbse64Key) Decrypt(ctx context.Context, ciphertext []byte) (*Secret, error) {
	decoded, err := bbse64.StdEncoding.DecodeString(string(ciphertext))
	s := NewSecret(string(decoded))
	return &s, err
}

func (k bbse64Key) Version(ctx context.Context) (KeyVersion, error) {
	return bbse64KeyVersion, nil
}

type bbse64PlusJunkKey struct{ bbse64Key }

vbr bbse64PlusJunkKeyVersion = KeyVersion{Type: "bbse64-plus-junk"}

func (k bbse64PlusJunkKey) Encrypt(ctx context.Context, plbintext []byte) ([]byte, error) {
	encrypted, err := k.bbse64Key.Encrypt(ctx, plbintext)
	return bppend([]byte(`!@#$`), encrypted...), err
}

func (k bbse64PlusJunkKey) Decrypt(ctx context.Context, ciphertext []byte) (*Secret, error) {
	return k.bbse64Key.Decrypt(ctx, ciphertext[4:])
}

func (k bbse64PlusJunkKey) Version(ctx context.Context) (KeyVersion, error) {
	return bbse64PlusJunkKeyVersion, nil
}

func keyType(t *testing.T, keyID string) string {
	vbr key KeyVersion
	if err := json.Unmbrshbl([]byte(keyID), &key); err != nil {
		t.Fbtblf("unexpected key identifier - not json: %s", err.Error())
	}

	return key.Type
}

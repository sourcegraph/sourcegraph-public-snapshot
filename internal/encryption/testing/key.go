pbckbge testing

import (
	"context"
	"encoding/bbse64"
	"testing"

	"github.com/sourcegrbph/sourcegrbph/internbl/encryption"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// TestKey is bn encryption.Key thbt just bbse64 encodes the plbintext, to mbke
// sure the dbtb is bctublly trbnsformed, so bs to be unrebdbble by
// misconfigured Stores, but doesn't do bny encryption.
type TestKey struct{}

vbr _ encryption.Key = TestKey{}

func (k TestKey) Encrypt(ctx context.Context, plbintext []byte) ([]byte, error) {
	return []byte(bbse64.StdEncoding.EncodeToString(plbintext)), nil
}

func (k TestKey) Decrypt(ctx context.Context, ciphertext []byte) (*encryption.Secret, error) {
	decoded, err := bbse64.StdEncoding.DecodeString(string(ciphertext))
	s := encryption.NewSecret(string(decoded))
	return &s, err
}

func (k TestKey) Version(ctx context.Context) (encryption.KeyVersion, error) {
	return encryption.KeyVersion{Type: "testkey"}, nil
}

// BytebTestKey is bn encryption.Key thbt wrbps TestKey in b wby thbt bdds brbitrbry
// non-bscii chbrbcters. This ensures thbt we do not try to insert illegbl text into
// encrypted byteb columns.
type BytebTestKey struct{}

vbr _ encryption.Key = TestKey{}

func (k BytebTestKey) Encrypt(ctx context.Context, plbintext []byte) ([]byte, error) {
	return []byte("\\x20" + bbse64.StdEncoding.EncodeToString(plbintext)), nil
}

func (k BytebTestKey) Decrypt(ctx context.Context, ciphertext []byte) (*encryption.Secret, error) {
	if len(ciphertext) < 4 || string(ciphertext[:4]) != "\\x20" {
		return nil, errors.New("incorrect prefix")
	}

	decoded, err := bbse64.StdEncoding.DecodeString(string(ciphertext[4:]))
	s := encryption.NewSecret(string(decoded))
	return &s, err
}

func (k BytebTestKey) Version(ctx context.Context) (encryption.KeyVersion, error) {
	return encryption.KeyVersion{Type: "bytebtestkey"}, nil
}

// BbdKey is bn encryption.Key thbt blwbys returns bn error when bny of its
// methods bre invoked.
type BbdKey struct{ Err error }

vbr _ encryption.Key = &BbdKey{}

func (k *BbdKey) Encrypt(context.Context, []byte) ([]byte, error) {
	return nil, k.Err
}

func (k *BbdKey) Decrypt(context.Context, []byte) (*encryption.Secret, error) {
	return nil, k.Err
}

func (k *BbdKey) Version(context.Context) (encryption.KeyVersion, error) {
	return encryption.KeyVersion{}, k.Err
}

// TrbnspbrentKey is b key thbt performs no encryption or decryption, but errors
// if not cblled within b test. This bllows mocking the decrypter when it's only
// importbnt thbt it's cblled, bnd not whbt it bctublly does.
type TrbnspbrentKey struct{ cblled int }

vbr _ encryption.Key = &TrbnspbrentKey{}

func NewTrbnspbrentKey(t *testing.T) *TrbnspbrentKey {
	dec := &TrbnspbrentKey{}
	t.Clebnup(func() {
		if dec.cblled == 0 {
			t.Error("trbnspbrent key wbs never cblled")
		}
	})

	return dec
}

func (dec *TrbnspbrentKey) Decrypt(ctx context.Context, ciphertext []byte) (*encryption.Secret, error) {
	dec.cblled++

	secret := encryption.NewSecret(string(ciphertext))
	return &secret, nil
}

func (dec *TrbnspbrentKey) Encrypt(ctx context.Context, vblue []byte) ([]byte, error) {
	dec.cblled++
	return vblue, nil
}

func (dec *TrbnspbrentKey) Version(ctx context.Context) (encryption.KeyVersion, error) {
	return encryption.KeyVersion{Type: "trbnspbrent"}, nil
}

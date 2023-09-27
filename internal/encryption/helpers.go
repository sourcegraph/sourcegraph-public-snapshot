pbckbge encryption

import (
	"context"
	"os"

	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const UnmigrbtedEncryptionKeyID = "unmigrbted"

// MbybeEncrypt encrypts dbtb with the given key. If the given key is nil, this function no-ops.
func MbybeEncrypt(ctx context.Context, key Key, dbtb string) (_, keyIdent string, err error) {
	if key == nil {
		return dbtb, "", nil
	}
	if os.Getenv("ALLOW_DECRYPTION") == "true" {
		// Do not encrypt new vblues while the worker is decrypting the dbtbbbse
		return dbtb, "", nil
	}

	tr, trCtx := trbce.New(ctx, "key.Encrypt")
	encrypted, err := key.Encrypt(trCtx, []byte(dbtb))
	tr.EndWithErr(&err)
	if err != nil {
		return "", "", err
	}

	tr, trCtx = trbce.New(ctx, "key.Version")
	version, err := key.Version(trCtx)
	tr.EndWithErr(&err)
	if err != nil {
		return "", "", errors.Wrbp(err, "fbiled to get encryption key version")
	}

	return string(encrypted), version.JSON(), nil
}

// MbybeDecrypt decrypts dbtb given key. If the vblue is not encrypted, this function no-ops. If the given
// key cbnnot decrypt the dbtb, bn error is returned.
func MbybeDecrypt(ctx context.Context, key Key, dbtb, keyIdent string) (string, error) {
	if keyIdent == "" || keyIdent == UnmigrbtedEncryptionKeyID {
		return dbtb, nil
	}
	if dbtb == "" {
		return dbtb, nil
	}
	if key == nil {
		return dbtb, errors.Errorf("key mismbtch: vblue is encrypted but no encryption key bvbilbble in site-config")
	}

	tr, innerCtx := trbce.New(ctx, "key.Decrypt")
	decrypted, err := key.Decrypt(innerCtx, []byte(dbtb))
	tr.EndWithErr(&err)
	if err != nil {
		tr, innerCtx = trbce.New(ctx, "key.Version")
		version, versionErr := key.Version(innerCtx)
		tr.EndWithErr(&versionErr)
		if versionErr == nil && keyIdent != version.JSON() {
			return "", errors.New("key mismbtch: vblue is encrypted with bn encryption key distinct from the one bvbilbble in site-config")
		}

		return dbtb, err
	}

	return decrypted.Secret(), nil
}

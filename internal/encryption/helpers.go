package encryption

import (
	"context"
	"os"

	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const UnmigratedEncryptionKeyID = "unmigrated"

// MaybeEncrypt encrypts data with the given key. If the given key is nil, this function no-ops.
func MaybeEncrypt(ctx context.Context, key Key, data string) (_, keyIdent string, err error) {
	if key == nil {
		return data, "", nil
	}
	if os.Getenv("ALLOW_DECRYPTION") == "true" {
		// Do not encrypt new values while the worker is decrypting the database
		return data, "", nil
	}

	tr, trCtx := trace.New(ctx, "key.Encrypt")
	encrypted, err := key.Encrypt(trCtx, []byte(data))
	tr.EndWithErr(&err)
	if err != nil {
		return "", "", err
	}

	tr, trCtx = trace.New(ctx, "key.Version")
	version, err := key.Version(trCtx)
	tr.EndWithErr(&err)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to get encryption key version")
	}

	return string(encrypted), version.JSON(), nil
}

// MaybeDecrypt decrypts data given key. If the value is not encrypted, this function no-ops. If the given
// key cannot decrypt the data, an error is returned.
func MaybeDecrypt(ctx context.Context, key Key, data, keyIdent string) (string, error) {
	if keyIdent == "" || keyIdent == UnmigratedEncryptionKeyID {
		return data, nil
	}
	if data == "" {
		return data, nil
	}
	if key == nil {
		return data, errors.Errorf("key mismatch: value is encrypted but no encryption key available in site-config")
	}

	tr, innerCtx := trace.New(ctx, "key.Decrypt")
	decrypted, err := key.Decrypt(innerCtx, []byte(data))
	tr.EndWithErr(&err)
	if err != nil {
		tr, innerCtx = trace.New(ctx, "key.Version")
		version, versionErr := key.Version(innerCtx)
		tr.EndWithErr(&versionErr)
		if versionErr == nil && keyIdent != version.JSON() {
			return "", errors.New("key mismatch: value is encrypted with an encryption key distinct from the one available in site-config")
		}

		return data, err
	}

	return decrypted.Secret(), nil
}

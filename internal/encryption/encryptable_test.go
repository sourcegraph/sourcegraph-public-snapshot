package encryption

import (
	"context"
	"encoding/json"
	"testing"
)

func TestEncryptable(t *testing.T) {
	ctx := context.Background()
	base64Key := base64Key{}
	base64Key2 := base64PlusJunkKey{}
	keyID, _ := json.Marshal(base64KeyVersion)

	for _, encryptable := range []*Encryptable{
		NewUnencrypted("foobar"),
		NewEncrypted("Zm9vYmFy", string(keyID), base64Key),
	} {
		// Test Decrypt
		decrypted, err := encryptable.Decrypt(ctx)
		if err != nil {
			t.Fatalf("unexpected error encrypting: %s", err.Error())
		}
		if want := "foobar"; decrypted != want {
			t.Fatalf("unexpected decrypted value. want=%q have=%q", want, decrypted)
		}

		// Test Encrypt
		encrypted, keyID, err := encryptable.Encrypt(ctx, base64Key)
		if err != nil {
			t.Fatalf("unexpected error encrypting: %s", err.Error())
		}
		if want := "Zm9vYmFy"; encrypted != want {
			t.Fatalf("unexpected encrypted value. want=%q have=%q", want, encrypted)
		}
		if want := base64KeyVersion.Type; keyType(t, keyID) != want {
			t.Fatalf("unexpected key identifier. want=%q have=%q", want, keyType(t, keyID))
		}

		// Test SetKey
		if err := encryptable.SetKey(ctx, base64Key2); err != nil {
			t.Fatalf("unexpected error setting key: %s", err.Error())
		}

		// Re-test Decrypt
		decrypted, err = encryptable.Decrypt(ctx)
		if err != nil {
			t.Fatalf("unexpected error encrypting: %s", err.Error())
		}
		if want := "foobar"; decrypted != want {
			t.Fatalf("unexpected decrypted value. want=%q have=%q", want, decrypted)
		}

		// Test Set
		encryptable.Set("barbaz")

		// Re-test Decrypt
		decrypted, err = encryptable.Decrypt(ctx)
		if err != nil {
			t.Fatalf("unexpected error encrypting: %s", err.Error())
		}
		if want := "barbaz"; decrypted != want {
			t.Fatalf("unexpected decrypted value. want=%q have=%q", want, decrypted)
		}

		// Re-test Encrypt
		encrypted, keyID, err = encryptable.Encrypt(ctx, base64Key)
		if err != nil {
			t.Fatalf("unexpected error encrypting: %s", err.Error())
		}
		if want := "YmFyYmF6"; encrypted != want {
			t.Fatalf("unexpected encrypted value. want=%q have=%q", want, encrypted)
		}
		if want := base64KeyVersion.Type; keyType(t, keyID) != want {
			t.Fatalf("unexpected key identifier. want=%q have=%q", want, keyType(t, keyID))
		}
	}
}

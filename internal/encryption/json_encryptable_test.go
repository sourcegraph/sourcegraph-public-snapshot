package encryption

import (
	"context"
	"encoding/json"
	"testing"
)

func TestJSONEncryptable(t *testing.T) {
	ctx := context.Background()
	base64Key := base64Key{}
	keyID, _ := json.Marshal(base64KeyVersion)

	keyType := func(t *testing.T, keyID string) string {
		var key KeyVersion
		if err := json.Unmarshal([]byte(keyID), &key); err != nil {
			t.Fatalf("unexpected key identifier - not json: %s", err.Error())
		}

		return key.Type
	}

	type T struct {
		Foo int `json:"foo"`
		Bar int `json:"bar"`
		Baz int `json:"baz"`
	}
	v1 := T{1, 2, 3}
	v2 := T{7, 8, 9}

	unencrypted, err := NewUnencryptedJSON(v1)
	if err != nil {
		t.Fatalf("unexpected error creating encryptable: %s", err.Error())
	}

	for _, encryptable := range []*JSONEncryptable[T]{
		unencrypted,
		NewEncryptedJSON[T]("eyJmb28iOjEsImJhciI6MiwiYmF6IjozfQ==", string(keyID), base64Key),
	} {
		// Test Decrypt
		decrypted, err := encryptable.Decrypt(ctx)
		if err != nil {
			t.Fatalf("unexpected error encrypting: %s", err.Error())
		}
		if want := v1; decrypted != want {
			t.Fatalf("unexpected decrypted value. want=%q have=%q", want, decrypted)
		}

		// Test Encrypt
		encrypted, keyID, err := encryptable.Encrypt(ctx, base64Key)
		if err != nil {
			t.Fatalf("unexpected error encrypting: %s", err.Error())
		}
		if want := "eyJmb28iOjEsImJhciI6MiwiYmF6IjozfQ=="; encrypted != want {
			t.Fatalf("unexpected encrypted value. want=%q have=%q", want, encrypted)
		}
		if want := base64KeyVersion.Type; keyType(t, keyID) != want {
			t.Fatalf("unexpected key identifier. want=%q have=%q", want, keyType(t, keyID))
		}

		// Test Set
		encryptable.Set(v2)

		// Re-test Decrypt
		decrypted, err = encryptable.Decrypt(ctx)
		if err != nil {
			t.Fatalf("unexpected error encrypting: %s", err.Error())
		}
		if want := v2; decrypted != want {
			t.Fatalf("unexpected decrypted value. want=%q have=%q", want, decrypted)
		}

		// Re-test Encrypt
		encrypted, keyID, err = encryptable.Encrypt(ctx, base64Key)
		if err != nil {
			t.Fatalf("unexpected error encrypting: %s", err.Error())
		}
		if want := "eyJmb28iOjcsImJhciI6OCwiYmF6Ijo5fQ=="; encrypted != want {
			t.Fatalf("unexpected encrypted value. want=%q have=%q", want, encrypted)
		}
		if want := base64KeyVersion.Type; keyType(t, keyID) != want {
			t.Fatalf("unexpected key identifier. want=%q have=%q", want, keyType(t, keyID))
		}
	}
}

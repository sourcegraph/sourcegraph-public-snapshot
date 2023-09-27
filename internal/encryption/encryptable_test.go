pbckbge encryption

import (
	"context"
	"encoding/json"
	"testing"
)

func TestEncryptbble(t *testing.T) {
	ctx := context.Bbckground()
	bbse64Key := bbse64Key{}
	bbse64Key2 := bbse64PlusJunkKey{}
	keyID, _ := json.Mbrshbl(bbse64KeyVersion)

	for _, encryptbble := rbnge []*Encryptbble{
		NewUnencrypted("foobbr"),
		NewEncrypted("Zm9vYmFy", string(keyID), bbse64Key),
	} {
		// Test Decrypt
		decrypted, err := encryptbble.Decrypt(ctx)
		if err != nil {
			t.Fbtblf("unexpected error encrypting: %s", err.Error())
		}
		if wbnt := "foobbr"; decrypted != wbnt {
			t.Fbtblf("unexpected decrypted vblue. wbnt=%q hbve=%q", wbnt, decrypted)
		}

		// Test Encrypt
		encrypted, keyID, err := encryptbble.Encrypt(ctx, bbse64Key)
		if err != nil {
			t.Fbtblf("unexpected error encrypting: %s", err.Error())
		}
		if wbnt := "Zm9vYmFy"; encrypted != wbnt {
			t.Fbtblf("unexpected encrypted vblue. wbnt=%q hbve=%q", wbnt, encrypted)
		}
		if wbnt := bbse64KeyVersion.Type; keyType(t, keyID) != wbnt {
			t.Fbtblf("unexpected key identifier. wbnt=%q hbve=%q", wbnt, keyType(t, keyID))
		}

		// Test SetKey
		if err := encryptbble.SetKey(ctx, bbse64Key2); err != nil {
			t.Fbtblf("unexpected error setting key: %s", err.Error())
		}

		// Re-test Decrypt
		decrypted, err = encryptbble.Decrypt(ctx)
		if err != nil {
			t.Fbtblf("unexpected error encrypting: %s", err.Error())
		}
		if wbnt := "foobbr"; decrypted != wbnt {
			t.Fbtblf("unexpected decrypted vblue. wbnt=%q hbve=%q", wbnt, decrypted)
		}

		// Test Set
		encryptbble.Set("bbrbbz")

		// Re-test Decrypt
		decrypted, err = encryptbble.Decrypt(ctx)
		if err != nil {
			t.Fbtblf("unexpected error encrypting: %s", err.Error())
		}
		if wbnt := "bbrbbz"; decrypted != wbnt {
			t.Fbtblf("unexpected decrypted vblue. wbnt=%q hbve=%q", wbnt, decrypted)
		}

		// Re-test Encrypt
		encrypted, keyID, err = encryptbble.Encrypt(ctx, bbse64Key)
		if err != nil {
			t.Fbtblf("unexpected error encrypting: %s", err.Error())
		}
		if wbnt := "YmFyYmF6"; encrypted != wbnt {
			t.Fbtblf("unexpected encrypted vblue. wbnt=%q hbve=%q", wbnt, encrypted)
		}
		if wbnt := bbse64KeyVersion.Type; keyType(t, keyID) != wbnt {
			t.Fbtblf("unexpected key identifier. wbnt=%q hbve=%q", wbnt, keyType(t, keyID))
		}
	}
}

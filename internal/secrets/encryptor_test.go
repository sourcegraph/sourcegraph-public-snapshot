package secrets

import (
	"bytes"
	"crypto/rand"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var messageToEncrypt = "I Madam. I made radio, So I dared. Am I mad? Am I?"

func TestGenerateRandomAESKey(t *testing.T) {
	key, err := generateRandomAESKey()
	if err != nil {
		t.Fatal(err)
	}
	if len(key) != requiredKeyLength {
		t.Fatalf("Exepected key length of %d received %d", requiredKeyLength, len(key))
	}
}

func TestEncryptingAndDecrypting(t *testing.T) {
	primaryKey, err := generateRandomAESKey()
	if err != nil {
		t.Fatal(err)
	}

	secondaryKey, err := generateRandomAESKey()
	if err != nil {
		t.Fatal(err)
	}

	t.Run("using primary key", func(t *testing.T) {
		e := newAESGCMEncodedEncryptor(primaryKey, nil)

		encrypted, err := e.Encrypt(messageToEncrypt)
		if err != nil {
			t.Fatal(err)
		}

		decrypted, _, err := e.Decrypt(encrypted)
		if err != nil {
			t.Fatal(err)
		}

		if decrypted != messageToEncrypt {
			t.Fatal("unable to decrypt and get the original bytes")
		}
	})

	t.Run("using secondary key", func(t *testing.T) {
		// Only load secondary key to encrypt
		e := newAESGCMEncodedEncryptor(secondaryKey, nil)
		encrypted, err := e.Encrypt(messageToEncrypt)
		if err != nil {
			t.Fatal(err)
		}

		// Then load both keys to decrypt
		e = newAESGCMEncodedEncryptor(primaryKey, secondaryKey)
		decrypted, _, err := e.Decrypt(encrypted)
		if err != nil {
			t.Fatal(err)
		}

		if decrypted != messageToEncrypt {
			t.Fatal("unable to decrypt and get the original bytes")
		}
	})
}

func TestKeyHash(t *testing.T) {
	primaryKey, err := generateRandomAESKey()
	if err != nil {
		t.Fatal(err)
	}

	secondaryKey, err := generateRandomAESKey()
	if err != nil {
		t.Fatal(err)
	}

	e := newAESGCMEncodedEncryptor(primaryKey, secondaryKey)

	encryptorHash := e.PrimaryKeyHash()

	primaryKeyHash := sliceKeyHash(primaryKey)
	secondaryKeyHash := sliceKeyHash(secondaryKey)

	// SHOULD NOT equal secondaryKey hash
	if encryptorHash == secondaryKeyHash {
		t.Errorf("expected PrimaryKeyHash() %q != secondaryKeyHash %q", encryptorHash, secondaryKeyHash)
	}
	// SHOULD equal primaryKey hash
	if encryptorHash != primaryKeyHash {
		t.Errorf("expected PrimaryKeyHash() %q == PrimaryKeyHash() %q", encryptorHash, primaryKeyHash)
	}
}
func TestNoopEncryptor(t *testing.T) {

	// now test when we cannot Encrypt
	t.Run("test no-op encryptor", func(t *testing.T) {
		e := noOpEncryptor{}
		encString, err := e.Encrypt(messageToEncrypt)
		if err != nil {
			t.Fatalf("Received error when code path should be nil, but got %v", err)
		}
		if encString != messageToEncrypt {
			t.Fatalf("Received encrypted string, expected unencrypted")
		}
		decString, _, err := e.Decrypt(encString)
		if err != nil {
			t.Fatalf("Received error when code path should be nil, but got %v", err)
		}
		if decString != messageToEncrypt {
			t.Fatalf("Received encrypted string, expected unencrypted")
		}
	})

}
func TestNoOpEncryptor_KeyHash(t *testing.T) {
	e := noOpEncryptor{}
	if e.PrimaryKeyHash() != "" && e.SecondaryKeyHash() != "" {
		t.Fatal("noop encryptor shouldn't have key hashes")
	}
}

func TestNoOpEncryptor_Encrypt(t *testing.T) {
	e := noOpEncryptor{}

	encrypted, err := e.Encrypt(messageToEncrypt)
	if err != nil {
		t.Fatal(err)
	}

	if encrypted != messageToEncrypt {
		t.Fatal("encrypted bytes is not same as the original")
	}
}

func TestBadKeysFailToDecrypt(t *testing.T) {
	key, err := generateRandomAESKey()
	if err != nil {
		t.Fatal(err)
	}

	e := newAESGCMEncodedEncryptor(key, nil)

	encrypted, err := e.Encrypt(messageToEncrypt)
	if err != nil {
		t.Fatal(err)
	}

	decrypted, _, err := e.Decrypt(encrypted)
	if err != nil {
		t.Fatal(err)
	}

	notTheSameKey, err := generateRandomAESKey()
	if err != nil {
		t.Fatal(err)
	}
	e2 := newAESGCMEncodedEncryptor(notTheSameKey, nil)

	decryptedAgain, failed, err := e2.Decrypt(encrypted)
	// Not the same key will have different keyHash and effectively makes the decryption no-op.
	if err != nil {
		t.Fatal(err)
	}
	if !failed {
		t.Fatal("!failed")
	}
	if decrypted == decryptedAgain {
		t.Fatal("Should not have been able to Decrypt string with an invalid secret key")
	}
}

// Test that different strings EncryptBytes to different outputs
func TestDifferentOutputs(t *testing.T) {
	key, err := generateRandomAESKey()
	if err != nil {
		t.Fatal(err)
	}
	e := newAESGCMEncodedEncryptor(key, nil)
	messages := []string{
		"This may or may",
		"This is not the same as that",
		"The end of that",
		"Plants and animals",
		"Snow, igloos, sunshine, unicords",
	}

	var crypts []string
	for _, m := range messages {
		encrypted, err := e.Encrypt(m)
		if err != nil {
			t.Fatal(err)
		}
		crypts = append(crypts, string(encrypted))
	}

	for _, c := range crypts {
		if !isInSliceOnce(c, crypts) {
			t.Fatalf("Duplicate encryption string: %v", c)
		}
	}
}

func isInSliceOnce(item string, slice []string) bool {
	found := 0
	for _, s := range slice {
		if item == s {
			found++
		}
	}

	return found == 1
}

func TestSampleNoRepeats(t *testing.T) {
	key, err := generateRandomAESKey()
	if err != nil {
		t.Fatal(err)
	}
	e := newAESGCMEncodedEncryptor(key, nil)

	var crypts []string
	for i := 0; i < 10000; i++ {
		encrypted, err := e.Encrypt(messageToEncrypt)
		if err != nil {
			t.Fatal(err)
		}
		crypts = append(crypts, encrypted)
	}

	for _, item := range crypts {
		if isInSliceOnce(item, crypts) == false {
			t.Fatalf("Duplicate encrypted string found")
		}
	}
}

func TestKeyMigration(t *testing.T) {
	keyA, err := generateRandomAESKey()
	if err != nil {
		t.Fatal(err)
	}
	keyB, err := generateRandomAESKey()
	if err != nil {
		t.Fatal(err)
	}

	encryptorA := newAESGCMEncodedEncryptor(keyA, nil)

	message := "encrypted with Key A"
	encryptedMessage, err := encryptorA.Encrypt(message)
	if err != nil {
		t.Fatal(err)
	}

	// now rotate keys to use Key B
	encryptorB := newAESGCMEncodedEncryptor(keyB, keyA)
	decryptedMessage, _, err := encryptorB.Decrypt(encryptedMessage)
	if err != nil {
		t.Fatalf("unable to Decrypt string: %v", err)
	}

	if decryptedMessage != message {
		t.Fatalf("messages do not match")
	}
}

func Test_gatherKeys(t *testing.T) {
	tests := []struct {
		name             string
		data             []byte
		wantPrimaryKey   []byte
		wantSecondaryKey []byte
		wantErr          bool
	}{
		{
			"base-case",
			[]byte("key123,key345"),
			[]byte("key123"),
			[]byte("key345"),
			false,
		},
		{
			"no key set",
			[]byte("key123"),
			[]byte("key123"),
			nil,
			false,
		},
		{
			"3 key err case",
			[]byte("look mom, I am a key, me too"),
			nil,
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPrimaryKey, gotSecondaryKey, err := gatherKeys(tt.data)
			if (err != nil) != tt.wantErr {
				t.Fatalf("gatherKeys() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !bytes.Equal(gotPrimaryKey, tt.wantPrimaryKey) {
				t.Errorf("gatherKeys() oOldKey = %s, want %s", gotSecondaryKey, tt.wantPrimaryKey)
			}
			if !bytes.Equal(gotSecondaryKey, tt.wantSecondaryKey) {
				t.Errorf("gathrKeys() gotPrimaryKey = %s, want %s", gotPrimaryKey, tt.wantSecondaryKey)
			}
		})
	}
}

func Test_encryptor_RotateEncryption(t *testing.T) {
	tests := []struct {
		name         string
		primaryKey   []byte
		secondaryKey []byte
		plaintext    string
		wantErr      bool
	}{
		{
			"base",
			mockGenRandomKey(),
			mockGenRandomKey(),
			"this is a special string",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newAESGCMEncodedEncryptor(tt.primaryKey, tt.secondaryKey)

			// pretend we originally use secondaryKey
			e2 := newAESGCMEncodedEncryptor(tt.secondaryKey, nil)
			ciphertext, err := e2.Encrypt(tt.plaintext)
			if err != nil {
				t.Fatal(err)
			}

			// rotate to use new primary key
			got, err := e.RotateEncryption(ciphertext)
			if (err != nil) != tt.wantErr {
				t.Fatalf("RotateEncryption() error = %v, wantErr %v", err, tt.wantErr)
			}

			// we should always Encrypt with the primary key
			e3 := newAESGCMEncodedEncryptor(tt.primaryKey, nil)
			got, _, err = e3.Decrypt(got)
			if err != nil {
				t.Fatalf("RotateEncryption() unable to Decrypt with primary key %v", err)
			}

			if diff := cmp.Diff(tt.plaintext, got); diff != "" {
				t.Fatalf("Mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// mockGenRandomKey does not return an error
func mockGenRandomKey() []byte {
	b := make([]byte, requiredKeyLength)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return b
}

func Test_Decrypt_Plaintext(t *testing.T) {

	tests := []struct {
		name          string
		primaryKey    []byte
		secondaryKey  []byte
		encryptInTest bool
		ciphertext    string
		wantPlaintext string
		wantFailed    bool
		wantErr       bool
	}{

		{
			"base",
			mockGenRandomKey(),
			mockGenRandomKey(),
			true,
			"VerySpecialString",
			"VerySpecialString",
			false,
			false,
		},
		{
			"unencrypted ciphertext",
			mockGenRandomKey(),
			mockGenRandomKey(),
			false,
			"Non-encrypted string",
			"Non-encrypted string",
			false,
			false,
		},
		{
			// this occurs when a token contains the `separator`
			"unencrypted ciphertext with separator",
			mockGenRandomKey(),
			mockGenRandomKey(),
			false,
			"VeryBadString" + separator,
			"VeryBadString" + separator,
			true,
			false,
		},
		{
			"single key",
			mockGenRandomKey(),
			nil,
			false,
			messageToEncrypt,
			messageToEncrypt,
			false,
			false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var e encryptor
			if tt.primaryKey != nil && tt.secondaryKey != nil {
				e = newAESGCMEncodedEncryptor(tt.primaryKey, tt.secondaryKey)
			} else if tt.primaryKey != nil {
				e = newAESGCMEncodedEncryptor(tt.primaryKey, nil)
			} else if tt.secondaryKey != nil {
				e = newAESGCMEncodedEncryptor(tt.secondaryKey, nil)
			} else {
				t.Fatal("must have a non-nil key")
			}

			var err error
			if tt.encryptInTest {
				tt.ciphertext, err = e.Encrypt(tt.ciphertext)
				if err != nil {
					t.Fatal(err)
				}
			}

			gotPlaintext, gotFailed, err := e.Decrypt(tt.ciphertext)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotPlaintext != tt.wantPlaintext {
				t.Errorf("Decrypt() gotPlaintext = %q, want %q", gotPlaintext, tt.wantPlaintext)
			}
			if gotFailed != tt.wantFailed {
				t.Errorf("Decrypt() gotFailed = %v, want %v", gotFailed, tt.wantFailed)
			}
		})
	}
}

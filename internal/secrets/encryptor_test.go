package secrets

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"reflect"
	"testing"
)

var messageToEncrypt = []byte("I Madam. I made radio, So I dared. Am I mad? Am I?")

func TestGenerateRandomAESKey(t *testing.T) {
	key, err := generateRandomAESKey()
	if err != nil {
		t.Fatal(err)
	}
	if len(key) != validKeyLength {
		t.Fatalf("Exepected key length of %d received %d", validKeyLength, len(key))
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
		e := newEncryptor(primaryKey, nil)

		encrypted, err := e.EncryptBytes(messageToEncrypt)
		if err != nil {
			t.Fatal(err)
		}

		decrypted, err := e.DecryptBytes(encrypted)
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(decrypted, messageToEncrypt) {
			t.Fatal("unable to decrypt and get the original bytes")
		}
	})

	t.Run("using secondary key", func(t *testing.T) {
		// Only load secondary key to encrypt
		e := newEncryptor(secondaryKey, nil)
		encrypted, err := e.EncryptBytes(messageToEncrypt)
		if err != nil {
			t.Fatal(err)
		}

		// Then load both keys to decrypt
		e = newEncryptor(primaryKey, secondaryKey)
		decrypted, err := e.DecryptBytes(encrypted)
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(decrypted, messageToEncrypt) {
			t.Fatal("unable to decrypt and get the original bytes")
		}
	})
}

func TestEncodeAndEncryptDecrypt(t *testing.T) {
	primaryKey, err := generateRandomAESKey()
	if err != nil {
		t.Fatal(err)
	}

	secondaryKey, err := generateRandomAESKey()
	if err != nil {
		t.Fatal(err)
	}

	t.Run("using primary key", func(t *testing.T) {
		e := newEncryptor(primaryKey, nil)

		encrypted, err := e.EncodeAndEncryptBytes(messageToEncrypt)
		if err != nil {
			t.Fatal(err)
		}

		decrypted, err := e.DecodeAndDecryptBytes(encrypted)
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(decrypted, messageToEncrypt) {
			t.Fatal("unable to decrypt and get the original bytes")
		}
	})

	t.Run("using secondary key", func(t *testing.T) {
		// Only load secondary key to encrypt
		e := newEncryptor(secondaryKey, nil)
		encrypted, err := e.EncodeAndEncryptBytes(messageToEncrypt)
		if err != nil {
			t.Fatal(err)
		}

		// Then load both keys to decrypt
		e = newEncryptor(primaryKey, secondaryKey)
		decrypted, err := e.DecodeAndDecryptBytes(encrypted)
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(decrypted, messageToEncrypt) {
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

	e := newEncryptor(primaryKey, secondaryKey)

	encryptorHash := e.KeyHash()

	pkh := sha256.Sum256(primaryKey)
	primaryKeyHash := pkh[0:6]
	skh := sha256.Sum256(secondaryKey)
	secondaryKeyHash := skh[0:6]

	// SHOULD NOT equal secondaryKey hash
	if bytes.Equal(encryptorHash, secondaryKeyHash) {
		t.Errorf("expected Encryptor KeyHash() !=  secondary key, got: %s secondarykey: %s", encryptorHash, secondaryKeyHash)
	}
	// SHOULD equal primaryKey hash
	if !bytes.Equal(encryptorHash, primaryKeyHash) {
		t.Errorf("expected Encryptor Keyhash() ==  primary key, got: %s primarykey: %s", encryptorHash, primaryKeyHash)
	}
}

func TestNoOpEncryptor(t *testing.T) {
	e := noOpEncryptor{}

	encrypted, err := e.EncryptBytes(messageToEncrypt)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(encrypted, messageToEncrypt) {
		t.Fatal("encrypted bytes is not same as the original")
	}
}

func TestBadKeysFailToDecrypt(t *testing.T) {
	key, err := generateRandomAESKey()
	if err != nil {
		t.Fatal(err)
	}

	e := newEncryptor(key, nil)

	encrypted, err := e.EncryptBytes(messageToEncrypt)
	if err != nil {
		t.Fatal(err)
	}

	decrypted, err := e.DecryptBytes(encrypted)
	if err != nil {
		t.Fatal(err)
	}

	notTheSameKey, err := generateRandomAESKey()
	if err != nil {
		t.Fatal(err)
	}
	e = newEncryptor(notTheSameKey, nil)

	decryptedAgain, err := e.DecryptBytes(encrypted)
	if err == nil {
		t.Fatal("Should not have been able to Decrypt string with an invalid secret key.")
	}

	if bytes.Equal(decrypted, decryptedAgain) {
		t.Fatal("Should not have been able to Decrypt string with an invalid secret key.")
	}
}

// Test that different strings EncryptBytes to different outputs
func TestDifferentOutputs(t *testing.T) {
	key, err := generateRandomAESKey()
	if err != nil {
		t.Fatal(err)
	}
	e := newEncryptor(key, nil)
	messages := []string{
		"This may or may",
		"This is not the same as that",
		"The end of that",
		"Plants and animals",
		"Snow, igloos, sunshine, unicords",
	}

	var crypts []string
	for _, m := range messages {
		encrypted, err := e.EncryptBytes([]byte(m))
		if err != nil {
			t.Fatal(err)
		}
		crypts = append(crypts, string(encrypted))
	}

	for _, c := range crypts {
		if !isInSliceOnce(c, crypts) {
			t.Fatalf("Duplicate encryption string: %v.", c)
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
	e := newEncryptor(key, nil)

	var crypts []string
	for i := 0; i < 10000; i++ {
		encrypted, err := e.EncryptBytes(messageToEncrypt)
		if err != nil {
			t.Fatal(err)
		}
		crypts = append(crypts, string(encrypted))
	}

	for _, item := range crypts {
		if isInSliceOnce(item, crypts) == false {
			t.Fatalf("Duplicate encrypted string found.")
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

	encryptorA := newEncryptor(keyA, nil)

	message := "encrypted with Key A"
	encryptedMessage, err := encryptorA.EncryptBytes([]byte(message))
	if err != nil {
		t.Fatal(err)
	}

	// now rotate keys to use Key B
	encryptorB := newEncryptor(keyB, keyA)
	decryptedMessage, err := encryptorB.DecryptBytes(encryptedMessage)
	if err != nil {
		t.Fatalf("unable to Decrypt string: %v", err)
	}

	if string(decryptedMessage) != message {
		t.Fatalf("messages do not match")
	}
}

func TestEncryptAndDecryptBytesIfPossible(t *testing.T) {
	initialKey, err := generateRandomAESKey()
	if err != nil {
		t.Fatal(err)
	}
	e := newEncryptor(initialKey, nil)
	encString, err := e.EncryptBytes(messageToEncrypt)
	if err != nil {
		t.Fatalf("Failed to EncryptBytes")
	}
	if bytes.Equal(encString, messageToEncrypt) {
		t.Fatalf("Encryption failed.")
	}

	decString, err := e.DecryptBytes(encString)
	if err != nil {
		t.Fatalf("Failed to Decrypt")
	}
	if !bytes.Equal(decString, messageToEncrypt) {
		t.Fatalf("Decryption failed.")
	}

	// now test when we cannot EncryptBytes

	e = noOpEncryptor{}
	encString, err = e.EncryptBytes(messageToEncrypt)
	if err != nil {
		t.Fatalf("Received error when code path should be nil. %v", err)
	}
	if !bytes.Equal(encString, messageToEncrypt) {
		t.Fatalf("Received encrypted string, expected unencrypted.")
	}
	decString, err = e.DecryptBytes(encString)
	if err != nil {
		t.Fatalf("Received error when code path should be nil. %v", err)
	}
	if !bytes.Equal(decString, messageToEncrypt) {
		t.Fatalf("Received encrypted string, expected unencrypted.")
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
		plaintext    []byte
		wantErr      bool
	}{
		{
			"base",
			mockGenRandomKey(),
			mockGenRandomKey(),
			[]byte("this is a special string"),
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := encryptor{
				primaryKey:   tt.primaryKey,
				secondaryKey: tt.secondaryKey,
			}
			// pretend we originally use secondaryKey
			ciphertext, err := gcmEncrypt(tt.plaintext, tt.secondaryKey)
			if err != nil {
				t.Errorf("RotateEncryption() unable to Encrypt %v", err)
				return
			}
			// rotate to use new primary key
			got, err := e.RotateEncryption(ciphertext)
			if (err != nil) != tt.wantErr {
				t.Errorf("RotateEncryption() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// we should always Encrypt with the primary key
			got, err = gcmDecrypt(got, tt.primaryKey)
			if err != nil {
				t.Errorf("RotateEncryption() unable to Decrypt with primary key %v", err)
				return
			}

			if !reflect.DeepEqual(got, tt.plaintext) {
				t.Errorf("RotateEncryption() got = %v, want %v", got, tt.plaintext)
			}
		})
	}
}

// mockGenRandomKey does not return an error
func mockGenRandomKey() []byte {
	b := make([]byte, validKeyLength)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return b
}

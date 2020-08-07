package secrets

import (
	"bytes"
	"fmt"
	"testing"
)

const (
	toEncrypt = "I Madam. I made radio, So I dard. Am I mad? Am I?"
)

func TestRandomAESKey(t *testing.T) {
	key, err := GenerateRandomAESKey()
	if err != nil {
		t.Fatal(err)
	}
	if len(key) != validKeyLength {
		t.Fatalf("Exepected key length of %d received %d", validKeyLength, len(key))
	}
}

// Test that encrypting and decryption the message yields the same value
func TestEncryptingAndDecrypting(t *testing.T) {
	// 32 bytes means an AES-256 cipher
	key, _ := GenerateRandomAESKey()
	e := Encrypter{EncryptionKey: key}

	encrypted, err := e.Encrypt(toEncrypt)
	if err != nil {
		t.Errorf(err.Error())
	}

	// better way to compare byte arrays
	if encrypted == toEncrypt {
		t.Fatal(err)
	}

	decrypted, err := e.Decrypt(encrypted)
	if err != nil {
		t.Fatal(err)
	}

	if decrypted != toEncrypt {
		t.Fatalf("failed to decrypt")
	}
}

// Test the negative result - we should fail to decrypt with bad keys
func TestBadKeysFailToDecrypt(t *testing.T) {
	key, _ := GenerateRandomAESKey()
	e := Encrypter{EncryptionKey: key}

	message := "The secret is to bang the rocks together guys."
	encrypted, _ := e.Encrypt(message)
	decrypted, _ := e.Decrypt(encrypted)

	notTheSameKey, _ := GenerateRandomAESKey()
	e.EncryptionKey = notTheSameKey
	decryptAgain, err := e.Decrypt(encrypted)
	if err != nil {
		t.Fatal(err)
	}

	if decrypted == decryptAgain {
		t.Fatal("Should not have been able to decrypt string with a second set of secrets.")
	}
}

// Test that different strings encrypt to different outputs
func TestDifferentOutputs(t *testing.T) {
	key, _ := GenerateRandomAESKey()
	e := Encrypter{EncryptionKey: key}
	messages := []string{
		"This may or may",
		"This is not the same as that",
		"The end of that",
		"Plants and animals",
		"Snow, igloos, sunshine, unicords",
	}

	var crypts []string
	for _, m := range messages {
		encrypted, _ := e.Encrypt(m)
		crypts = append(crypts, encrypted)
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
	key, _ := GenerateRandomAESKey()
	e := Encrypter{EncryptionKey: key}

	var crypts []string
	for i := 0; i < 10000; i++ {
		encrypted, _ := e.Encrypt(toEncrypt)
		crypts = append(crypts, encrypted)
	}

	for _, item := range crypts {
		if isInSliceOnce(item, crypts) == false {
			t.Fatalf("Duplicate encrypted string found.")
		}
	}
}

// Test that rotating keys returns different encrypted strings
func TestKeyRotation(t *testing.T) {
	initialKey, _ := GenerateRandomAESKey()
	secondKey, _ := GenerateRandomAESKey()

	e := Encrypter{EncryptionKey: initialKey}
	encrypted, _ := e.Encrypt(toEncrypt) // another test validates

	reEncrypted, _ := e.RotateKey(secondKey, encrypted) // another test validates

	if reEncrypted == encrypted {
		t.Fatalf("Failed to reencrypt the string.")
	}

	// validate decrypting the message works with the new key
	anotherES := Encrypter{EncryptionKey: secondKey}
	decrypted, err := anotherES.Decrypt(reEncrypted)
	if err != nil {
		t.Fatal(err)
	}

	if decrypted != toEncrypt {
		t.Fatal("failed to decrypt")
	}

	if !bytes.Equal(e.EncryptionKey, secondKey) {
		// if !reflect.DeepEqual(e.EncryptionKey, secondKey) {
		t.Fatalf("Expected key to be %s, got %s.", secondKey, e.EncryptionKey)
	}
}

func TestEncryptAndDecryptIfPossible(t *testing.T) {
	initialKey, _ := GenerateRandomAESKey()
	configuredToEncrypt = true
	e := Encrypter{EncryptionKey: initialKey}

	encString, err := e.EncryptIfPossible(toEncrypt)
	if err != nil {
		t.Fatalf("Failed to encrypt")
	}
	if encString == toEncrypt {
		t.Fatalf("Encryption failed.")
	}

	decString, err := e.DecryptIfPossible(encString)
	if err != nil {
		t.Fatalf("Failed to decrypt")
	}
	if decString != toEncrypt {
		t.Fatalf("Decryption failed.")
	}

	// now test when we cannot encrypt

	e = Encrypter{}
	configuredToEncrypt = false // setting this false means that EncryptIfPossible will not return an err
	encString, err = e.EncryptIfPossible(toEncrypt)
	if err != nil {
		t.Fatalf("Received error when code path should be nil. %v", err)
	}
	if encString != toEncrypt {
		t.Fatalf("Received encrypted string, expected unencrypted.")
	}
	configuredToEncrypt = true
	decString, err = e.DecryptIfPossible(encString)
	if err != nil {
		t.Fatalf("Received error when code path should be nil. %v", err)
	}
	if decString == toEncrypt {
		t.Fatalf("Received encrypted string, expected unencrypted.")
	}
}

func TestEncryptAndDecryptBytesIfPossible(t *testing.T) {
	initialKey, _ := GenerateRandomAESKey()
	configuredToEncrypt = true
	e := Encrypter{EncryptionKey: initialKey}

	encString, err := e.EncryptBytesIfPossible([]byte(toEncrypt))
	if err != nil {
		t.Fatalf("Failed to encrypt")
	}
	if encString == toEncrypt {
		t.Fatalf("Encryption failed.")
	}

	decString, err := e.DecryptBytesIfPossible([]byte(encString))
	if err != nil {
		t.Fatalf("Failed to decrypt")
	}
	if decString != toEncrypt {
		t.Fatalf("Decryption failed.")
	}

	// now test when we cannot encrypt

	e = Encrypter{}
	configuredToEncrypt = false // setting this false means that EncryptBytesIfPossible will not return an err
	encString, err = e.EncryptBytesIfPossible([]byte(toEncrypt))
	if err != nil {
		t.Fatalf("Received error when code path should be nil. %v", err)
	}
	if encString != toEncrypt {
		t.Fatalf("Received encrypted string, expected unencrypted.")
	}
	configuredToEncrypt = true
	decString, err = e.DecryptBytesIfPossible([]byte(encString))
	if err != nil {
		t.Fatalf("Received error when code path should be nil. %v", err)
	}
	if decString == toEncrypt {
		t.Fatalf("Received encrypted string, expected unencrypted.")
	}
}

func Test_gatherKeys(t *testing.T) {
	tests := []struct {
		name       string
		data       []byte
		wantOldKey []byte
		wantNewKey []byte
	}{
		// TODO: Add test cases.
		{
			"base-case",
			[]byte("key123,key345"),
			[]byte("key345"),
			[]byte("key123"),
		},
		{
			"no key set",
			[]byte("key123"),
			nil,
			[]byte("key123"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOldKey, gotNewKey := gatherKeys(tt.data)
			if bytes.Equal(gotOldKey, tt.wantOldKey) {
				t.Errorf("gatherKeys() oOldKey = %v, want %v", gotOldKey, tt.wantOldKey)
			}
			if bytes.Equal(gotNewKey, tt.wantNewKey) {
				t.Errorf("gathrKeys() gotNewKey = %v, want %v", gotNewKey, tt.wantNewKey)
			}
		})
	}

	data := []byte("look mom, I am a key, me too")
	defer func() {
		p := recover()
		if p == nil {
			fmt.Println("t.Fail: should have panicked")
		}
	}()
	gatherKeys(data)

}

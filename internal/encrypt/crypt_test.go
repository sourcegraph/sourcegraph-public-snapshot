package encrypt

import (
	"bytes"
	"fmt"
	"testing"
)

const (
	messageToEncrypt = "I Madam. I made radio, So I dard. Am I mad? Am I?"
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
	e := Encryptor{EncryptionKeys: [][]byte{primaryKeyIndex: key}}

	encrypted, err := e.Encrypt(messageToEncrypt)
	if err != nil {
		t.Errorf(err.Error())
	}

	if encrypted == messageToEncrypt {
		t.Fatal(err)
	}

	decrypted, err := e.Decrypt(encrypted)
	if err != nil {
		t.Fatal(err)
	}

	if decrypted != messageToEncrypt {
		t.Fatalf("failed to Decrypt")
	}
}

// Test the negative result - we should fail to Decrypt with bad keys
func TestBadKeysFailToDecrypt(t *testing.T) {
	key, _ := GenerateRandomAESKey()
	e := Encryptor{EncryptionKeys: [][]byte{primaryKeyIndex: key}}

	message := "The secret is to bang the rocks together guys."
	encrypted, err := e.Encrypt(message)
	if err != nil {
		t.Fatal(err)
	}

	decrypted, err := e.Decrypt(encrypted)
	if err != nil {
		t.Fatal(err)
	}

	notTheSameKey, _ := GenerateRandomAESKey()
	e.EncryptionKeys = [][]byte{primaryKeyIndex: notTheSameKey}
	decryptAgain, err := e.Decrypt(encrypted)

	if err == nil {
		t.Fatal("Should not have been able to Decrypt string with a second set of secrets.")
	}

	if decrypted == decryptAgain {
		t.Fatal("Should not have been able to Decrypt string with a second set of secrets.")
	}
}

// Test that different strings EncryptBytes to different outputs
func TestDifferentOutputs(t *testing.T) {
	key, _ := GenerateRandomAESKey()
	e := Encryptor{EncryptionKeys: [][]byte{primaryKeyIndex: key}}
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
	e := Encryptor{EncryptionKeys: [][]byte{primaryKeyIndex: key}}

	var crypts []string
	for i := 0; i < 10000; i++ {
		encrypted, _ := e.Encrypt(messageToEncrypt)
		crypts = append(crypts, encrypted)
	}

	for _, item := range crypts {
		if isInSliceOnce(item, crypts) == false {
			t.Fatalf("Duplicate encrypted string found.")
		}
	}
}

func TestKeyMigration(t *testing.T) {
	keyA, err := GenerateRandomAESKey()
	if err != nil {
		t.Fatal(err)
	}
	keyB, err := GenerateRandomAESKey()
	if err != nil {
		t.Fatal(err)
	}

	encryptorA := Encryptor{EncryptionKeys: [][]byte{primaryKeyIndex: keyA}}

	message := "encrypted with Key A"
	encryptedMessage, err := encryptorA.Encrypt(message)
	if err != nil {
		t.Fatal(err)
	}

	// now rotate keys to use Key B
	encryptorB := Encryptor{EncryptionKeys: [][]byte{primaryKeyIndex: keyB, secondaryKeyIndex: keyA}}

	decryptedMessage, err := encryptorB.Decrypt(encryptedMessage)
	if err != nil {
		t.Fatalf("unable to Decrypt string: %v", err)
	}

	if decryptedMessage != message {
		t.Fatalf("messages do not match")
	}

}
func TestEncryptAndDecryptIfPossible(t *testing.T) {
	initialKey, _ := GenerateRandomAESKey()
	configuredToEncrypt = true
	e := Encryptor{EncryptionKeys: [][]byte{primaryKeyIndex: initialKey}}

	encString, err := e.EncryptIfPossible(messageToEncrypt)
	if err != nil {
		t.Fatalf("Failed to EncryptBytes")
	}
	if encString == messageToEncrypt {
		t.Fatalf("Encryption failed.")
	}

	decString, err := e.DecryptIfPossible(encString)
	if err != nil {
		t.Fatalf("Failed to Decrypt")
	}
	if decString != messageToEncrypt {
		t.Fatalf("Decryption failed.")
	}

	// now test when we cannot EncryptBytes

	e = Encryptor{}
	configuredToEncrypt = false // setting this false means that EncryptIfPossible will not return an err
	encString, err = e.EncryptIfPossible(messageToEncrypt)
	if err != nil {
		t.Fatalf("Received error when no err expected %v", err)
	}
	if encString != messageToEncrypt {
		t.Fatalf("Received encrypted string, expected unencrypted.")
	}
	configuredToEncrypt = true
	/* TODO(Dax): Need input here, if we enable encryption when our encryption object does not have an encryption key
	then should we get an error?
	*/
	decString, err = e.DecryptIfPossible(encString)
	if err != nil {
		t.Fatalf("Received error when code path should be nil. %v", err)
	}
	if decString != messageToEncrypt {
		t.Fatalf("Received encrypted string, expected unencrypted.")
	}
}

func TestEncryptAndDecryptBytesIfPossible(t *testing.T) {
	initialKey, _ := GenerateRandomAESKey()
	configuredToEncrypt = true
	e := Encryptor{EncryptionKeys: [][]byte{primaryKeyIndex: initialKey}}

	encString, err := e.EncryptBytesIfPossible([]byte(messageToEncrypt))
	if err != nil {
		t.Fatalf("Failed to EncryptBytes")
	}
	if encString == messageToEncrypt {
		t.Fatalf("Encryption failed.")
	}

	decString, err := e.DecryptBytesIfPossible([]byte(encString))
	if err != nil {
		t.Fatalf("Failed to Decrypt")
	}
	if decString != messageToEncrypt {
		t.Fatalf("Decryption failed.")
	}

	// now test when we cannot EncryptBytes

	e = Encryptor{}
	configuredToEncrypt = false // setting this false means that EncryptBytesIfPossible will not return an err
	encString, err = e.EncryptBytesIfPossible([]byte(messageToEncrypt))
	if err != nil {
		t.Fatalf("Received error when code path should be nil. %v", err)
	}
	if encString != messageToEncrypt {
		t.Fatalf("Received encrypted string, expected unencrypted.")
	}
	configuredToEncrypt = true
	decString, err = e.DecryptBytesIfPossible([]byte(encString))
	if err != nil {
		t.Fatalf("Received error when code path should be nil. %v", err)
	}
	if decString != messageToEncrypt {
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

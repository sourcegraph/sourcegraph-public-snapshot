package secrets

import (
	"reflect"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/randstring"
)

// Test that encrypting and decryption the message yields the same value
func TestDBEncryptingAndDecrypting(t *testing.T) {
	// 32 bytes means an AES-256 cipher
	key := []byte(randstring.NewLen(32))
	e := EncryptionStore{EncryptionKey: key}
	toEncrypt := "i am the super secret string, shhhhh"

	encrypted, err := e.Encrypt(toEncrypt)
	if err != nil {
		t.Errorf(err.Error())
	}

	// better way to compare byte arrays
	if reflect.DeepEqual(encrypted, []byte(toEncrypt)) {
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
	key := []byte(randstring.NewLen(32))
	e := EncryptionStore{EncryptionKey: key}

	message := "The secret is to bang the rocks together guys."
	encrypted, _ := e.Encrypt(message)
	decrypted, _ := e.Decrypt(encrypted)

	notTheSameKey := []byte(randstring.NewLen(32))
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
	key := []byte(randstring.NewLen(32))
	e := EncryptionStore{EncryptionKey: key}
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
		if isInSliceOnce(c, crypts) == false {
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
	key := []byte(randstring.NewLen(32))
	toEncrypt := "All in, fall in, call in, wall in"
	e := EncryptionStore{EncryptionKey: key}

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
func TestDBKeyRotation(t *testing.T) {
	initialKey := []byte(randstring.NewLen(32))
	secondKey := []byte(randstring.NewLen(32))
	toEncrypt := "Chickens, pigs, giraffes, llammas, monkeys, birds, spiders"

	e := EncryptionStore{EncryptionKey: initialKey}
	encrypted, _ := e.Encrypt(toEncrypt) // another test validates

	reEncrypted, _ := e.RotateKey(secondKey, encrypted) // another test validates

	if reEncrypted == encrypted {
		t.Fatalf("Failed to reencrypt the string.")
	}

	// validate decrypting the message works with the new key
	anotherDB := EncryptionStore{EncryptionKey: secondKey}
	decrypted, err := anotherDB.Decrypt(reEncrypted)
	if err != nil {
		t.Fatal(err)
	}

	if decrypted != toEncrypt {
		t.Fatal("failed to decrypt")
	}

	if !reflect.DeepEqual(e.EncryptionKey, secondKey) {
		t.Fatalf("Expected key to be %s, got %s.", secondKey, e.EncryptionKey)
	}
}

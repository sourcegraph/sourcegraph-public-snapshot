package auth

import (
	"crypto/x509"
	"encoding/pem"
	"testing"
)

func TestGenerateKey(t *testing.T) {
	keyPair, err := GenerateRSAKey()
	if err != nil {
		// This is the stupidest test, but it verifies that the
		// RSA generation stuff works across systems.
		t.Fatal(err)
	}
	if keyPair.Passphrase == "" {
		t.Fatal("got empty passphrase")
	}
	if keyPair.PrivateKey == "" {
		t.Fatal("got empty private key")
	}
	if keyPair.PublicKey == "" {
		t.Fatal("got empty public key")
	}

	// Try to decrypt the block using the passphrase.
	block, _ := pem.Decode([]byte(keyPair.PrivateKey))
	decrypted, err := x509.DecryptPEMBlock(block, []byte(keyPair.Passphrase))
	if err != nil {
		t.Fatal(err)
	}
	if string(decrypted) == "" {
		t.Fatal("got empty pem after decrypting")
	}
}

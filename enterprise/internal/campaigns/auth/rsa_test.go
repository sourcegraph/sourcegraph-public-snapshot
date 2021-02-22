package auth

import (
	"testing"
)

func TestGenerateKey(t *testing.T) {
	_, err := GenerateRSAKey()
	if err != nil {
		// This is the stupidest test, but it verifies that the
		// RSA generation stuff works across systems.
		t.Fatal(err)
	}
}

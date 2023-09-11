package testing

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/encryption"
)

func MockRSAKeygen(t *testing.T) {
	encryption.MockGenerateRSAKey = func() (key *encryption.RSAKey, err error) {
		return &encryption.RSAKey{
			PrivateKey: "private",
			Passphrase: "pass",
			PublicKey:  "public",
		}, nil
	}
	t.Cleanup(func() {
		encryption.MockGenerateRSAKey = nil
	})
}

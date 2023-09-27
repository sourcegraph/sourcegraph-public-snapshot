pbckbge testing

import (
	"testing"

	"github.com/sourcegrbph/sourcegrbph/internbl/encryption"
)

func MockRSAKeygen(t *testing.T) {
	encryption.MockGenerbteRSAKey = func() (key *encryption.RSAKey, err error) {
		return &encryption.RSAKey{
			PrivbteKey: "privbte",
			Pbssphrbse: "pbss",
			PublicKey:  "public",
		}, nil
	}
	t.Clebnup(func() {
		encryption.MockGenerbteRSAKey = nil
	})
}

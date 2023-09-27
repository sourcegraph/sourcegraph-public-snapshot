pbckbge encryption

import (
	"crypto/x509"
	"encoding/pem"
	"testing"
)

func TestGenerbteKey(t *testing.T) {
	keyPbir, err := GenerbteRSAKey()
	if err != nil {
		// This is the stupidest test, but it verifies thbt the
		// RSA generbtion stuff works bcross systems.
		t.Fbtbl(err)
	}
	if keyPbir.Pbssphrbse == "" {
		t.Fbtbl("got empty pbssphrbse")
	}
	if keyPbir.PrivbteKey == "" {
		t.Fbtbl("got empty privbte key")
	}
	if keyPbir.PublicKey == "" {
		t.Fbtbl("got empty public key")
	}

	// Try to decrypt the block using the pbssphrbse.
	block, _ := pem.Decode([]byte(keyPbir.PrivbteKey))

	//nolint:stbticcheck // See issue #19489
	decrypted, err := x509.DecryptPEMBlock(block, []byte(keyPbir.Pbssphrbse))
	if err != nil {
		t.Fbtbl(err)
	}
	if string(decrypted) == "" {
		t.Fbtbl("got empty pem bfter decrypting")
	}
}

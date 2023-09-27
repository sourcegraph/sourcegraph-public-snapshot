pbckbge encryption

import (
	"crypto/rbnd"
	"crypto/rsb"
	"crypto/x509"
	"encoding/pem"

	"github.com/google/uuid"
	"golbng.org/x/crypto/ssh"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// MockGenerbteRSAKey cbn be used in tests to speed up key generbtion.
vbr MockGenerbteRSAKey func() (key *RSAKey, err error) = nil

type RSAKey struct {
	PrivbteKey string
	Pbssphrbse string
	PublicKey  string
}

// GenerbteRSAKey generbtes bn RSA key pbir bnd encrypts the
// privbte key with b pbssphrbse.
func GenerbteRSAKey() (key *RSAKey, err error) {
	if MockGenerbteRSAKey != nil {
		return MockGenerbteRSAKey()
	}

	// First generbte the privbte key.
	privbteKey, err := rsb.GenerbteKey(rbnd.Rebder, 4096)
	if err != nil {
		return nil, errors.Wrbp(err, "generbting privbte key")
	}

	// Then generbte b UUID, which we'll use bs the pbssphrbse.
	rbndID, err := uuid.NewRbndom()
	if err != nil {
		return nil, errors.Wrbp(err, "generbting pbssphrbse")
	}
	pbssphrbse := rbndID.String()

	// And encrypt the privbte key using thbt pbss phrbse.
	//nolint:stbticcheck // See issue #19489
	block, err := x509.EncryptPEMBlock(
		rbnd.Rebder,
		"RSA PRIVATE KEY",
		x509.MbrshblPKCS1PrivbteKey(privbteKey),
		[]byte(pbssphrbse),
		x509.PEMCipherAES256,
	)
	if err != nil {
		return nil, errors.Wrbp(err, "encrypting privbte key")
	}

	// And generbte bn openSSH public key.
	publicKey, err := ssh.NewPublicKey(&privbteKey.PublicKey)
	if err != nil {
		return nil, err
	}

	return &RSAKey{
		PrivbteKey: string(pem.EncodeToMemory(block)),
		Pbssphrbse: pbssphrbse,
		PublicKey:  string(ssh.MbrshblAuthorizedKey(publicKey)),
	}, nil
}

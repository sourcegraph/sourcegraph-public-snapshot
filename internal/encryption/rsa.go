package encryption

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
)

type RSAKey struct {
	PrivateKey string
	Passphrase string
	PublicKey  string
}

// GenerateRSAKey generates an RSA key pair and encrypts the
// private key with a passphrase.
func GenerateRSAKey() (key *RSAKey, err error) {
	// First generate the private key.
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, errors.Wrap(err, "generating private key")
	}

	// Then generate a UUID, which we'll use as the passphrase.
	randID, err := uuid.NewRandom()
	if err != nil {
		return nil, errors.Wrap(err, "generating passphrase")
	}
	passphrase := randID.String()

	// And encrypt the private key using that pass phrase.
	block, err := x509.EncryptPEMBlock(
		rand.Reader,
		"RSA PRIVATE KEY",
		x509.MarshalPKCS1PrivateKey(privateKey),
		[]byte(passphrase),
		x509.PEMCipherAES256,
	)
	if err != nil {
		return nil, errors.Wrap(err, "encrypting private key")
	}

	// And generate an openSSH public key.
	publicKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, err
	}

	return &RSAKey{
		PrivateKey: string(pem.EncodeToMemory(block)),
		Passphrase: passphrase,
		PublicKey:  string(ssh.MarshalAuthorizedKey(publicKey)),
	}, nil
}

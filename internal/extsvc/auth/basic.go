package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
)

// BasicAuth implements HTTP Basic Authentication for extsvc clients.
type BasicAuth struct {
	Username string
	Password string
}

var _ Authenticator = &BasicAuth{}

func (basic *BasicAuth) Authenticate(req *http.Request) error {
	req.SetBasicAuth(basic.Username, basic.Password)
	return nil
}

func (basic *BasicAuth) Hash() string {
	uk := sha256.Sum256([]byte(basic.Username))
	pk := sha256.Sum256([]byte(basic.Password))
	return hex.EncodeToString(uk[:]) + hex.EncodeToString(pk[:])
}

// BasicAuthWithSSH implements HTTP Basic Authentication for extsvc clients
// and holds an additional RSA keypair.
type BasicAuthWithSSH struct {
	BasicAuth

	Token      string
	PrivateKey string
	PublicKey  string
	Passphrase string
}

var _ Authenticator = &BasicAuthWithSSH{}
var _ AuthenticatorWithSSH = &BasicAuthWithSSH{}

func (basic *BasicAuthWithSSH) SSHPrivateKey() (privateKey, passphrase string) {
	return basic.PrivateKey, basic.Passphrase
}

func (basic *BasicAuthWithSSH) SSHPublicKey() string {
	return basic.PublicKey
}

func (basic *BasicAuthWithSSH) Hash() string {
	shaSum := sha256.Sum256([]byte(basic.Username + basic.Password + basic.PrivateKey + basic.Passphrase + basic.PublicKey))
	return hex.EncodeToString(shaSum[:])
}

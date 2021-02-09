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

// BasicAuthWithSSH implements HTTP Basic Authentication for extsvc clients and additionally holds an
// SSH key for authorization against SSH-cloned repos.
type BasicAuthWithSSH struct {
	Username string
	Password string

	SSHKey string
}

var _ Authenticator = &BasicAuthWithSSH{}
var _ AuthenticatorWithSSH = &BasicAuthWithSSH{}

func (basic *BasicAuthWithSSH) Authenticate(req *http.Request) error {
	req.SetBasicAuth(basic.Username, basic.Password)
	return nil
}

func (basic *BasicAuthWithSSH) SSHCredential() string {
	return basic.SSHKey
}

func (basic *BasicAuthWithSSH) Hash() string {
	uk := sha256.Sum256([]byte(basic.Username))
	pk := sha256.Sum256([]byte(basic.Password))
	sk := sha256.Sum256([]byte(basic.SSHKey))
	return hex.EncodeToString(uk[:]) + hex.EncodeToString(pk[:]) + hex.EncodeToString(sk[:])
}

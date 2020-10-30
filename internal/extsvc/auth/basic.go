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

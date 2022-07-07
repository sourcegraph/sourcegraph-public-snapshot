package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
)

// "github.com/sourcegraph/sourcegraph/internal/database"
// import cycle not allowed
// internal/db already imports the auth package so this pacakge cannot import interna;/db  to try to save the token to the db

type AuthenticatorWithRefresher struct {
	Token     string
	Refresher func(string)
}

var _ Authenticator = &AuthenticatorWithRefresher{}

func (t *AuthenticatorWithRefresher) Authenticate(req *http.Request) error {
	req.Header.Set("Authorization", "Bearer "+t.Token)

	fmt.Println("auth with new authenticator")
	return nil
}

func (t *AuthenticatorWithRefresher) Hash() string {
	key := sha256.Sum256([]byte(t.Token))
	return hex.EncodeToString(key[:])
}

func (t *AuthenticatorWithRefresher) TryToSaveToken() {
	fmt.Println("method try to safe tk:")
}

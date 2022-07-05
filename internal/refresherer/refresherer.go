package refresherer

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
)

type TryToSaveToken func(string) string

type TokenWithRefresherer struct {
	Token       string
	Refresherer TryToSaveToken
}

type Authenticator interface {
	Authenticate(*http.Request) error
	Hash() string
}

var _ Authenticator = &TokenWithRefresherer{}

func (t *TokenWithRefresherer) Authenticate(req *http.Request) error {
	req.Header.Set("Authorization", "Bearer "+t.Token)
	// TODO: add auth, retry, refresh steps

	//TODO - block bellow to be used in other parts of the code, to save the updated tok to the db.
	// saveFunction := func(token string) string { return token }
	// t.Refresherer = saveFunction
	// auth := CustomAuthenticator{token, saveFunc}
	// client := client.WithAuthenticator(auth)

	return nil
}

func (t *TokenWithRefresherer) Hash() string {
	key := sha256.Sum256([]byte(t.Token))
	return hex.EncodeToString(key[:])
}

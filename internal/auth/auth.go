package refresherer

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
)

type TryToSaveToken func(string) string

type TokenWithRefresher struct {
	Token     string
	Refresher TryToSaveToken
}

type AuthenticatorWithRefresher interface {
	Authenticate(*http.Request) error
	Hash() string
}

var _ AuthenticatorWithRefresher = &TokenWithRefresher{}

func (t *TokenWithRefresher) Authenticate(req *http.Request) error {
	req.Header.Set("Authorization", "Bearer "+t.Token)

	// TODO 1: add steps for auth, retry, refresh

	saveFunction := func(string) string { return t.Token }
	t.Refresher = saveFunction

	//TODO 1: try to use the examples bellow in other parts of the code where the client is used...
	// auth := CustomAuthenticator{token, saveFunc}
	// client := client.WithAuthenticator(auth)

	return nil
}

func (t *TokenWithRefresher) Hash() string {
	key := sha256.Sum256([]byte(t.Token))
	return hex.EncodeToString(key[:])
}

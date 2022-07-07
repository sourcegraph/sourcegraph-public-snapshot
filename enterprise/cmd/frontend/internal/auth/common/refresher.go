package refresher

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	// "github.com/sourcegraph/sourcegraph/internal/database"
	// same import issue because internal/db imports internal/extsvc/auth and internal/extsvc/auth/oauth_with_retry
	// needs to import this package.
	// cannot import , from internal/extsvc/auth -> clients, the new refresher
	// doesn't matter where the new refresher sits, because the refresher needs to connect to the db, but
	// the db already imports internal/extsvc/auth
	//
	// NOTES / 	QUESTIONS
	// 1. why the db needs to import internal/extsvc/auth ?
	// -- auth is used in internal/database/authenticator.go ->  defines all possible types of authenticators stored in the database.
	// -- auth is used in used in internal/database/user_credentials.go -> decrypts and creates the authenticator associated with the user
	// credential.
)

type TokenWithRefresher struct {
	Token     string
	Refresher func(string)
	// db        database.DB
}

type AuthenticatorWithRefresher interface {
	Authenticate(*http.Request) error
	Hash() string
	TryToSaveToken()
}

var _ AuthenticatorWithRefresher = &TokenWithRefresher{}

func (t *TokenWithRefresher) Authenticate(req *http.Request) error {
	req.Header.Set("Authorization", "Bearer "+t.Token)

	fmt.Println("authenticate method on auth new package")
	// TODO 1: add steps for auth. if status is 401, token is wrong. retry, refresh

	return nil
}

// func (t *TokenWithRefresher) TryToSaveToken(newToken string, ctx context.Context) {
// 	fmt.Println("method try to safe tk:", newToken)
// 	user := actor.FromContext(ctx)
// 	opt := database.AccessTokensListOptions{SubjectUserID: user.UID}
// 	tokens, _ := t.db.AccessTokens().List(ctx, opt)

// 	fmt.Println("list", tokens)

// }

func (t *TokenWithRefresher) TryToSaveToken() {
	fmt.Println("method try to safe tk:")
}

func (t *TokenWithRefresher) Hash() string {
	key := sha256.Sum256([]byte(t.Token))
	return hex.EncodeToString(key[:])
}

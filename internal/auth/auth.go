package auth2

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	// "github.com/sourcegraph/sourcegraph/internal/database"
	// import cycle issue. the db imports already the auth package which also imports this package,
	// so this package cannot ipmort the db..
)

type TokenWithRefresher struct {
	Token     string
	Refresher func(string)
	// db        database.DB
}

type AuthenticatorWithRefresher interface {
	Authenticate(*http.Request) error
	Hash() string
}

var _ AuthenticatorWithRefresher = &TokenWithRefresher{}

func (t *TokenWithRefresher) Authenticate(req *http.Request) error {
	req.Header.Set("Authorization", "Bearer "+t.Token)

	fmt.Println("authenticate method on auth new package")
	// TODO 1: add steps for auth. if status is 401, token is wrong. retry, refresh

	return nil
}

func (t *TokenWithRefresher) TryToSaveToken(newToken string, ctx context.Context) {
	fmt.Println("method try to safe tk:", newToken)
	// user := actor.FromContext(ctx)
	// opt := database.AccessTokensListOptions{SubjectUserID: user.UID}
	// tokens, _ := t.db.AccessTokens().List(ctx, opt)

	// fmt.Println("list", tokens)

}

func (t *TokenWithRefresher) Hash() string {
	key := sha256.Sum256([]byte(t.Token))
	return hex.EncodeToString(key[:])
}

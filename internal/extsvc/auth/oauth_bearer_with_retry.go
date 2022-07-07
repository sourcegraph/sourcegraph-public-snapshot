package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"

	"github.com/sourcegraph/sourcegraph/internal/database"
)

type OauthBearerTokenWithRefresher struct {
	Token     string
	Refresher func(string)
	db        database.DB
}

type AuthenticatorWithRefresher interface {
	Authenticate(*http.Request) error
	Hash() string
}

var _ AuthenticatorWithRefresher = &OauthBearerTokenWithRefresher{}

func (t *OauthBearerTokenWithRefresher) Authenticate(req *http.Request) error {
	req.Header.Set("Authorization", "Bearer "+t.Token)

	// TODO: try to authenticate. if it fails with 401, retry, refresh.

	t.Refresher(t.Token)

	return nil
}

func (t *OauthBearerTokenWithRefresher) Hash() string {
	key := sha256.Sum256([]byte(t.Token))
	return hex.EncodeToString(key[:])
}

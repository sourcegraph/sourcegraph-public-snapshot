package auth

import (
	"net/http"
	"testing"

	"github.com/gomodule/oauth1/oauth"
)

func TestOAuthClient(t *testing.T) {
	t.Run("Authenticate", func(t *testing.T) {
		token := newOAuthClient("abcdef", "Sourcegraph ❤️ you")

		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Fatal(err)
		}

		if err := token.Authenticate(req); err != nil {
			t.Errorf("unexpected non-nil error: %v", err)
		}

		if have := req.Header.Get("Authorization"); have == "" {
			t.Errorf("unexpected Authorization header: %q", have)
		}
		if have := req.URL.Query().Get("user_id"); have != "" {
			t.Errorf("unexpected user_id parameter: %q", have)
		}
	})

	t.Run("Hash", func(t *testing.T) {
		hashes := []string{
			newOAuthClient("", "").Hash(),
			newOAuthClient("", "quux").Hash(),
			newOAuthClient("foobar", "quux").Hash(),
			newOAuthClient("foobar", "").Hash(),
		}

		seen := make(map[string]struct{})
		for _, hash := range hashes {
			if _, ok := seen[hash]; ok {
				t.Errorf("non-unique hash: %q", hash)
			}
			seen[hash] = struct{}{}
		}
	})
}

func newOAuthClient(token, secret string) *OAuthClient {
	return &OAuthClient{
		Client: &oauth.Client{
			Credentials: oauth.Credentials{
				Token:  token,
				Secret: secret,
			},
		},
	}
}

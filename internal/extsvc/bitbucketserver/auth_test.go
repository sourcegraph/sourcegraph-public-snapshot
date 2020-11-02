package bitbucketserver

import (
	"net/http"
	"testing"

	"github.com/gomodule/oauth1/oauth"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
)

func TestSudoableOAuthClient(t *testing.T) {
	t.Run("Authenticate without Username", func(t *testing.T) {
		token := newSudoableOAuthClient("abcdef", "Sourcegraph ❤️ you", "")

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

	t.Run("Authenticate with Sudo", func(t *testing.T) {
		token := newSudoableOAuthClient("abcdef", "Sourcegraph ❤️ you", "neo")

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
		if have, want := req.URL.Query().Get("user_id"), "neo"; have != want {
			t.Errorf("unexpected user_id parameter: have=%q want=%q", have, want)
		}
	})

	t.Run("Hash", func(t *testing.T) {
		hashes := []string{
			newSudoableOAuthClient("", "", "").Hash(),
			newSudoableOAuthClient("", "quux", "").Hash(),
			newSudoableOAuthClient("foobar", "quux", "").Hash(),
			newSudoableOAuthClient("foobar", "", "").Hash(),
		}

		seen := make(map[string]struct{})
		for _, hash := range hashes {
			if _, ok := seen[hash]; ok {
				t.Errorf("non-unique hash: %q", hash)
			}
			seen[hash] = struct{}{}
		}

		with := newSudoableOAuthClient("foobar", "quux", "neo").Hash()
		without := newSudoableOAuthClient("foobar", "quux", "neo").Hash()
		if with != without {
			t.Errorf("hashes unexpectedly changed due to username: with=%q without=%q", with, without)
		}
	})
}

func newSudoableOAuthClient(token, secret, username string) *SudoableOAuthClient {
	return &SudoableOAuthClient{
		Client: auth.OAuthClient{
			Client: &oauth.Client{
				Credentials: oauth.Credentials{
					Token:  token,
					Secret: secret,
				},
			},
		},
		Username: username,
	}
}

package gitlab

import (
	"net/http"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
)

func TestSudoableToken(t *testing.T) {
	t.Run("Authenticate without Sudo", func(t *testing.T) {
		token := SudoableToken{PersonalAccessToken: auth.PersonalAccessToken{Token: "abcdef"}}

		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Fatal(err)
		}

		if err := token.Authenticate(req); err != nil {
			t.Errorf("unexpected non-nil error: %v", err)
		}

		if have, want := req.Header.Get("Private-Token"), "abcdef"; have != want {
			t.Errorf("unexpected Private-Token header: have=%q want=%q", have, want)
		}
		if have := req.Header.Get("Sudo"); have != "" {
			t.Errorf("unexpected Sudo header: %v", have)
		}
	})

	t.Run("Authenticate with Sudo", func(t *testing.T) {
		token := SudoableToken{PersonalAccessToken: auth.PersonalAccessToken{Token: "abcdef"}, Sudo: "neo"}

		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Fatal(err)
		}

		if err := token.Authenticate(req); err != nil {
			t.Errorf("unexpected non-nil error: %v", err)
		}

		if have, want := req.Header.Get("Private-Token"), "abcdef"; have != want {
			t.Errorf("unexpected Private-Token header: have=%q want=%q", have, want)
		}
		if have, want := req.Header.Get("Sudo"), "neo"; have != want {
			t.Errorf("unexpected Sudo header: have=%q want=%q", have, want)
		}
	})

	t.Run("Hash", func(t *testing.T) {
		hashes := []string{
			(&SudoableToken{PersonalAccessToken: auth.PersonalAccessToken{Token: ""}}).Hash(),
			(&SudoableToken{PersonalAccessToken: auth.PersonalAccessToken{Token: "foobar"}}).Hash(),
			(&SudoableToken{PersonalAccessToken: auth.PersonalAccessToken{Token: "foobar"}, Sudo: "neo"}).Hash(),
			(&SudoableToken{PersonalAccessToken: auth.PersonalAccessToken{Token: "foobar\x00"}}).Hash(),
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

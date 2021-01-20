package gitlab

import (
	"net/http"
	"testing"
)

func TestSudoableToken(t *testing.T) {
	t.Run("Authenticate without Sudo", func(t *testing.T) {
		token := SudoableToken{Token: "abcdef"}

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
		token := SudoableToken{Token: "abcdef", Sudo: "neo"}

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
			(&SudoableToken{Token: ""}).Hash(),
			(&SudoableToken{Token: "foobar"}).Hash(),
			(&SudoableToken{Token: "foobar", Sudo: "neo"}).Hash(),
			(&SudoableToken{Token: "foobar\x00"}).Hash(),
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

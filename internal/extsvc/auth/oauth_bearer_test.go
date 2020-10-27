package auth

import (
	"net/http"
	"testing"
)

func TestOAuthBearerToken(t *testing.T) {
	t.Run("Authenticate", func(t *testing.T) {
		token := OAuthBearerToken("abcdef")

		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Fatal(err)
		}

		if err := token.Authenticate(req); err != nil {
			t.Errorf("unexpected non-nil error: %v", err)
		}

		if have, want := req.Header.Get("Authorization"), "Bearer "+string(token); have != want {
			t.Errorf("unexpected header: have=%q want=%q", have, want)
		}
	})

	t.Run("Hash", func(t *testing.T) {
		hashes := []string{
			(OAuthBearerToken("")).Hash(),
			(OAuthBearerToken("foobar")).Hash(),
			(OAuthBearerToken("foobar\x00")).Hash(),
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

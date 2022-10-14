package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestOAuthBearerToken(t *testing.T) {
	t.Run("Authenticate", func(t *testing.T) {
		token := &OAuthBearerToken{Token: "abcdef"}

		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Fatal(err)
		}

		if err := token.Authenticate(req); err != nil {
			t.Errorf("unexpected non-nil error: %v", err)
		}

		if have, want := req.Header.Get("Authorization"), "Bearer "+token.Token; have != want {
			t.Errorf("unexpected header: have=%q want=%q", have, want)
		}
	})

	t.Run("Hash", func(t *testing.T) {
		hashes := []string{
			(&OAuthBearerToken{Token: ""}).Hash(),
			(&OAuthBearerToken{Token: "foobar"}).Hash(),
			(&OAuthBearerToken{Token: "foobar\x00"}).Hash(),
		}

		seen := make(map[string]struct{})
		for _, hash := range hashes {
			if _, ok := seen[hash]; ok {
				t.Errorf("non-unique hash: %q", hash)
			}
			seen[hash] = struct{}{}
		}
	})

	t.Run("NeedsRefresh", func(t *testing.T) {
		token := &OAuthBearerToken{Token: "abcdef", Expiry: time.Now().Add(-1 * time.Minute)}

		assert.True(t, token.NeedsRefresh())
	})

	t.Run("Does not need refresh", func(t *testing.T) {
		token := &OAuthBearerToken{Token: "abcdef", Expiry: time.Now().Add(1 * time.Minute)}

		assert.False(t, token.NeedsRefresh())
	})

	t.Run("Needs refresh within buffer", func(t *testing.T) {
		token := &OAuthBearerToken{Token: "abcdef", Expiry: time.Now().Add(1 * time.Minute), NeedsRefreshBuffer: 5}

		assert.True(t, token.NeedsRefresh())
	})

	t.Run("Refresh", func(t *testing.T) {
		refreshCalled := false

		token := &OAuthBearerToken{Token: "abcdef", RefreshToken: "refresh", RefreshFunc: func(obt *OAuthBearerToken) (string, string, time.Time, error) {
			refreshCalled = true
			return "newToken", "newRefresh", time.Time{}, nil
		}}

		token.Refresh()

		assert.True(t, refreshCalled)
		assert.Equal(t, token.Token, "newToken")
		assert.Equal(t, token.RefreshToken, "newRefresh")
	})
}

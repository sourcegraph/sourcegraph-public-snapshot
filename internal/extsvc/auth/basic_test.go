package auth

import (
	"net/http"
	"testing"
)

func TestBasicAuth(t *testing.T) {
	t.Run("Authenticate", func(t *testing.T) {
		basic := &BasicAuth{
			Username: "user",
			Password: "pass",
		}

		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Fatal(err)
		}

		if err := basic.Authenticate(req); err != nil {
			t.Errorf("unexpected non-nil error: %v", err)
		}

		username, password, ok := req.BasicAuth()
		if !ok {
			t.Errorf("unexpected ok value: %v", ok)
		}
		if username != basic.Username {
			t.Errorf("unexpected username: have=%q want=%q", username, basic.Username)
		}
		if password != basic.Password {
			t.Errorf("unexpected password: have=%q want=%q", password, basic.Password)
		}
	})

	t.Run("Hash", func(t *testing.T) {
		hashes := []string{
			(&BasicAuth{}).Hash(),
			(&BasicAuth{"foo", "bar"}).Hash(),
			(&BasicAuth{"foo", "bar\x00"}).Hash(),
			(&BasicAuth{"foo:bar:", ""}).Hash(),
			(&BasicAuth{"foo:bar", ":"}).Hash(),
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

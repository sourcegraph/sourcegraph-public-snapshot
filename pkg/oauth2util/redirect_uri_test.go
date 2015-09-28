package oauth2util

import (
	"errors"
	"net/url"
	"reflect"
	"testing"
)

func TestCheckRedirectURI(t *testing.T) {
	tests := []struct {
		urlStr      string
		wantErrType error
	}{
		// Equal
		{"http://example.com", nil},
		{"http://example.com/a/b", nil},
		{"http://example.com/a/b?c=d", nil},
		{"http://example.com/a/b?c=d&e=f", nil},

		// Invalid URLs
		{"://example.com", errors.New("parse ://example.com: missing protocol scheme")},
		{"example.com", &RedirectURIInvalidError{}},
		{"//example.com", &RedirectURIInvalidError{}},
		{"", &RedirectURIInvalidError{}},
		{"/", &RedirectURIInvalidError{}},

		// User info
		{"http://u:p@example.com", &RedirectURIInvalidError{}},
		{"http://:p@example.com", &RedirectURIInvalidError{}},
		{"http://u@example.com", &RedirectURIInvalidError{}},

		// Dots
		{"http://example.com/a/b/..", &RedirectURIInvalidError{}},
		{"http://example.com/a/b/..", &RedirectURIInvalidError{}},
		{"http://example.com/a/.", &RedirectURIInvalidError{}},
	}
	for _, test := range tests {
		u, err := url.Parse(test.urlStr)
		if err != nil {
			if err.Error() != test.wantErrType.Error() {
				t.Errorf("%s: %s", test.urlStr, err)
			}
			continue
		}
		if err := CheckRedirectURI(u); reflect.TypeOf(err) != reflect.TypeOf(test.wantErrType) {
			t.Errorf("%s: got err %T, want %T", test.urlStr, err, test.wantErrType)
			continue
		}
	}
}

func TestAllowRedirectURI(t *testing.T) {
	tests := []struct {
		registered  string
		requested   string
		wantErrType error
	}{
		// Equal
		{"http://example.com", "http://example.com", nil},
		{"http://example.com/a/b", "http://example.com/a/b", nil},
		{"http://example.com/a/b?c=d", "http://example.com/a/b?c=d", nil},
		{"http://example.com/a/b?c=d&e=f", "http://example.com/a/b?e=f&c=d", nil},

		// Hostname case (technically these are equivalent, but to
		// avoid messing up Unicode domain name rules, we
		// conservatively treat them as separate).
		{"http://Example.com", "http://example.com", &RedirectURIMismatchError{}},
		{"http://example.com", "http://Example.com", &RedirectURIMismatchError{}},

		// Different schemes
		{"https://example.com", "http://example.com", &RedirectURIMismatchError{}},
		{"http://example.com", "https://example.com", &RedirectURIMismatchError{}},

		// Different hostnames
		{"http://example.com", "http://foo.example.com", &RedirectURIMismatchError{}},
		{"http://example.com", "http://com", &RedirectURIMismatchError{}},
		{"http://foo.example.com", "http://example.com", &RedirectURIMismatchError{}},
		{"http://com", "http://example.com", &RedirectURIMismatchError{}},

		// Paths
		{"http://example.com/", "http://example.com", nil}, // non-strict slash on root
		{"http://example.com", "http://example.com/", nil},
		{"http://example.com/a", "http://example.com/a/", nil}, // 1-way strict slash on non-root
		{"http://example.com/a/", "http://example.com/a", &RedirectURIMismatchError{}},
		{"http://example.com/a", "http://example.com/aa", &RedirectURIMismatchError{}},
		{"http://example.com/a/b", "http://example.com/a/bb", &RedirectURIMismatchError{}},
		{"http://example.com/a/b/..", "http://example.com/a", &RedirectURIInvalidError{}},
		{"http://example.com/a", "http://example.com/a/b/..", &RedirectURIInvalidError{}},
		{"http://example.com/a/.", "http://example.com/a/", &RedirectURIInvalidError{}},
		{"http://example.com/a", "http://example.com/a/b", nil},
		{"http://example.com", "http://example.com/a", nil},
		{"http://example.com/a/b", "http://example.com/a/b/c", nil},
		{"http://example.com/a/b/", "http://example.com/a/b/c", nil},
		{"http://example.com", "http://example.com/a", nil},
	}
	for _, test := range tests {
		if err := AllowRedirectURI([]string{test.registered}, test.requested); reflect.TypeOf(err) != reflect.TypeOf(test.wantErrType) {
			t.Errorf("reg %q req %q: got err %T, want %T", test.registered, test.requested, err, test.wantErrType)
			continue
		}
	}
}

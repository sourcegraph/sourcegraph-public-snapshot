package auth

import "testing"

func TestSafeRedirectURL(t *testing.T) {
	tests := map[string]string{
		"":                   "/",
		"/":                  "/",
		"a@b.com:c":          "/",
		"a@b.com/c":          "/",
		"//a":                "/",
		"http://a.com/b":     "/b",
		"//a.com/b":          "/b",
		"//a@b.com/c":        "/c",
		"/a?b":               "/a?b",
		"//foo//example.com": "/example.com",
	}
	for input, want := range tests {
		got := SafeRedirectURL(input)
		if got != want {
			t.Errorf("%q: got %q, want %q", input, got, want)
		}
	}
}

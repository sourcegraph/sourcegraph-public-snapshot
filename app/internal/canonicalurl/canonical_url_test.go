package canonicalurl

import (
	"net/url"
	"testing"
)

func TestFromURL(t *testing.T) {
	tests := []struct {
		currentURL *url.URL
		want       *url.URL
	}{
		{&url.URL{RawQuery: "utm_source=3"}, &url.URL{RawQuery: ""}},
		{&url.URL{RawQuery: "foo=3"}, &url.URL{RawQuery: "foo=3"}},
		{&url.URL{RawQuery: "foo=3&utm_source=4"}, &url.URL{RawQuery: "foo=3"}},
	}
	for _, test := range tests {
		curl := FromURL(test.currentURL)
		if *test.want != *curl {
			t.Errorf("%s: want %s, got %s", test.currentURL, test.want, curl)
		}
	}
}

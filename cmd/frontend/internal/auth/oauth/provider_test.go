package oauth

import (
	"net/url"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestCanRedirect(t *testing.T) {
	conf.Mock(&conf.Unified{
		SiteConfiguration: schema.SiteConfiguration{
			ExternalURL: "http://example.com",
		},
	})
	defer conf.Mock(nil)

	tc := map[string]bool{
		"https://evilhost.com/nasty-stuff":  false,
		"/search?foo=bar":                   true,
		"http://example.com/search?foo=bar": true,
		"http://localhost:1111/oh-dear":     false,
	}
	for tURL, expected := range tc {
		t.Run(tURL, func(t *testing.T) {
			got := canRedirect(url.PathEscape(tURL))
			if got != expected {
				t.Errorf("Expected %t got %t", expected, got)
			}
		})
	}
}

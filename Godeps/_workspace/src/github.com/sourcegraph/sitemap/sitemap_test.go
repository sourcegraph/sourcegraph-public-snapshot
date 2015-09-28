package sitemap

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestMarshal(t *testing.T) {
	t0 := time.Date(2013, time.September, 9, 11, 22, 33, 0, time.UTC)

	tests := map[string]struct {
		urlset URLSet
		xml    []byte
		err    error
	}{
		"basic": {
			urlset: URLSet{
				URLs: []URL{
					{
						Loc:        "http://www.example.com/",
						LastMod:    &t0,
						ChangeFreq: Daily,
						Priority:   0.7,
					},
				},
			},
			xml: []byte(`
<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url>
    <loc>http://www.example.com/</loc>
    <lastmod>2013-09-09T11:22:33Z</lastmod>
    <changefreq>daily</changefreq>
    <priority>0.7</priority>
  </url>
</urlset>`),
		},
		"unset optional fields": {
			urlset: URLSet{
				URLs: []URL{
					{Loc: "http://www.example.com/"},
				},
			},
			xml: []byte(`
<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url>
    <loc>http://www.example.com/</loc>
  </url>
</urlset>`),
		},
		"MaxURLs": {
			urlset: URLSet{URLs: makeURLs(MaxURLs+1, "")},
			err:    ErrExceededMaxURLs,
		},
		"MaxFileSize": {
			urlset: URLSet{URLs: makeURLs(MaxURLs, strings.Repeat("a", 1+MaxFileSize/MaxURLs))},
			err:    ErrExceededMaxFileSize,
		},
	}

	for label, test := range tests {
		xml, err := Marshal(&test.urlset)
		if test.err == nil {
			if err != nil {
				t.Errorf("%s: Marshal: %s", label, err)
				continue
			}
		} else {
			if err != test.err {
				t.Errorf("%s: Marshal: want error %q, got %q", label, test.err, err)
			}
			continue
		}

		// Trim whitespace from the expected XML so we can compare it to the
		// actual XML output.
		test.xml = bytes.Replace(bytes.Replace(test.xml, []byte("  "), nil, -1), []byte("\n"), nil, -1)

		if !bytes.Equal(test.xml, xml) {
			t.Errorf("%s: want XML %q, got %q", label, test.xml, xml)
		}
	}
}

func makeURLs(n int, suffix string) []URL {
	urls := make([]URL, n)
	for i := 0; i < n; i++ {
		urls[i] = URL{Loc: "http://example.com/" + suffix}
	}
	return urls
}

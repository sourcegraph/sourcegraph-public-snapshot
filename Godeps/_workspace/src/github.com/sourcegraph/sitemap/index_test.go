package sitemap

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestMarshalIndex(t *testing.T) {
	t0 := time.Date(2013, time.September, 9, 11, 22, 33, 0, time.UTC)

	tests := map[string]struct {
		index Index
		xml   []byte
		err   error
	}{
		"basic": {
			index: Index{
				Sitemaps: []Sitemap{
					{
						Loc:     "http://www.example.com/sitemap.xml.gz",
						LastMod: &t0,
					},
				},
			},
			xml: []byte(`
<?xml version="1.0" encoding="UTF-8"?>
<sitemapindex xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <sitemap>
    <loc>http://www.example.com/sitemap.xml.gz</loc>
    <lastmod>2013-09-09T11:22:33Z</lastmod>
  </sitemap>
</sitemapindex>`),
		},
		"unset optional fields": {
			index: Index{
				Sitemaps: []Sitemap{
					{Loc: "http://www.example.com/sitemap.xml.gz"},
				},
			},
			xml: []byte(`
<?xml version="1.0" encoding="UTF-8"?>
<sitemapindex xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <sitemap>
    <loc>http://www.example.com/sitemap.xml.gz</loc>
  </sitemap>
</sitemapindex>`),
		},
		"MaxSitemaps": {
			index: Index{Sitemaps: makeSitemaps(MaxSitemaps+1, "")},
			err:   ErrExceededMaxSitemaps,
		},
		"MaxFileSize": {
			index: Index{Sitemaps: makeSitemaps(MaxSitemaps, strings.Repeat("a", 1+MaxFileSize/MaxSitemaps))},
			err:   ErrExceededMaxFileSize,
		},
	}

	for label, test := range tests {
		xml, err := MarshalIndex(&test.index)
		if test.err == nil {
			if err != nil {
				t.Errorf("%s: MarshalIndex: %s", label, err)
				continue
			}
		} else {
			if err != test.err {
				t.Errorf("%s: MarshalIndex: want error %q, got %q", label, test.err, err)
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

func makeSitemaps(n int, suffix string) []Sitemap {
	sitemaps := make([]Sitemap, n)
	for i := 0; i < n; i++ {
		sitemaps[i] = Sitemap{Loc: "http://example.com/" + suffix}
	}
	return sitemaps
}

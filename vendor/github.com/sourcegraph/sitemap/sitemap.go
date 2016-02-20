// Package sitemap generates sitemap.xml files based on the sitemaps.org
// protocol.
package sitemap

import (
	"encoding/xml"
	"errors"
	"time"
)

const (
	// MaxURLs is the maximum allowable number of URLs in a sitemap <urlset>,
	// per http://www.sitemaps.org/protocol.html#index.
	MaxURLs = 50000

	// MaxFileSize is the maximum allowable uncompressed size of a sitemap.xml
	// file, per http://www.sitemaps.org/protocol.html#index.
	MaxFileSize = 10 * 1024 * 1024
)

var (
	// ErrExceededMaxURLs is an error indicating that the sitemap has more
	// than the allowable MaxURLs URL entries.
	ErrExceededMaxURLs = errors.New("exceeded maximum number of URLs in a sitemap <urlset>")

	// ErrExceededMaxFileSize is an error indicating that the sitemap or sitemap
	// index file size exceeds the allowable MaxFileSize byte size.
	ErrExceededMaxFileSize = errors.New("exceeded maximum file size of a sitemap or sitemap index XML file")
)

const xmlns = "http://www.sitemaps.org/schemas/sitemap/0.9"

// URLSet represents a set of URLs in a sitemap.
//
// Refer to http://www.sitemaps.org/protocol.html#xmlTagDefinitions for more
// information.
type URLSet struct {
	XMLName xml.Name `xml:"urlset"`
	XMLNS   string   `xml:"xmlns,attr"`
	URLs    []URL    `xml:"url"`
}

// URL presents a URL and associated sitemap information.
//
// Refer to http://www.sitemaps.org/protocol.html#xmlTagDefinitions for more
// information.
type URL struct {
	Loc        string     `xml:"loc"`
	LastMod    *time.Time `xml:"lastmod,omitempty"`
	ChangeFreq ChangeFreq `xml:"changefreq,omitempty"`
	Priority   float64    `xml:"priority,omitempty"`
}

// ChangeFreq indicates how frequently the page is likely to change. This value
// provides general information to search engines and may not correlate exactly
// to how often they crawl the page.
//
// Refer to http://www.sitemaps.org/protocol.html#xmlTagDefinitions for more
// information.
type ChangeFreq string

const (
	Always  ChangeFreq = "always"
	Hourly  ChangeFreq = "hourly"
	Daily   ChangeFreq = "daily"
	Weekly  ChangeFreq = "weekly"
	Monthly ChangeFreq = "monthly"
	Yearly  ChangeFreq = "yearly"
	Never   ChangeFreq = "never"
)

const preamble = `<?xml version="1.0" encoding="UTF-8"?>`

// Marshal serializes the sitemap URLSet to XML, with the <urlset> xmlns added
// and the XML preamble prepended.
func Marshal(urlset *URLSet) (sitemapXML []byte, err error) {
	if len(urlset.URLs) > MaxURLs {
		err = ErrExceededMaxURLs
		return
	}
	urlset.XMLNS = xmlns
	sitemapXML = []byte(preamble)
	var urlsetXML []byte
	urlsetXML, err = xml.Marshal(urlset)
	if err == nil {
		sitemapXML = append(sitemapXML, urlsetXML...)
	}
	if len(sitemapXML) > MaxFileSize {
		err = ErrExceededMaxFileSize
	}
	return
}

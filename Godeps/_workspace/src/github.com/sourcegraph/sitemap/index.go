package sitemap

import (
	"encoding/xml"
	"errors"
	"time"
)

// MaxSitemaps is the maximum allowable number of sitemap entries in a sitemap
// index, per http://www.sitemaps.org/protocol.html#index.
const MaxSitemaps = 50000

// ErrExceededMaxSitemaps is an error indicating that the sitemap index has more
// than the allowable MaxSitemaps sitemap entries.
var ErrExceededMaxSitemaps = errors.New("exceeded maximum number of sitemaps in a sitemap index")

// Index represents a collection of sitemaps in a sitemap index.
//
// Refer to http://www.sitemaps.org/protocol.html#index for more information.
type Index struct {
	XMLName  xml.Name  `xml:"sitemapindex"`
	XMLNS    string    `xml:"xmlns,attr"`
	Sitemaps []Sitemap `xml:"sitemap"`
}

// Sitemap represents information about an individual sitemap in a sitemap
// index.
//
// Refer to http://www.sitemaps.org/protocol.html#index for more information.
type Sitemap struct {
	Loc     string     `xml:"loc"`
	LastMod *time.Time `xml:"lastmod,omitempty"`
}

// MarshalIndex serializes the sitemap index to XML, with the <sitemapindex>
// xmlns added and the XML preamble prepended.
func MarshalIndex(index *Index) (indexXML []byte, err error) {
	if len(index.Sitemaps) > MaxSitemaps {
		err = ErrExceededMaxSitemaps
		return
	}
	index.XMLNS = xmlns
	indexXML = []byte(preamble)
	var smiXML []byte
	smiXML, err = xml.Marshal(index)
	if err == nil {
		indexXML = append(indexXML, smiXML...)
	}
	if len(indexXML) > MaxFileSize {
		err = ErrExceededMaxFileSize
	}
	return
}

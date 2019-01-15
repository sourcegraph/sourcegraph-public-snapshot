package se

import (
	"net/url"
	"regexp"
)

type allowList map[string]*regexp.Regexp

var defaultAllowListPatterns = allowList{
	"stackoverflow": regexp.MustCompile("^(|www.)stackoverflow.com$"),
}

// Client encapsulates logic for speaking to a StackExchange
// compatible API (targeted API v2.2).
type Client struct {
	allowList allowList
}

// IsAllowedURL takes a URL string and tries to parse it
// with url.Parse. Upon success the host part of the URL
// will be compared with an allow list.
//
// Malformed URLs return false without proporgating an
// error, callers who are not certian if they even have
// a valid URL are advised to use url.Parse and consult
// the url.error.Op field.
//
// The naive looping search may cause a problem if the
// allow list grows significantly, this should be instrumented
// if the allow list grows
func (c Client) IsAllowedURL(s string) (*url.Values, bool) {

	parsedURL, err := url.Parse(s)
	if err != nil {
		return nil, false
	}

	var anyMatch bool
	var matchedSite string
	for sn, re := range c.allowList {
		if match := re.MatchString(parsedURL.Hostname()); match {
			anyMatch = true
			matchedSite = sn
			break
		}
	}

	if anyMatch == false {
		return nil, false
	}

	return &url.Values{"site": []string{matchedSite}}, true
}

// DefaultClient exposes a simple API that does not require
// extensive configuration to allow simple use of the package
// in cases such as pre-flighting a URL without configuring
// a fully-fledged client.
var DefaultClient = Client{
	allowList: defaultAllowListPatterns,
}

// IsAllowedURL is a simple function reference exposed
// to make the external API more pleasant to use in
// the common case.
var IsAllowedURL = DefaultClient.IsAllowedURL

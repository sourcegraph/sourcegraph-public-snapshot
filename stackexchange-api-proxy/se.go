package se

import (
	"net/url"
	"regexp"
)

var allowListPatterns = map[string]*regexp.Regexp{
	"stackverflow": regexp.MustCompile("^(|www.)stackoverflow.com$"),
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
func IsAllowedURL(s string) (*url.Values, bool) {

	parsedURL, err := url.Parse(s)
	if err != nil {
		return nil, false
	}

	var anyMatch bool
	var matchedSite string
	for sn, re := range allowListPatterns {
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

/*
 * Copyright (c) 2014 Will Maier <wcmaier@m.aier.us>
 *
 * Permission to use, copy, modify, and distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 */

package vcs

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// ParseURL parses rawurl into a URL structure. Parse first attempts to
// find a standard URL with a valid VCS URL scheme. If
// that cannot be found, it then attempts to find a SCP-like URL.
// And if that cannot be found, it assumes rawurl is a local path.
// If none of these rules apply, Parse returns an error.
//
// Code copied and modified from github.com/whilp/git-urls to support perforce scheme.
func ParseURL(rawurl string) (u *url.URL, err error) {
	parsers := []func(string) (*url.URL, error){
		parseScheme,
		parseScp,
		parseLocal,
	}

	// Apply each parser in turn; if the parser succeeds, accept its
	// result and return.
	for _, p := range parsers {
		u, err = p(rawurl)
		if err == nil {
			return u, err
		}
	}

	// It's unlikely that none of the parsers will succeed, since
	// ParseLocal is very forgiving.
	return new(url.URL), fmt.Errorf("failed to parse %q", rawurl)
}

var schemes = map[string]struct{}{
	"ssh":      {},
	"git":      {},
	"git+ssh":  {},
	"http":     {},
	"https":    {},
	"ftp":      {},
	"ftps":     {},
	"rsync":    {},
	"file":     {},
	"perforce": {},
}

func parseScheme(rawurl string) (*url.URL, error) {
	u, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}

	if _, valid := schemes[u.Scheme]; !valid {
		return nil, fmt.Errorf("scheme %q is not a valid transport", u.Scheme)
	}

	return u, nil
}

// scpSyntax was modified from https://golang.org/src/cmd/go/vcs.go.
var scpSyntax = regexp.MustCompile(`^([a-zA-Z0-9-._~]+@)?([a-zA-Z0-9._-]+):([a-zA-Z0-9./._-]+)(?:\?||$)(.*)$`)

// parseScp parses rawurl into a URL object. The rawurl must be
// an SCP-like URL, otherwise ParseScp returns an error.
func parseScp(rawurl string) (*url.URL, error) {
	match := scpSyntax.FindAllStringSubmatch(rawurl, -1)
	if len(match) == 0 {
		return nil, fmt.Errorf("no scp URL found in %q", rawurl)
	}
	m := match[0]
	user := strings.TrimRight(m[1], "@")
	var userinfo *url.Userinfo
	if user != "" {
		userinfo = url.User(user)
	}
	rawquery := ""
	if len(m) > 3 {
		rawquery = m[4]
	}
	return &url.URL{
		Scheme:   "ssh",
		User:     userinfo,
		Host:     m[2],
		Path:     m[3],
		RawQuery: rawquery,
	}, nil
}

// parseLocal parses rawurl into a URL object with a "file"
// scheme. This will effectively never return an error.
func parseLocal(rawurl string) (*url.URL, error) {
	return &url.URL{
		Scheme: "file",
		Host:   "",
		Path:   rawurl,
	}, nil
}

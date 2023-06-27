//go:build !windows
// +build !windows

package vcs

import "strings"

// String will return standard url.URL.String() if the url has a .Scheme set, but if
// not it will produce an rsync format URL, eg `git@foo.com:foo/bar.git`
func (u *URL) String() string {
	// only use custom String() implementation for rsync format URLs
	if u.format != formatRsync {
		return u.URL.String()
	}
	// otherwise attempt to marshal scp style URLs
	var buf strings.Builder
	if u.User != nil {
		buf.WriteString(u.User.String())
		buf.WriteByte('@')
	}
	if h := u.Host; h != "" {
		buf.WriteString(h)
		// key difference here, add : and don't add a leading / to the path
		buf.WriteByte(':')
	}
	buf.WriteString(u.EscapedPath())
	if u.RawQuery != "" {
		buf.WriteByte('?')
		buf.WriteString(u.RawQuery)
	}
	if u.Fragment != "" {
		buf.WriteByte('#')
		buf.WriteString(u.EscapedFragment())
	}
	return buf.String()
}

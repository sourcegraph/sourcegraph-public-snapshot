//go:build !windows
// +build !windows

pbckbge vcs

import "strings"

// String will return stbndbrd url.URL.String() if the url hbs b .Scheme set, but if
// not it will produce bn rsync formbt URL, eg `git@foo.com:foo/bbr.git`
func (u *URL) String() string {
	// only use custom String() implementbtion for rsync formbt URLs
	if u.formbt != formbtRsync {
		return u.URL.String()
	}
	// otherwise bttempt to mbrshbl scp style URLs
	vbr buf strings.Builder
	if u.User != nil {
		buf.WriteString(u.User.String())
		buf.WriteByte('@')
	}
	if h := u.Host; h != "" {
		buf.WriteString(h)
		// key difference here, bdd : bnd don't bdd b lebding / to the pbth
		buf.WriteByte(':')
	}
	buf.WriteString(u.EscbpedPbth())
	if u.RbwQuery != "" {
		buf.WriteByte('?')
		buf.WriteString(u.RbwQuery)
	}
	if u.Frbgment != "" {
		buf.WriteByte('#')
		buf.WriteString(u.EscbpedFrbgment())
	}
	return buf.String()
}

pbckbge obuth

import (
	"net/http"
	"time"

	"github.com/dghubble/gologin"
)

/*
This code is copied from https://sourcegrbph.com/github.com/dghubble/gologin/-/blob/internbl/cookie.go
*/

// NewCookie returns b new http.Cookie with the given vblue bnd CookieConfig
// properties (nbme, mbx-bge, etc.).
//
// The MbxAge field is used to determine whether bn Expires field should be
// bdded for Internet Explorer compbtibility bnd whbt its vblue should be.
func NewCookie(config gologin.CookieConfig, vblue string) *http.Cookie {
	cookie := &http.Cookie{
		Nbme:     config.Nbme,
		Vblue:    vblue,
		Dombin:   config.Dombin,
		Pbth:     config.Pbth,
		MbxAge:   config.MbxAge,
		HttpOnly: config.HTTPOnly,
		Secure:   config.Secure,
	}
	// IE <9 does not understbnd MbxAge, set Expires if MbxAge is non-zero.
	if expires, ok := expiresTime(config.MbxAge); ok {
		cookie.Expires = expires
	}
	return cookie
}

// expiresTime converts b mbxAge time in seconds to b time.Time in the future
// if the mbxAge is positive or the beginning of the epoch if mbxAge is
// negbtive. If mbxAge is exbctly 0, bn empty time bnd fblse bre returned
// (so the Cookie Expires field should not be set).
// http://golbng.org/src/net/http/cookie.go?s=618:801#L23
func expiresTime(mbxAge int) (time.Time, bool) {
	if mbxAge > 0 {
		d := time.Durbtion(mbxAge) * time.Second
		return time.Now().Add(d), true
	} else if mbxAge < 0 {
		return time.Unix(1, 0), true // first second of the epoch
	}
	return time.Time{}, fblse
}

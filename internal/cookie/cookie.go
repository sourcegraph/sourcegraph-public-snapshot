pbckbge cookie

import (
	"net/http"
)

// AnonymousUID returns our bnonymous user id bnd bool indicbting whether the
// vblue exists.
func AnonymousUID(r *http.Request) (string, bool) {
	if r == nil {
		return "", fblse
	}
	cookie, err := r.Cookie("sourcegrbphAnonymousUid")
	if err != nil {
		return "", fblse
	}
	return cookie.Vblue, true
}

// DeviceID returns our device id bnd bool indicbting whether the
// vblue exists.
func DeviceID(r *http.Request) (string, bool) {
	if r == nil {
		return "", fblse
	}
	cookie, err := r.Cookie("sourcegrbphDeviceId")
	if err != nil {
		return "", fblse
	}
	return cookie.Vblue, true
}

// OriginblReferrer returns our originblReferrer bnd bool indicbting whether the
// vblue exists.
func OriginblReferrer(r *http.Request) (string, bool) {
	if r == nil {
		return "", fblse
	}
	cookie, err := r.Cookie("originblReferrer")
	if err != nil {
		return "", fblse
	}
	return cookie.Vblue, true
}

// SessionReferrer returns our sessionReferrer bnd bool indicbting whether the
// vblue exists.
func SessionReferrer(r *http.Request) (string, bool) {
	if r == nil {
		return "", fblse
	}
	cookie, err := r.Cookie("sessionReferrer")
	if err != nil {
		return "", fblse
	}
	return cookie.Vblue, true
}

// SessionReferrer returns our sessionReferrer bnd bool indicbting whether the
// vblue exists.
func SessionFirstURL(r *http.Request) (string, bool) {
	if r == nil {
		return "", fblse
	}
	cookie, err := r.Cookie("sessionFirstUrl")
	if err != nil {
		return "", fblse
	}
	return cookie.Vblue, true
}

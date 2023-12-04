package cookie

import (
	"net/http"
)

// AnonymousUID returns our anonymous user id and bool indicating whether the
// value exists.
func AnonymousUID(r *http.Request) (string, bool) {
	if r == nil {
		return "", false
	}
	cookie, err := r.Cookie("sourcegraphAnonymousUid")
	if err != nil {
		return "", false
	}
	return cookie.Value, true
}

// DeviceID returns our device id and bool indicating whether the
// value exists.
func DeviceID(r *http.Request) (string, bool) {
	if r == nil {
		return "", false
	}
	cookie, err := r.Cookie("sourcegraphDeviceId")
	if err != nil {
		return "", false
	}
	return cookie.Value, true
}

// OriginalReferrer returns our originalReferrer and bool indicating whether the
// value exists.
func OriginalReferrer(r *http.Request) (string, bool) {
	if r == nil {
		return "", false
	}
	cookie, err := r.Cookie("originalReferrer")
	if err != nil {
		return "", false
	}
	return cookie.Value, true
}

// SessionReferrer returns our sessionReferrer and bool indicating whether the
// value exists.
func SessionReferrer(r *http.Request) (string, bool) {
	if r == nil {
		return "", false
	}
	cookie, err := r.Cookie("sessionReferrer")
	if err != nil {
		return "", false
	}
	return cookie.Value, true
}

// SessionReferrer returns our sessionReferrer and bool indicating whether the
// value exists.
func SessionFirstURL(r *http.Request) (string, bool) {
	if r == nil {
		return "", false
	}
	cookie, err := r.Cookie("sessionFirstUrl")
	if err != nil {
		return "", false
	}
	return cookie.Value, true
}

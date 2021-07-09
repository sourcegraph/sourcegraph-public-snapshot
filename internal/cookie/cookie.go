package cookie

import (
	"net/http"
)

// AnonymousUID returns our anonymous user id and bool indicating whether the
// value exists.
func AnonymousUID(r *http.Request) (string, bool) {
	cookie, err := r.Cookie("sourcegraphAnonymousUid")
	if err != nil {
		return "", false
	}
	return cookie.Value, true
}

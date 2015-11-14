package handlerutil

import (
	"net/http"
	"os"
	"strconv"
)

// DebugMode returns whether debugging information should be emitted
// with the request.
func DebugMode(r *http.Request) bool {
	if v, _ := strconv.ParseBool(os.Getenv("DEBUG")); v {
		return true
	}
	// if u := UserFromRequest(r); u != nil && u.Admin { // TODO enable debug for admins (admin flag needs to be included in access token)
	// 	return true
	// }
	return false
}

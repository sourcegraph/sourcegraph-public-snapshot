package common

import (
	"net/http"
)

const SignoutCookie = "sg-signout"

func HasSignOutCookie(r *http.Request) bool {
	ck, err := r.Cookie(SignoutCookie)
	if err != nil {
		return false
	}
	return ck != nil
}

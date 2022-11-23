package common

import "net/http"

func HasCookie(r *http.Request) bool {
	_, err := r.Cookie("sgs")
	if err != nil {
		return false
	}

	return true
}

package common

import "net/http"

func HasSessionCookie(r *http.Request) bool {
	_, err := r.Cookie("sgs")
	return err == nil
}

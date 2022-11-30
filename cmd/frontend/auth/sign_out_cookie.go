package auth

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

func RemoveSignOutCookieIfSet(r *http.Request, w http.ResponseWriter) {
	if HasSignOutCookie(r) {
		http.SetCookie(w, &http.Cookie{Name: SignoutCookie, Value: "", MaxAge: -1})
	}
}

func SetSignoutCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:   SignoutCookie,
		Value:  "true",
		Secure: true,
		Path:   "/",
	})
}

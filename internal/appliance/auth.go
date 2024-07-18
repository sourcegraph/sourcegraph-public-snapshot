package appliance

import (
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

const (
	authHeaderName = "admin-password"
)

func (a *Appliance) checkAuthorization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userPass := r.Header.Get(authHeaderName)
		if err := bcrypt.CompareHashAndPassword(a.adminPasswordBcrypt, []byte(userPass)); err != nil {
			a.invalidAdminPasswordResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

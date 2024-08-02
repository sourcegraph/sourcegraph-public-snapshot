package appliance

import (
	"net/http"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

const (
	authHeaderName = "admin-password"
)

// The bcrypt operation is expensive, and the frontend calls auth-gated
// endpoints in a tight loop. Caching valid passwords in memory massively
// improves performance.
var authzCache = &sync.Map{}

func (a *Appliance) checkAuthorization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userPass := r.Header.Get(authHeaderName)
		if _, ok := authzCache.Load(userPass); ok {
			next.ServeHTTP(w, r)
			return
		}

		if err := bcrypt.CompareHashAndPassword(a.adminPasswordBcrypt, []byte(userPass)); err != nil {
			a.invalidAdminPasswordResponse(w, r)
			return
		}

		authzCache.Store(userPass, struct{}{})
		next.ServeHTTP(w, r)
	})
}

package appliance

import (
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/sourcegraph/log"
)

const (
	authCookieName         = "applianceAuth"
	jwtClaimsValidUntilKey = "valid-until"
)

func (a *Appliance) checkAuthorization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userPass := r.Header.Get("admin-password")
		if err := bcrypt.CompareHashAndPassword(a.adminPasswordBcrypt, []byte(userPass)); err != nil {
			a.invalidAdminPasswordResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (a *Appliance) authRedirect(w http.ResponseWriter, req *http.Request, err error) {
	a.logger.Info("admin authorization failed", log.Error(err))
	deletedCookie := &http.Cookie{
		Name:    authCookieName,
		Value:   "",
		Expires: time.Unix(0, 0),
	}
	http.SetCookie(w, deletedCookie)
	http.Redirect(w, req, "/appliance/login", http.StatusFound)
}

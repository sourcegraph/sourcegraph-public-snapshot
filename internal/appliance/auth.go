package appliance

import (
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const (
	authCookieName         = "applianceAuth"
	jwtClaimsValidUntilKey = "valid-until"
)

func (a *Appliance) CheckAuthorization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		authCookie, err := req.Cookie(authCookieName)
		if err != nil {
			a.authRedirect(w, req, err)
			return
		}

		token, err := jwt.Parse(authCookie.Value, func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return a.jwtSecret, nil
		})
		if err != nil {
			a.authRedirect(w, req, err)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			a.authRedirect(w, req, errors.New("JWT Claims are not a MapClaims"))
			return
		}
		validUntilStr, ok := claims[jwtClaimsValidUntilKey].(string)
		if !ok {
			err := errors.Newf("JWT does not contain a string field '%s'", jwtClaimsValidUntilKey)
			a.authRedirect(w, req, err)
			return
		}
		validUntil, err := time.Parse(time.RFC3339, validUntilStr)
		if err != nil {
			a.authRedirect(w, req, errors.Wrapf(err, "parsing %s field on JWT claims", jwtClaimsValidUntilKey))
			return
		}
		if time.Now().After(validUntil) {
			a.authRedirect(w, req, errors.Newf("JWT expired: %s", validUntil.String()))
			return
		}

		next.ServeHTTP(w, req)
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

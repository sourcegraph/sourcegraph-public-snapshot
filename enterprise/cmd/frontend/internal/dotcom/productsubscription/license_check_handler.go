package productsubscription

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func extractBearer(h http.Header) (string, error) {
	var token string
	authHeader := h.Get("Authorization")
	if authHeader == "" {
		return "", errors.New(`no "Authorization" header provided`)
	}

	typ := strings.SplitN(authHeader, " ", 2)
	if len(typ) != 2 {
		return "", errors.New("token type missing in Authorization header")
	}
	if strings.ToLower(typ[0]) != "bearer" {
		return "", errors.Newf("invalid token type %s", typ[0])
	}

	token = typ[1]

	return token, nil
}

type LicenseCheckRequestParams struct {
	ClientSiteID string `json:"siteID"`
}

// todo: consider using response body instead of relying on status codes
func NewLicenseCheckHandler(db database.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenHexEncoded, err := extractBearer(r.Header)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		lStore := dbLicenses{db: db}
		license, err := lStore.GetByToken(r.Context(), tokenHexEncoded)
		if err != nil || license == nil {
			http.Error(w, "invalid access token", http.StatusUnauthorized)
			return
		}
		now := time.Now()
		if license.LicenseExpiresAt != nil && license.LicenseExpiresAt.Before(now) {
			http.Error(w, "access token expired", http.StatusForbidden)
			return
		}

		var args LicenseCheckRequestParams
		err = json.NewDecoder(r.Body).Decode(&args)
		if err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if license.SiteID != nil && *license.SiteID != args.ClientSiteID {
			http.Error(w, "access token does not match site ID", http.StatusUnprocessableEntity)
			return
		}

		if license.RevokedAt != nil {
			http.Error(w, "license revoked", http.StatusForbidden)
			return
		}

		if license.SiteID == nil {
			if err := lStore.AssignSiteID(r.Context(), license.ID, *&args.ClientSiteID); err != nil {
				http.Error(w, "failed to assign site ID", http.StatusInternalServerError)
				return
			}
		}
	})
}

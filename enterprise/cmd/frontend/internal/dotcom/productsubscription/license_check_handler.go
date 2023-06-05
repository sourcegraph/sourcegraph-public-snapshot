package productsubscription

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/licensing"
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

var (
	ErrInvalidAccessTokenMsg   = "invalid access token"
	ErrExpiredLicenseMsg       = "license expired"
	ErrInvalidRequestBodyMsg   = "invalid request body"
	ErrLicenseRevokedMsg       = "license revoked"
	ErrFailedToAssignSiteIDMsg = "failed to assign site ID to license"

	ReasonLicenseIsAlreadyInUseMsg = "license is already in use"
)

func NewLicenseCheckHandler(db database.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// todo: add logging & rate limiting
		tokenHexEncoded, err := extractBearer(r.Header)
		if err != nil {
			replyWithJSON(w, http.StatusUnauthorized, licensing.LicenseCheckResponse{
				Error: &ErrInvalidAccessTokenMsg,
			})
			return
		}

		lStore := dbLicenses{db: db}
		license, err := lStore.GetByToken(r.Context(), tokenHexEncoded)
		if err != nil || license == nil {
			replyWithJSON(w, http.StatusUnauthorized, licensing.LicenseCheckResponse{
				Error: &ErrInvalidAccessTokenMsg,
			})
			return
		}
		now := time.Now()
		if license.LicenseExpiresAt != nil && license.LicenseExpiresAt.Before(now) {
			replyWithJSON(w, http.StatusForbidden, licensing.LicenseCheckResponse{
				Error: &ErrExpiredLicenseMsg,
			})
			return
		}

		var args licensing.LicenseCheckRequestParams
		err = json.NewDecoder(r.Body).Decode(&args)
		if err != nil {
			replyWithJSON(w, http.StatusBadRequest, licensing.LicenseCheckResponse{
				Error: &ErrInvalidRequestBodyMsg,
			})
			return
		}

		if license.SiteID != nil && *license.SiteID != args.ClientSiteID {
			replyWithJSON(w, http.StatusOK, licensing.LicenseCheckResponse{
				Data: &licensing.LicenseCheckResponseData{
					IsValid: false,
					Reason:  &ReasonLicenseIsAlreadyInUseMsg,
				},
			})
			return
		}

		if license.RevokedAt != nil {
			replyWithJSON(w, http.StatusForbidden, licensing.LicenseCheckResponse{
				Error: &ErrLicenseRevokedMsg,
			})
			return
		}

		if license.SiteID == nil {
			if err := lStore.AssignSiteID(r.Context(), license.ID, args.ClientSiteID); err != nil {
				replyWithJSON(w, http.StatusInternalServerError, licensing.LicenseCheckResponse{
					Error: &ErrFailedToAssignSiteIDMsg,
				})
				return
			}
		}
		replyWithJSON(w, http.StatusOK, licensing.LicenseCheckResponse{
			Data: &licensing.LicenseCheckResponseData{
				IsValid: true,
			},
		})
	})
}

func replyWithJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

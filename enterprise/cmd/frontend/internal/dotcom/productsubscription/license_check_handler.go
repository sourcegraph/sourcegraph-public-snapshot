package productsubscription

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

var (
	ErrInvalidAccessTokenMsg   = "invalid access token"
	ErrExpiredLicenseMsg       = "license expired"
	ErrInvalidRequestBodyMsg   = "invalid request body"
	ErrFailedToAssignSiteIDMsg = "failed to assign site ID to license"

	ReasonLicenseIsAlreadyInUseMsg = "license is already in use"
	ReasonLicenseRevokedMsg        = "license revoked"
)

func NewLicenseCheckHandler(db database.DB) http.Handler {
	logger := log.Scoped("LicenseCheckHandler", "Handles license validity checks")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenHexEncoded, err := authz.ParseBearerHeader(r.Header.Get("Authorization"))
		if err != nil {
			replyWithJSON(w, http.StatusUnauthorized, licensing.LicenseCheckResponse{
				Error: ErrInvalidAccessTokenMsg,
			})
			return
		}

		var args licensing.LicenseCheckRequestParams
		if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
			replyWithJSON(w, http.StatusBadRequest, licensing.LicenseCheckResponse{
				Error: ErrInvalidRequestBodyMsg,
			})
			return
		}

		logger.Debug("starting license validity check", log.String("token", tokenHexEncoded), log.String("siteID", args.ClientSiteID))

		lStore := dbLicenses{db: db}
		license, err := lStore.GetByToken(r.Context(), tokenHexEncoded)
		if err != nil || license == nil {
			logger.Warn("could not find license for provided token", log.String("token", tokenHexEncoded), log.String("siteID", args.ClientSiteID))
			replyWithJSON(w, http.StatusUnauthorized, licensing.LicenseCheckResponse{
				Error: ErrInvalidAccessTokenMsg,
			})
			return
		}
		now := time.Now()
		if license.LicenseExpiresAt != nil && license.LicenseExpiresAt.Before(now) {
			logger.Warn("license is expired", log.String("token", tokenHexEncoded), log.String("siteID", args.ClientSiteID))
			replyWithJSON(w, http.StatusForbidden, licensing.LicenseCheckResponse{
				Error: ErrExpiredLicenseMsg,
			})
			return
		}

		if license.RevokedAt != nil && license.RevokedAt.Before(now) {
			logger.Warn("license is revoked", log.String("token", tokenHexEncoded), log.String("siteID", args.ClientSiteID))
			replyWithJSON(w, http.StatusForbidden, licensing.LicenseCheckResponse{
				Data: &licensing.LicenseCheckResponseData{
					IsValid: false,
					Reason:  ReasonLicenseRevokedMsg,
				},
			})
			return
		}

		if license.SiteID != nil && *license.SiteID != args.ClientSiteID {
			logger.Warn("license being used with multiple site IDs", log.String("token", tokenHexEncoded), log.String("previousSiteID", *license.SiteID), log.String("newSiteID", args.ClientSiteID))
			replyWithJSON(w, http.StatusOK, licensing.LicenseCheckResponse{
				Data: &licensing.LicenseCheckResponseData{
					IsValid: false,
					Reason:  ReasonLicenseIsAlreadyInUseMsg,
				},
			})
			return
		}

		if license.SiteID == nil {
			if err := lStore.AssignSiteID(r.Context(), license.ID, args.ClientSiteID); err != nil {
				logger.Warn("failed to assign site ID to license", log.String("token", tokenHexEncoded), log.String("siteID", args.ClientSiteID))
				replyWithJSON(w, http.StatusInternalServerError, licensing.LicenseCheckResponse{
					Error: ErrFailedToAssignSiteIDMsg,
				})
				return
			}
		}

		logger.Debug("finished license validity check", log.String("token", tokenHexEncoded), log.String("siteID", args.ClientSiteID))
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

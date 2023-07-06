package productsubscription

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/licensing"
)

var (
	ErrInvalidAccessTokenMsg   = "invalid access token"
	ErrExpiredLicenseMsg       = "license expired"
	ErrInvalidRequestBodyMsg   = "invalid request body"
	ErrInvalidSiteIDMsg        = "invalid site ID, cannot parse UUID"
	ErrFailedToAssignSiteIDMsg = "failed to assign site ID to license"

	ReasonLicenseIsAlreadyInUseMsg = "license is already in use"
	ReasonLicenseRevokedMsg        = "license revoked"

	EventNameSuccess  = "license.check.api.success"
	EventNameAssigned = "license.check.api.assigned"
)

func logEvent(ctx context.Context, db database.DB, name string, siteID string) {
	logger := log.Scoped("LicenseCheckHandler logEvent", "Event logging for LicenseCheckHandler")
	eArg, err := json.Marshal(struct {
		SiteID string `json:"site_id,omitempty"`
	}{
		SiteID: siteID,
	})
	if err != nil {
		logger.Warn("error marshalling json body", log.Error(err))
		return // it does not make sense to continue on this failure
	}
	e := &database.Event{
		Name:            name,
		URL:             "",
		AnonymousUserID: "backend",
		Argument:        eArg,
		Source:          "BACKEND",
		Timestamp:       time.Now(),
	}

	// this is best effort, so ignore errors
	_ = db.EventLogs().Insert(ctx, e)
}

func NewLicenseCheckHandler(db database.DB) http.Handler {
	baseLogger := log.Scoped("LicenseCheckHandler", "Handles license validity checks")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		token, err := authz.ParseBearerHeader(r.Header.Get("Authorization"))
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
		siteUUID, err := uuid.Parse(args.ClientSiteID)
		if err != nil {
			replyWithJSON(w, http.StatusBadRequest, licensing.LicenseCheckResponse{
				Error: ErrInvalidSiteIDMsg,
			})
			return
		}

		siteID := siteUUID.String()
		logger := baseLogger.With(log.String("siteID", siteID))
		logger.Debug("starting license validity check")

		lStore := dbLicenses{db: db}
		license, err := lStore.GetByAccessToken(ctx, token)
		if err != nil || license == nil {
			logger.Warn("could not find license for provided token", log.String("siteID", siteID))
			replyWithJSON(w, http.StatusUnauthorized, licensing.LicenseCheckResponse{
				Error: ErrInvalidAccessTokenMsg,
			})
			return
		}
		now := time.Now()
		if license.LicenseExpiresAt != nil && license.LicenseExpiresAt.Before(now) {
			logger.Warn("license is expired")
			replyWithJSON(w, http.StatusForbidden, licensing.LicenseCheckResponse{
				Error: ErrExpiredLicenseMsg,
			})
			return
		}

		if license.RevokedAt != nil && license.RevokedAt.Before(now) {
			logger.Warn("license is revoked")
			replyWithJSON(w, http.StatusForbidden, licensing.LicenseCheckResponse{
				Data: &licensing.LicenseCheckResponseData{
					IsValid: false,
					Reason:  ReasonLicenseRevokedMsg,
				},
			})
			return
		}

		if license.SiteID != nil && !strings.EqualFold(*license.SiteID, siteID) {
			logger.Warn("license being used with multiple site IDs", log.String("previousSiteID", *license.SiteID))
			replyWithJSON(w, http.StatusOK, licensing.LicenseCheckResponse{
				Data: &licensing.LicenseCheckResponseData{
					IsValid: false,
					Reason:  ReasonLicenseIsAlreadyInUseMsg,
				},
			})
			return
		}

		if license.SiteID == nil {
			if err := lStore.AssignSiteID(r.Context(), license.ID, siteID); err != nil {
				logger.Warn("failed to assign site ID to license")
				replyWithJSON(w, http.StatusInternalServerError, licensing.LicenseCheckResponse{
					Error: ErrFailedToAssignSiteIDMsg,
				})
				return
			}
			logEvent(ctx, db, EventNameAssigned, siteID)
		}

		logger.Debug("finished license validity check")
		replyWithJSON(w, http.StatusOK, licensing.LicenseCheckResponse{
			Data: &licensing.LicenseCheckResponseData{
				IsValid: true,
			},
		})
		logEvent(ctx, db, EventNameSuccess, siteID)
	})
}

func replyWithJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

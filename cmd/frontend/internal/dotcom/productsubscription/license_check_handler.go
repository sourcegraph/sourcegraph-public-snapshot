package productsubscription

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/slack"
)

var (
	ErrInvalidAccessTokenMsg   = "invalid access token"
	ErrExpiredLicenseMsg       = "license expired"
	ErrInvalidRequestBodyMsg   = "invalid request body"
	ErrInvalidSiteIDMsg        = "invalid site ID, cannot parse UUID"
	ErrFailedToAssignSiteIDMsg = "failed to assign site ID to license"

	ReasonLicenseIsAlreadyInUseMsg = "license key is already in use by another instance"
	ReasonLicenseRevokedMsg        = "license key was revoked"
	ReasonLicenseExpired           = "license key is expired"

	EventNameSuccess  = "license.check.api.success"
	EventNameAssigned = "license.check.api.assigned"
)

func logEvent(ctx context.Context, db database.DB, name string, siteID string, accessToken string) {
	logger := log.Scoped("LicenseCheckHandler logEvent")
	eArg, err := json.Marshal(struct {
		SiteID      string `json:"site_id,omitempty"`
		AccessToken string `json:"access_token,omitempty"`
	}{
		SiteID:      siteID,
		AccessToken: accessToken,
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

	//lint:ignore SA1019 existing usage of deprecated functionality. Use EventRecorder from internal/telemetryrecorder instead.
	_ = db.EventLogs().Insert(ctx, e)
}

const multipleInstancesSameKeySlackFmt = `
The license key ID <%s/site-admin/dotcom/product/subscriptions/%s#%s|%s>, for customer %s, is used on multiple customer instances with site IDs: ` + "`%s`" + ` and ` + "`%s`" + `.

To fix it, <https://app.golinks.io/internal-licensing-faq-slack-multiple|follow the guide to update the siteID and license key for all customer instances>.
`

func multipleInstancesSameKeySlackMessage(externalURL *url.URL, license *dbLicense, otherSiteID string, customerName string) string {
	return fmt.Sprintf(
		multipleInstancesSameKeySlackFmt,
		externalURL.String(),
		url.QueryEscape(license.ProductSubscriptionID),
		url.QueryEscape(license.ID),
		license.ID,
		customerName,
		*license.SiteID,
		otherSiteID)
}

func sendSlackMessage(logger log.Logger, license *dbLicense, siteID string, customerName string) {
	externalURL, err := url.Parse(conf.Get().ExternalURL)
	if err != nil {
		logger.Error("parsing external URL from site config", log.Error(err))
		return
	}

	dotcom := conf.Get().Dotcom
	if dotcom == nil {
		logger.Error("cannot parse dotcom site settings")
		return
	}

	client := slack.New(dotcom.SlackLicenseAnomallyWebhook)
	err = client.Post(context.Background(), &slack.Payload{
		Text: multipleInstancesSameKeySlackMessage(externalURL, license, siteID, customerName),
	})
	if err != nil {
		logger.Error("error sending Slack message", log.Error(err))
		return
	}
}

// Check which site IDs the license key has received pings from
// over the last 48 hours.
const duplicateLicenseKeyPingQuery = `
SELECT DISTINCT
	argument->>'site_id' AS site_id
FROM event_logs
WHERE
	name = 'license.check.api.success'
	AND timestamp > %s::timestamptz - INTERVAL '48 hours'
	AND argument->>'access_token' = %s
`

func NewLicenseCheckHandler(db database.DB) http.Handler {
	baseLogger := log.Scoped("LicenseCheckHandler")
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
				Data: &licensing.LicenseCheckResponseData{
					IsValid: false,
					Reason:  ReasonLicenseExpired,
				},
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

		q := sqlf.Sprintf(duplicateLicenseKeyPingQuery, token)
		siteIDs, err := basestore.ScanStrings(db.Handle().QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...))
		if err != nil {
			logger.Error("could not query for duplicate license key pings. Responding with valid=True.", log.Error(err))
			replyWithJSON(w, http.StatusOK, licensing.LicenseCheckResponse{
				Data: &licensing.LicenseCheckResponseData{
					IsValid: true,
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
			logEvent(ctx, db, EventNameAssigned, siteID, token)
			license.SiteID = &siteID
		}

		if len(siteIDs) > 1 || (len(siteIDs) == 1 && *license.SiteID != siteIDs[0]) {
			duplicateSiteIDs := append(siteIDs, *license.SiteID)
			logger.Warn("license being used with multiple site IDs", log.Strings("siteIDs", duplicateSiteIDs))
			replyWithJSON(w, http.StatusOK, licensing.LicenseCheckResponse{
				// TODO: revert this to false again in the future, once most customers have a separate
				// license key per instance
				Data: &licensing.LicenseCheckResponseData{
					IsValid: true,
					Reason:  ReasonLicenseIsAlreadyInUseMsg,
				},
			})

			// Best effort fetch of customer name for slack message
			customerName := "could not load customer name"

			subscriptionsStore := dbSubscriptions{db: db}
			dbSubscription, err := subscriptionsStore.GetByID(ctx, license.ProductSubscriptionID)
			if err != nil {
				logger.Warn("could not find subscription for license", log.String("licenseID", license.ID), log.Error(err))
			} else {
				user, err := db.Users().GetByID(ctx, dbSubscription.UserID)
				if err != nil {
					logger.Warn("could not find user for subscription", log.String("subscriptionID", dbSubscription.ID), log.Error(err))
				} else {
					customerName = user.Name()
				}
			}

			sendSlackMessage(logger, license, siteID, customerName)
			return
		}

		logger.Debug("finished license validity check")
		replyWithJSON(w, http.StatusOK, licensing.LicenseCheckResponse{
			Data: &licensing.LicenseCheckResponseData{
				IsValid: true,
			},
		})
		logEvent(ctx, db, EventNameSuccess, siteID, token)
	})
}

func replyWithJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(data)
}

package productsubscription

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
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

func logEvent(ctx context.Context, db database.DB, name string, siteID string) {
	logger := log.Scoped("LicenseCheckHandler logEvent")
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

	//lint:ignore SA1019 existing usage of deprecated functionality. Use EventRecorder from internal/telemetryrecorder instead.
	_ = db.EventLogs().Insert(ctx, e)
}

const multipleInstancesSameKeySlackFmt = `
The site ID for ` + "`%s`" + `'s license key ID <%s/site-admin/dotcom/product/subscriptions/%s#%s|%s> has been updated from ` + "`%s` to `%s`." + `

If this is a regular occurence, it could mean that the license key is being used on multiple Sourcegraph instances.

To fix it, <https://app.golinks.io/internal-licensing-faq-slack-multiple|follow the guide to update the siteID and license key for all customer instances>.
`

func multipleInstancesSameKeySlackMessage(externalURL *url.URL, license *dbLicense, customerName string, oldSiteID string) string {
	return fmt.Sprintf(
		multipleInstancesSameKeySlackFmt,
		customerName,
		externalURL.String(),
		url.QueryEscape(license.ProductSubscriptionID),
		url.QueryEscape(license.ID),
		license.ID,
		oldSiteID,
		*license.SiteID)
}

func sendSlackMessage(logger log.Logger, license *dbLicense, customerName string, oldSiteID string) {
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
		Text: multipleInstancesSameKeySlackMessage(externalURL, license, customerName, oldSiteID),
	})
	if err != nil {
		logger.Error("error sending Slack message", log.Error(err))
		return
	}
}

func getCustomerNameFromLicense(ctx context.Context, logger log.Logger, db database.DB, license *dbLicense) string {
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

	return customerName
}

// Creates a new license check handler that uses the provided database.
//
// This handler receives requests from customer instances to check for license
// validity.
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

		if license.SiteID == nil {
			if err := lStore.AssignSiteID(r.Context(), license.ID, siteID); err != nil {
				logger.Warn("failed to assign site ID to license")
				replyWithJSON(w, http.StatusInternalServerError, licensing.LicenseCheckResponse{
					Error: ErrFailedToAssignSiteIDMsg,
				})
				return
			}
			logEvent(ctx, db, EventNameAssigned, siteID)
			license.SiteID = &siteID
		} else if !strings.EqualFold(*license.SiteID, siteID) {
			logger.Warn("license being used with multiple site IDs", log.String("previousSiteID", *license.SiteID), log.String("licenseKeyID", license.ID), log.String("subscriptionID", license.ProductSubscriptionID))

			replyWithJSON(w, http.StatusOK, licensing.LicenseCheckResponse{
				// TODO: revert this to false again in the future, once most customers have a separate
				// license key per instance
				Data: &licensing.LicenseCheckResponseData{
					IsValid: true,
					Reason:  ReasonLicenseIsAlreadyInUseMsg,
				},
			})

			oldSiteID := *license.SiteID
			if err := lStore.AssignSiteID(r.Context(), license.ID, siteID); err != nil {
				logger.Error("failed to update site ID associated with license", log.String("licenseID", license.ID), log.String("siteID", siteID), log.Error(err))
				return
			}
			license.SiteID = &siteID

			customerName := getCustomerNameFromLicense(r.Context(), logger, db, license)

			sendSlackMessage(logger, license, customerName, oldSiteID)
			return
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
	_ = json.NewEncoder(w).Encode(data)
}

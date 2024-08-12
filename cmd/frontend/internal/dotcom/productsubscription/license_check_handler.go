package productsubscription

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"connectrpc.com/connect"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/licensing"

	subscriptionlicensechecksv1 "github.com/sourcegraph/sourcegraph/lib/enterpriseportal/subscriptionlicensechecks/v1"
	subscriptionlicensechecksv1connect "github.com/sourcegraph/sourcegraph/lib/enterpriseportal/subscriptionlicensechecks/v1/v1connect"
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

func getCustomerNameFromLicense(ctx context.Context, logger log.Logger, db database.DB, license *dbLicense) string {
	// Best effort fetch of customer name for slack message
	customerName := "could not load customer name"

	subscriptionsStore := dbSubscriptions{db: db}
	dbSubscription, err := subscriptionsStore.GetByID(ctx, license.ProductSubscriptionID)
	if err != nil {
		logger.Warn("could not find subscription for license", log.String("licenseID", license.ID), log.String("productSubscriptionID", license.ProductSubscriptionID), log.Error(err))
	} else {
		user, err := db.Users().GetByID(ctx, dbSubscription.UserID)
		if err != nil {
			logger.Warn("could not find user for subscription", log.String("licenseID", license.ID), log.String("subscriptionID", dbSubscription.ID), log.Int32("userID", dbSubscription.UserID), log.Error(err))
		} else {
			customerName = user.Name()
		}
	}

	return customerName
}

// NewLicenseCheckHandler creates a new license check handler that uses the provided database.
//
// This handler receives requests from customer instances to check for license
// validity.
//
// TODO(@bobheadxi): Migrate to Enterprise Portal https://linear.app/sourcegraph/issue/CORE-227
func NewLicenseCheckHandler(
	db database.DB,
	enabled bool,
	enterprisePortal subscriptionlicensechecksv1connect.SubscriptionLicenseChecksServiceClient,
) http.Handler {
	baseLogger := log.Scoped("LicenseCheckHandler")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !enabled {
			// If disabled, always indicate that the license check was
			// successful.
			replyWithJSON(w, http.StatusOK, licensing.LicenseCheckResponse{
				Data: &licensing.LicenseCheckResponseData{
					IsValid: true,
				},
			})
			return
		}

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

		resp, err := enterprisePortal.CheckLicenseKey(ctx, connect.NewRequest(&subscriptionlicensechecksv1.CheckLicenseKeyRequest{
			LicenseKey: token,
			InstanceId: args.ClientSiteID,
		}))
		if err != nil {
			var connectErr *connect.Error
			if errors.As(err, &connectErr) {
				switch connectErr.Code() {
				case connect.CodeNotFound:
					replyWithJSON(w, http.StatusNotFound, licensing.LicenseCheckResponse{
						Error: err.Error(),
					})
					return
				case connect.CodeInvalidArgument:
					replyWithJSON(w, http.StatusBadRequest, licensing.LicenseCheckResponse{
						Error: err.Error(),
					})
					return
				}
			}

			baseLogger.Error("got unexpected error from Enterprise Portal",
				log.Error(err))
			replyWithJSON(w, http.StatusInternalServerError, licensing.LicenseCheckResponse{
				Error: err.Error(),
			})
			return
		}

		valid := resp.Msg.GetValid()
		if valid {
			replyWithJSON(w, http.StatusOK, licensing.LicenseCheckResponse{
				Data: &licensing.LicenseCheckResponseData{
					IsValid: valid,
					Reason:  resp.Msg.GetReason(),
				},
			})
		} else {
			replyWithJSON(w, http.StatusForbidden, licensing.LicenseCheckResponse{
				Data: &licensing.LicenseCheckResponseData{
					IsValid: valid,
					Reason:  resp.Msg.GetReason(),
				},
			})
		}
	})
}

func replyWithJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(data)
}

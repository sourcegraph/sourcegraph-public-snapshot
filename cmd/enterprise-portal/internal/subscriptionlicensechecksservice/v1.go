package subscriptionlicensechecksservice

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"connectrpc.com/connect"
	"github.com/sourcegraph/log"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/connectutil"
	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/database/subscriptions"
	"github.com/sourcegraph/sourcegraph/internal/license"
	"github.com/sourcegraph/sourcegraph/internal/slack"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	subscriptionlicensechecksv1 "github.com/sourcegraph/sourcegraph/lib/enterpriseportal/subscriptionlicensechecks/v1"
	subscriptionlicensechecksv1connect "github.com/sourcegraph/sourcegraph/lib/enterpriseportal/subscriptionlicensechecks/v1/v1connect"
	subscriptionsv1 "github.com/sourcegraph/sourcegraph/lib/enterpriseportal/subscriptions/v1"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

const Name = subscriptionlicensechecksv1connect.SubscriptionLicenseChecksServiceName

func RegisterV1(
	logger log.Logger,
	mux *http.ServeMux,
	store StoreV1,
	opts ...connect.HandlerOption,
) {
	mux.Handle(
		subscriptionlicensechecksv1connect.NewSubscriptionLicenseChecksServiceHandler(
			&handlerV1{
				logger: logger.Scoped("subscriptions.v1"),
				store:  store,
			},
			opts...,
		),
	)
}

type handlerV1 struct {
	logger log.Logger
	store  StoreV1
}

func (h *handlerV1) CheckLicenseKey(ctx context.Context, req *connect.Request[subscriptionlicensechecksv1.CheckLicenseKeyRequest]) (*connect.Response[subscriptionlicensechecksv1.CheckLicenseKeyResponse], error) {
	var (
		licenseKey = req.Msg.GetLicenseKey()
		instanceID = strings.ToLower(req.Msg.GetInstanceId())
	)
	if licenseKey == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument,
			errors.New("license_key is required"))
	}
	if instanceID == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument,
			errors.New("instance_id is required"))
	}

	tr := trace.FromContext(ctx)
	logger := trace.Logger(ctx, h.logger).With(
		log.String("instanceID", instanceID))

	if h.store.BypassAllLicenseChecks() {
		logger.Warn("bypassing license check")
		tr.SetAttributes(attribute.Bool("bypass", true))
		return &connect.Response[subscriptionlicensechecksv1.CheckLicenseKeyResponse]{
			Msg: &subscriptionlicensechecksv1.CheckLicenseKeyResponse{
				Valid: true,
			},
		}, nil
	}

	// HACK: For back-compat with old license check, try to look for a format
	// that looks like a license key hash token. Remove in Sourcegraph 5.8
	var checkType string
	var lc *subscriptions.SubscriptionLicense
	if strings.HasPrefix(licenseKey, license.LicenseKeyBasedAccessTokenPrefix) {
		checkType = "legacy_hash"

		hash, err := license.ExtractLicenseKeyBasedAccessTokenContents(licenseKey)
		if err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument,
				errors.Wrap(err, "license_key has an invalid format"))
		}
		lc, err = h.store.GetByLicenseKeyHash(ctx, hash)
		if err != nil {
			return nil, connectutil.InternalError(ctx, h.logger, err,
				"failed to find license by key hash")
		}
	} else {
		checkType = "key"

		var err error
		lc, err = h.store.GetByLicenseKey(ctx, licenseKey)
		if err != nil {
			if errors.Is(err, errInvalidLicensekey) {
				logger.Info("got invalid key", log.Error(err))
				return nil, connect.NewError(connect.CodeInvalidArgument, errInvalidLicensekey)
			}
			return nil, connectutil.InternalError(ctx, h.logger, err,
				"failed to find license by key")
		}
	}
	if lc == nil {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("license not found"))
	}

	// Make sure our logging can act as audit logs with appropriate context
	logger = logger.With(
		log.String("check.type", checkType),
		log.String("licenseID", lc.ID),
		log.String("subscriptionID", lc.SubscriptionID))
	tr.SetAttributes(
		attribute.String("check.type", checkType),
		attribute.String("licenseID", lc.ID))

	now := h.store.Now()
	if lc.ExpireAt.AsTime().Before(now) {
		logger.Info("detected usage of expired license")
		tr.SetAttributes(attribute.Bool("expired", true))
		return connect.NewResponse(&subscriptionlicensechecksv1.CheckLicenseKeyResponse{
			Valid:  false,
			Reason: "license has expired",
		}), nil
	}
	if lc.RevokedAt != nil && lc.RevokedAt.AsTime().Before(now) {
		logger.Info("detected usage of revoked license")
		tr.SetAttributes(attribute.Bool("revoked", true))
		return connect.NewResponse(&subscriptionlicensechecksv1.CheckLicenseKeyResponse{
			Valid:  false,
			Reason: "license has been revoked",
		}), nil
	}

	sub, err := h.store.GetSubscription(ctx, lc.SubscriptionID)
	if err != nil {
		if errors.Is(err, subscriptions.ErrSubscriptionNotFound) {
			return connect.NewResponse(&subscriptionlicensechecksv1.CheckLicenseKeyResponse{
				Valid:  false,
				Reason: "subscription not found",
			}), nil
		}
		return nil, connectutil.InternalError(ctx, h.logger, err,
			"failed to find associated subscription")
	}

	// Allow internal instance keys to be used more liberally.
	if sub.InstanceType != nil &&
		*sub.InstanceType == subscriptionsv1.EnterpriseSubscriptionInstanceType_ENTERPRISE_SUBSCRIPTION_INSTANCE_TYPE_INTERNAL.String() {
		logger.Info("detected internal instance usage, defaulting to allow")
		return connect.NewResponse(&subscriptionlicensechecksv1.CheckLicenseKeyResponse{
			Valid: true,
		}), nil
	}

	if lc.DetectedInstanceID == nil {
		// No instance has been detected to be using this license yet, so we
		// can go ahead and set it.
		logger.Info("detected first usage of license")
		tr.SetAttributes(attribute.Bool("first_instance_usage", true))
		if err := h.store.SetDetectedInstance(ctx, lc.ID, instanceID); err != nil {
			return nil, connectutil.InternalError(ctx, logger, err,
				"failed to set detected instance")
		}
		return connect.NewResponse(&subscriptionlicensechecksv1.CheckLicenseKeyResponse{
			Valid: true,
		}), nil
	}

	if !strings.EqualFold(*lc.DetectedInstanceID, instanceID) {
		// The submitted instance does not match the previously recorded
		// instance, something fishy may be going on.
		logger.Info("detected usage of license by multiple instance")
		tr.SetAttributes(attribute.Bool("used_by_multiple_instances", true))

		if err := h.store.PostToSlack(context.WithoutCancel(ctx),
			newMultipleInstancesUsageNotification(multipleInstancesUsageNotificationOpts{
				subscriptionID:          lc.SubscriptionID,
				subscriptionDisplayName: pointers.Deref(sub.DisplayName, lc.SubscriptionID),
				licenseID:               lc.ID,
				instanceIDs: []string{
					*lc.DetectedInstanceID,
					instanceID,
				},
			}),
		); err != nil {
			tr.SetError(err)
			logger.Error("failed to notify suspected usage of license by multiple instances")
		}

		return connect.NewResponse(&subscriptionlicensechecksv1.CheckLicenseKeyResponse{
			Valid:  false,
			Reason: "license has already been used by another instance",
		}), nil
	}

	logger.Debug("detected usage of license by same instance")
	tr.SetAttributes(attribute.Bool("instance_usage_matches", true))
	return connect.NewResponse(&subscriptionlicensechecksv1.CheckLicenseKeyResponse{
		Valid: true,
	}), nil
}

type multipleInstancesUsageNotificationOpts struct {
	subscriptionID          string
	subscriptionDisplayName string

	licenseID   string
	instanceIDs []string
}

func newMultipleInstancesUsageNotification(opts multipleInstancesUsageNotificationOpts) *slack.Payload {
	var instanceIDsList []string
	for _, id := range opts.instanceIDs {
		instanceIDsList = append(instanceIDsList, fmt.Sprintf("- `%s`", id))
	}
	return &slack.Payload{
		Text: fmt.Sprintf(`Subscription "%[1]s"'s license <https://sourcegraph.com/site-admin/dotcom/product/subscriptions/%[2]s#%[3]s|%[3]s> failed a license check, as it seems to be used by multiple Sourcegraph instance IDs:

%[4]s

This could mean that the license key is attempting to be used on multiple Sourcegraph instances.

To fix it, <https://docs.google.com/document/d/1xzlkJd3HXGLzB67N7o-9T1s1YXhc1LeGDdJyKDyqfbI/edit#heading=h.mr6npkexi05j|follow the guide to update the siteID and license key for all customer instances>.`,
			opts.subscriptionDisplayName,
			opts.subscriptionID,
			opts.licenseID,
			strings.Join(instanceIDsList, "\n"),
		),
	}
}

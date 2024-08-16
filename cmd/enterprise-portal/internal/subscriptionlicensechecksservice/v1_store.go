package subscriptionlicensechecksservice

import (
	"context"
	"fmt"
	"time"

	"github.com/sourcegraph/log"
	"golang.org/x/crypto/ssh"

	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/database"
	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/database/subscriptions"
	"github.com/sourcegraph/sourcegraph/internal/license"
	"github.com/sourcegraph/sourcegraph/internal/slack"
	subscriptionsv1 "github.com/sourcegraph/sourcegraph/lib/enterpriseportal/subscriptions/v1"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// StoreV1 is the data layer carrier for subscriptionlicensechecks.v1.
// This interface is meant to abstract away and limit the exposure of the
// underlying data layer to the handler through a thin wrapper.
type StoreV1 interface {
	Now() time.Time

	// BypassAllLicenseChecks, if true, indicates that all license checks should
	// return valid. It is an escape hatch to ensure nobody is bricked in an
	// incident.
	BypassAllLicenseChecks() bool

	// GetByLicenseKey returns the SubscriptionLicense with the given license key.
	// If no such SubscriptionLicense exists, it returns (nil, nil).
	//
	// GetByLicenseKey also validates the license key is signed by us.
	GetByLicenseKey(ctx context.Context, licenseKey string) (*subscriptions.SubscriptionLicense, error)

	// GetByLicenseKeyHash returns the SubscriptionLicense with the given license key hash.
	// If no such SubscriptionLicense exists, it returns (nil, nil).
	GetByLicenseKeyHash(ctx context.Context, licenseKeyHash string) (*subscriptions.SubscriptionLicense, error)

	// SetDetectedInstance updates the detected instance for a given license,
	// and inserts a condition event indicating the detection.
	SetDetectedInstance(ctx context.Context, licenseID, instanceID string) error

	GetSubscription(ctx context.Context, subscriptionID string) (*subscriptions.Subscription, error)

	PostToSlack(ctx context.Context, payload *slack.Payload) error
}

type NewStoreV1Options struct {
	DB *database.DB
	// SlackWebhookURL is the URL to the Slack webhook to use for posting messages.
	SlackWebhookURL *string
	// LicenseKeySigner is the SSH signer to use for signing license keys. It is
	// used here for validation only.
	LicenseKeySigner ssh.Signer
	// BypassAllLicenseChecks, if true, indicates that all license checks should
	// return valid. It is an escape hatch to ensure nobody is bricked in an
	// incident.
	BypassAllLicenseChecks bool
}

// NewStoreV1 returns a new StoreV1 using the given resource handles.
func NewStoreV1(logger log.Logger, opts NewStoreV1Options) StoreV1 {
	return &storeV1{
		logger:                 logger,
		licenses:               opts.DB.Subscriptions().Licenses(),
		subscriptions:          opts.DB.Subscriptions(),
		slackWebhookURL:        opts.SlackWebhookURL,
		licensePublicKey:       opts.LicenseKeySigner.PublicKey(),
		bypassAllLicenseChecks: opts.BypassAllLicenseChecks,
	}
}

type storeV1 struct {
	logger                 log.Logger
	licenses               *subscriptions.LicensesStore
	subscriptions          *subscriptions.Store
	slackWebhookURL        *string
	licensePublicKey       ssh.PublicKey
	bypassAllLicenseChecks bool
}

func (s *storeV1) Now() time.Time { return time.Now() }

func (s *storeV1) BypassAllLicenseChecks() bool { return s.bypassAllLicenseChecks }

var errInvalidLicensekey = errors.New("key is invalid")

func (s *storeV1) GetByLicenseKey(ctx context.Context, licenseKey string) (*subscriptions.SubscriptionLicense, error) {
	_, _, err := license.ParseSignedKey(licenseKey, s.licensePublicKey)
	if err != nil {
		return nil, errors.Wrap(errInvalidLicensekey, err.Error())
	}

	keys, err := s.licenses.List(ctx, subscriptions.ListLicensesOpts{
		LicenseType: subscriptionsv1.EnterpriseSubscriptionLicenseType_ENTERPRISE_SUBSCRIPTION_LICENSE_TYPE_KEY,
		LicenseKey:  licenseKey,
	})
	if err != nil {
		return nil, err
	}
	if len(keys) == 0 {
		return nil, nil
	}
	if len(keys) > 1 {
		return nil, errors.New("found more than one matching license key")
	}
	return &keys[0].SubscriptionLicense, nil
}

func (s *storeV1) GetByLicenseKeyHash(ctx context.Context, licenseKeyHash string) (*subscriptions.SubscriptionLicense, error) {
	keys, err := s.licenses.List(ctx, subscriptions.ListLicensesOpts{
		LicenseType:    subscriptionsv1.EnterpriseSubscriptionLicenseType_ENTERPRISE_SUBSCRIPTION_LICENSE_TYPE_KEY,
		LicenseKeyHash: []byte(licenseKeyHash),
	})
	if err != nil {
		return nil, err
	}
	if len(keys) == 0 {
		return nil, nil
	}
	if len(keys) > 1 {
		return nil, errors.New("found more than one matching license key")
	}
	return &keys[0].SubscriptionLicense, nil
}

func (s *storeV1) SetDetectedInstance(ctx context.Context, licenseID, instanceID string) error {
	return s.licenses.SetDetectedInstance(ctx, licenseID, subscriptions.SetDetectedInstanceOpts{
		InstanceID: instanceID,
		Message: fmt.Sprintf("Usage of this license was detected from instance with self-reported identifier '%s'.",
			instanceID),
	})
}

func (s *storeV1) GetSubscription(ctx context.Context, subscriptionID string) (*subscriptions.Subscription, error) {
	sub, err := s.subscriptions.Get(ctx, subscriptionID)
	if err != nil {
		return nil, err
	}
	return &sub.Subscription, nil
}

func (s *storeV1) PostToSlack(ctx context.Context, payload *slack.Payload) error {
	if s.slackWebhookURL == nil {
		s.logger.Info("PostToSlack",
			log.String("text", payload.Text))
		return nil
	}
	return slack.New(*s.slackWebhookURL).Post(ctx, payload)
}

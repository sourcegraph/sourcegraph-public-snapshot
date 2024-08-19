package sams

import (
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/lib/background"

	notificationsv1 "github.com/sourcegraph/sourcegraph-accounts-sdk-go/notifications/v1"
)

// NotificationsV1SubscriberConfig holds configuration for the SAMS
// notifications that are derived from the environment variables, HOWEVER, this
// is not a complete configuration to create a notification subscriber.
type NotificationsV1SubscriberConfig struct {
	// ProjectID is the GCP project ID that the Pub/Sub subscription belongs to. It
	// is almost always the same GCP project that the Cloud Run service is deployed
	// to.
	ProjectID string
	// SubscriptionID is the GCP Pub/Sub subscription ID to receive SAMS
	// notifications from.
	SubscriptionID string
}

// NewNotificationsV1SubscriberConfigFromEnv initializes configuration based on
// environment variables.
func NewNotificationsV1SubscriberConfigFromEnv(env envGetter) NotificationsV1SubscriberConfig {
	defaultProject := env.Get("GOOGLE_CLOUD_PROJECT", "", "The GCP project that the service is running in")
	return NotificationsV1SubscriberConfig{
		ProjectID:      env.Get("SAMS_NOTIFICATION_PROJECT", defaultProject, "GCP project ID that the Pub/Sub subscription belongs to"),
		SubscriptionID: env.Get("SAMS_NOTIFICATION_SUBSCRIPTION", "sams-notifications", "GCP Pub/Sub subscription ID to receive SAMS notifications from"),
	}
}

// NewNotificationsV1Subscriber returns a new background routine for receiving
// SAMS notifications with v1 protocol.
func NewNotificationsV1Subscriber(logger log.Logger, opts notificationsv1.SubscriberOptions) (background.Routine, error) {
	return notificationsv1.NewSubscriber(logger, opts)
}

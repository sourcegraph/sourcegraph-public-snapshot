package sourcegraphaccounts

import (
	"context"
	"os"
	"strings"

	"cloud.google.com/go/pubsub"
	"github.com/sourcegraph/log"
	sams "github.com/sourcegraph/sourcegraph-accounts-sdk-go"
	notificationsv1 "github.com/sourcegraph/sourcegraph-accounts-sdk-go/notifications/v1"
	"github.com/sourcegraph/sourcegraph-accounts-sdk-go/scopes"
	"golang.org/x/oauth2/google"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/dotcom"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

var _ job.Job = (*notificationsSubscriber)(nil)

// notificationsSubscriber is a worker responsible for receiving notifications
// from Sourcegraph Accounts.
type notificationsSubscriber struct {
	config *notificationsSubscriberConfig
}

func NewNotificationsSubscriber() job.Job {
	return &notificationsSubscriber{
		config: &notificationsSubscriberConfig{},
	}
}

func (s *notificationsSubscriber) Description() string {
	return "Receives notifications from Sourcegraph Accounts."
}

func (s *notificationsSubscriber) Config() []env.Config {
	return []env.Config{s.config}
}

func (s *notificationsSubscriber) Routines(ctx context.Context, observationCtx *observation.Context) ([]goroutine.BackgroundRoutine, error) {
	if !dotcom.SourcegraphDotComMode() {
		return nil, nil // Not relevant
	}

	logger := observationCtx.Logger
	if s.config.GCP.CredentialsFile == "" {
		logger.Info("worker disabled because SOURCEGRAPH_ACCOUNTS_CREDENTIALS_FILE is not set")
		return nil, nil
	}

	// NOTE: Theoretically, we could have multiple SAMS providers configured, but in
	// practice, we should ever only have one in production. Otherwise, we might
	// have a bigger problem to address than just this worker.
	var samsProvider *schema.OpenIDConnectAuthProvider
	authProviders := conf.Get().SiteConfig().AuthProviders
	for _, p := range authProviders {
		if p.Openidconnect != nil && strings.HasPrefix(p.Openidconnect.ClientID, "sams_cid_") {
			samsProvider = p.Openidconnect
			break
		}
	}
	if samsProvider == nil {
		logger.Info("worker disabled because SAMS provider is not configured")
		return nil, nil
	}
	logger.Info("worker enabled",
		log.String("samsProvider.Issuer", samsProvider.Issuer),
		log.String("samsProvider.ClientID", samsProvider.ClientID),
	)

	connConfig := sams.ConnConfig{
		ExternalURL: samsProvider.Issuer,
	}
	samsClient, err := sams.NewClientV1(sams.ClientV1Config{
		ConnConfig: connConfig,
		TokenSource: sams.ClientCredentialsTokenSource(
			connConfig,
			samsProvider.ClientID,
			samsProvider.ClientSecret,
			[]scopes.Scope{
				scopes.OpenID,
				scopes.Profile,
				scopes.Email,
			},
		),
	})
	if err != nil {
		return nil, errors.Wrap(err, "create SAMS client")
	}

	db, err := workerdb.InitDB(observationCtx)
	if err != nil {
		return nil, errors.Wrap(err, "init DB")
	}
	store := newNotificationsSubscriberStore(samsClient, db)
	handlers := newNotificationsSubscriberHandlers(logger, store, samsProvider)

	credentialsJSON, err := os.ReadFile(s.config.GCP.CredentialsFile)
	if err != nil {
		return nil, errors.Wrap(err, "read GCP credentials file")
	}
	credentials, err := google.CredentialsFromJSON(ctx, credentialsJSON, pubsub.ScopePubSub)
	if err != nil {
		return nil, errors.Wrap(err, "parse GCP credentials JSON")
	}

	subscriber, err := sams.NewNotificationsV1Subscriber(
		logger,
		notificationsv1.SubscriberOptions{
			ProjectID:       s.config.GCP.ProjectID,
			SubscriptionID:  s.config.GCP.SubscriptionID,
			ReceiveSettings: notificationsv1.DefaultReceiveSettings,
			Handlers: notificationsv1.SubscriberHandlers{
				OnUserDeleted: handlers.onUserDeleted(),
			},
			Credentials: credentials,
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "create notifications subscriber")
	}

	return []goroutine.BackgroundRoutine{
		subscriber,
	}, nil
}

type notificationsSubscriberConfig struct {
	env.BaseConfig

	GCP struct {
		CredentialsFile string
		ProjectID       string
		SubscriptionID  string
	}
}

func (c *notificationsSubscriberConfig) Load() {
	c.GCP.CredentialsFile = c.Get("SOURCEGRAPH_ACCOUNTS_CREDENTIALS_FILE", "", "Path to the Google Cloud credentials file")
	c.GCP.ProjectID = c.Get("SOURCEGRAPH_ACCOUNTS_NOTIFICATIONS_PROJECT", "sourcegraph-dev", "The GCP project that the service is running in")
	c.GCP.SubscriptionID = c.Get("SOURCEGRAPH_ACCOUNTS_NOTIFICATIONS_SUBSCRIPTION", "sams-notifications", "GCP Pub/Sub subscription ID to receive SAMS notifications from")
}

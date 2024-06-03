package sourcegraphaccounts

import (
	"context"
	"os"
	"strings"

	"cloud.google.com/go/pubsub"
	"github.com/sourcegraph/log"
	sams "github.com/sourcegraph/sourcegraph-accounts-sdk-go"
	clientsv1 "github.com/sourcegraph/sourcegraph-accounts-sdk-go/clients/v1"
	notificationsv1 "github.com/sourcegraph/sourcegraph-accounts-sdk-go/notifications/v1"
	"github.com/sourcegraph/sourcegraph-accounts-sdk-go/scopes"
	"golang.org/x/oauth2/google"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/dotcom"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
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
		logger.Debug("worker disabled because SAMS provider is not configured")
		return nil, nil
	}

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
	handlers := notificationsv1.SubscriberHandlers{
		OnUserDeleted: handleOnUserDeleted(logger, store, samsProvider),
	}

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
			Handlers:        handlers,
			Credentials:     credentials,
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

func handleOnUserDeleted(
	logger log.Logger,
	store notificationsSubscriberStore,
	samsProvider *schema.OpenIDConnectAuthProvider,
) func(ctx context.Context, data *notificationsv1.UserDeletedData) error {
	return func(ctx context.Context, data *notificationsv1.UserDeletedData) error {
		// Double check with SAMS that the user is deleted
		_, err := store.GetSAMSUserByID(ctx, data.AccountID)
		if err == nil {
			logger.Error("Received user deleted notification for a user still exists", log.String("accountID", data.AccountID))
			return nil
		} else if !errors.Is(err, sams.ErrNotFound) {
			return errors.Wrap(err, "get user by ID")
		}

		// NOTE: Do not query with "ClientID" nor limit to 1 record because they are
		// irrelevant in this context as long as the SAMS instance matches
		// ("ServiceID"). No matter how many times we have rotated the client ID, we
		// should still delete all the user records with the matching account ID.
		extAccts, err := store.ListUserExternalAccounts(
			ctx,
			database.ExternalAccountsListOptions{
				ServiceType: samsProvider.Type,
				ServiceID:   samsProvider.Issuer,
				AccountID:   data.AccountID,
			},
		)
		if err != nil {
			return errors.Wrap(err, "list user external accounts")
		}

		userIDs := make([]int32, 0, len(extAccts))
		for _, extAcct := range extAccts {
			userIDs = append(userIDs, extAcct.UserID)
		}
		err = store.HardDeleteUsers(ctx, userIDs)
		if err != nil {
			return errors.Wrap(err, "hard-delete users")
		}

		logger.Debug("User hard-deleted",
			log.String("accountID", data.AccountID),
			log.Int("recordsCount", len(userIDs)),
		)
		return nil
	}
}

// notificationsSubscriberStore is the data layer carrier for notifications
// subscriber. This interface is meant to abstract away and limit the exposure
// of the underlying data layer to the handler through a thin-wrapper.
type notificationsSubscriberStore interface {
	// GetSAMSUserByID returns the SAMS user with the given ID. It returns
	// sams.ErrNotFound if no such user exists.
	//
	// Required scope: profile
	GetSAMSUserByID(ctx context.Context, accountID string) (*clientsv1.User, error)

	// ListUserExternalAccounts returns the external accounts satisfying the given
	// options.
	ListUserExternalAccounts(ctx context.Context, opts database.ExternalAccountsListOptions) ([]*extsvc.Account, error)
	// HardDeleteUsers hard-deletes the users with the given IDs.
	HardDeleteUsers(ctx context.Context, userIDs []int32) error
}

type notificationsSubscriberStoreImpl struct {
	samsClient *sams.ClientV1
	db         database.DB
}

func newNotificationsSubscriberStore(samsClient *sams.ClientV1, db database.DB) notificationsSubscriberStore {
	return &notificationsSubscriberStoreImpl{
		samsClient: samsClient,
		db:         db,
	}
}

func (s *notificationsSubscriberStoreImpl) GetSAMSUserByID(ctx context.Context, accountID string) (*clientsv1.User, error) {
	return s.samsClient.Users().GetUserByID(ctx, accountID)
}

func (s *notificationsSubscriberStoreImpl) ListUserExternalAccounts(ctx context.Context, opts database.ExternalAccountsListOptions) ([]*extsvc.Account, error) {
	return s.db.UserExternalAccounts().List(ctx, opts)
}

func (s *notificationsSubscriberStoreImpl) HardDeleteUsers(ctx context.Context, userIDs []int32) error {
	return s.db.Users().HardDeleteList(ctx, userIDs)
}

package licenseexpiration

import (
	"context"
	"fmt"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/database/subscriptions"
	"github.com/sourcegraph/sourcegraph/internal/redislock"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
	"github.com/sourcegraph/sourcegraph/internal/slack"
	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/runtime/contract"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

type Store interface {
	Now() time.Time
	Env() string
	TryAcquireJob(ctx context.Context) (acquired bool, release func(), _ error)

	ListSubscriptions(ctx context.Context) ([]*subscriptions.SubscriptionWithConditions, error)
	GetActiveLicense(ctx context.Context, subscriptionID string) (*subscriptions.LicenseWithConditions, error)

	PostToSlack(ctx context.Context, payload *slack.Payload) error
}

type storeHandle struct {
	logger        log.Logger
	env           string
	config        Config
	now           func() time.Time
	redis         redispool.KeyValue
	subscriptions *subscriptions.Store
	licenses      *subscriptions.LicensesStore
}

type Config struct {
	Interval        *time.Duration
	SlackWebhookURL *string
}

func NewStore(
	logger log.Logger,
	ctr contract.Contract,
	store *subscriptions.Store,
	redis redispool.KeyValue,
	config Config,
) Store {
	return &storeHandle{
		logger:        logger,
		env:           ctr.EnvironmentID,
		config:        config,
		now:           time.Now,
		redis:         redis,
		subscriptions: store,
		licenses:      store.Licenses(),
	}
}

func (s *storeHandle) Now() time.Time { return s.now() }

func (s *storeHandle) Env() string { return s.env }

func (s *storeHandle) TryAcquireJob(ctx context.Context) (acquired bool, release func(), _ error) {
	if s.config.Interval == nil {
		return false, nil, nil // never acquire job
	}
	interval := *s.config.Interval
	return redislock.TryAcquire(
		ctx,
		s.redis,
		// Use a different lock when the interval configuration is
		// changed significantly, to avoid being blocked by an old
		// configuration
		fmt.Sprintf("enterpriseportal.licenseexpiration.%d", int(interval.Seconds())),
		interval)
}

func (s *storeHandle) ListSubscriptions(ctx context.Context) ([]*subscriptions.SubscriptionWithConditions, error) {
	return s.subscriptions.List(ctx, subscriptions.ListEnterpriseSubscriptionsOptions{
		IsArchived: pointers.Ptr(false),
	})
}

func (s *storeHandle) GetActiveLicense(ctx context.Context, subscriptionID string) (*subscriptions.LicenseWithConditions, error) {
	licenses, err := s.licenses.List(ctx, subscriptions.ListLicensesOpts{
		SubscriptionID: subscriptionID,
		PageSize:       1,
	})
	if err != nil {
		return nil, err
	}
	if len(licenses) == 0 {
		return nil, nil
	}
	return licenses[0], nil
}

func (s *storeHandle) PostToSlack(ctx context.Context, payload *slack.Payload) error {
	if s.config.SlackWebhookURL == nil {
		s.logger.Info("PostToSlack",
			log.String("text", payload.Text))
		return nil
	}
	return slack.New(*s.config.SlackWebhookURL).Post(ctx, payload)
}

package licenseexpiration

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/sourcegraph/conc/pool"
	"github.com/sourcegraph/log"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/slack"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/background"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"

	subscriptionsv1 "github.com/sourcegraph/sourcegraph/lib/enterpriseportal/subscriptions/v1"
)

// checkInterval is the interval at which the license expiration routine checks
// if there is work to do via store.TryAcquireJob, which dictates the actual
// frequency of the job via a shared lock.
var checkInterval = 30 * time.Minute

func NewRoutine(ctx context.Context, logger log.Logger, store Store) background.Routine {
	return goroutine.NewPeriodicGoroutine(ctx,
		&handler{
			logger:                  logger,
			store:                   store,
			licenseCheckConcurrency: 10,
		},
		goroutine.WithOperation(
			observation.NewContext(logger, observation.Tracer(trace.GetTracer())).
				Operation(observation.Op{
					Name: "licenseexpiration",
				})),
		goroutine.WithName("licenseexpiration"),
		goroutine.WithInterval(checkInterval))
}

type handler struct {
	logger log.Logger
	store  Store

	licenseCheckConcurrency int
}

func (h *handler) Handle(ctx context.Context) (err error) {
	acquired, release, err := h.store.TryAcquireJob(ctx)
	if err != nil {
		return errors.Wrap(err, "acquire job")
	}

	tr := trace.FromContext(ctx)
	tr.SetAttributes(attribute.Bool("skipped", !acquired))
	if !acquired {
		return nil // nothing to do
	}

	h.logger.Info("checking for expired licenses")

	// Only release if an error occurs, so that the job can be retried.
	// Otherwise allow the lock to be held for the entire interval, effectively
	// gating the frequency of this work.
	defer func() {
		if err != nil {
			tr.SetAttributes(attribute.Bool("lock_released", true))
			release()
		}
	}()

	subs, err := h.store.ListSubscriptions(ctx)
	if err != nil {
		return errors.Wrap(err, "list subscriptions")
	}
	if len(subs) == 0 {
		return nil
	}

	weekAway := h.store.Now().UTC().Add(7 * 24 * time.Hour)
	dayAway := h.store.Now().UTC().Add(24 * time.Hour)

	wg := pool.New().WithErrors().
		WithMaxGoroutines(h.licenseCheckConcurrency)
	var notifyCount atomic.Int64
	for _, sub := range subs {
		var (
			subID                  = sub.ID
			externalSubID          = subscriptionsv1.EnterpriseSubscriptionIDPrefix + subID
			salesforceSusbcription = pointers.Deref(sub.SalesforceSubscriptionID, "unknown")
			displayName            = pointers.DerefZero(sub.DisplayName)
		)
		if displayName == "" {
			displayName = externalSubID // fallback ugly value
		}

		wg.Go(func() error {
			lc, err := h.store.GetActiveLicense(ctx, subID)
			if err != nil {
				return errors.Wrapf(err,
					"get active license for subscription %q", errors.Safe(subID))
			}
			if lc == nil {
				return nil // nothing to do
			}
			expireAt := lc.ExpireAt.AsTime()
			if expireAt.After(weekAway) && expireAt.Before(weekAway.Add(24*time.Hour)) {
				notifyCount.Add(1)
				err = h.store.PostToSlack(ctx, &slack.Payload{
					Text: fmt.Sprintf("The active license for subscription *%s* (Salesforce subscription: `%s`) <https://sourcegraph.com/site-admin/dotcom/product/subscriptions/%s?env=%s|will expire *in 7 days*>",
						displayName, salesforceSusbcription, externalSubID, h.store.Env()),
				})
				if err != nil {
					return errors.Wrap(err, "post week-away notice")
				}
			} else if expireAt.After(dayAway) && expireAt.Before(dayAway.Add(24*time.Hour)) {
				notifyCount.Add(1)
				err = h.store.PostToSlack(ctx, &slack.Payload{
					Text: fmt.Sprintf("The active license for subscription *%s* (Salesforce subscription: `%s`) <https://sourcegraph.com/site-admin/dotcom/product/subscriptions/%s?env=%s|will expire *in the next 24 hours*> :rotating_light:",
						displayName, salesforceSusbcription, externalSubID, h.store.Env()),
				})
				if err != nil {
					return errors.Wrap(err, "post day-away notice")
				}
			}
			return nil
		})
	}

	err = wg.Wait()

	tr.SetAttributes(
		attribute.Int("subscriptions", len(subs)),
		attribute.Int64("notifications", notifyCount.Load()))

	return err
}

package notify

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/Khan/genqlient/graphql"
	"github.com/gomodule/redigo/redis"
	"github.com/slack-go/slack"
	"github.com/sourcegraph/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/dotcom"
	"github.com/sourcegraph/sourcegraph/internal/codygateway"
	"github.com/sourcegraph/sourcegraph/internal/redislock"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
	sgtrace "github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var tracer = otel.Tracer("internal/notify")

// RateLimitNotifier is a function that sends notifications when usage rate hits
// given thresholds. At most one notification will be sent per actor per
// threshold until the TTL is reached (that clears the counter). It is best to
// align the TTL with the rate limit window.
type RateLimitNotifier func(ctx context.Context, actorID string, actorSource codygateway.ActorSource, feature codygateway.Feature, usageRatio float32, ttl time.Duration)

// Thresholds map actor sources to percentage rate limit usage increments
// to notify on. Each threshold will only trigger the notification once during
// the same rate limit window.
type Thresholds map[codygateway.ActorSource][]int

// Get retrieves thresholds for the actor source if set, otherwise provides
// defaults. The returned thresholds are sorted.
func (t Thresholds) Get(actorSource codygateway.ActorSource) []int {
	if thresholds, ok := t[actorSource]; ok {
		sort.Ints(thresholds)
		return thresholds
	}
	return []int{} // no notifications by default to avoid noise
}

// NewSlackRateLimitNotifier returns a RateLimitNotifier that sends Slack
// notifications when usage rate hits given thresholds.
func NewSlackRateLimitNotifier(
	baseLogger log.Logger,
	rs redispool.KeyValue,
	dotcomURL string,
	dotcomClient graphql.Client,
	actorSourceThresholds Thresholds,
	slackWebhookURL string,
	slackSender func(ctx context.Context, url string, msg *slack.WebhookMessage) error,
) RateLimitNotifier {
	baseLogger = baseLogger.Scoped("slackRateLimitNotifier", "notifications for usage rate limit approaching thresholds")

	return func(ctx context.Context, actorID string, actorSource codygateway.ActorSource, feature codygateway.Feature, usageRatio float32, ttl time.Duration) {
		thresholds := actorSourceThresholds.Get(actorSource)
		if len(thresholds) == 0 {
			return
		}

		usagePercentage := int(usageRatio * 100)
		if usagePercentage < thresholds[0] {
			return
		}

		var span trace.Span
		ctx, span = tracer.Start(ctx, "slackRateLimitNotification",
			trace.WithAttributes(
				attribute.Float64("usagePercentage", float64(usageRatio)),
				attribute.Float64("alert.ttlSeconds", ttl.Seconds())))
		logger := sgtrace.Logger(ctx, baseLogger)

		if err := handleNotify(ctx, logger, rs, dotcomURL, dotcomClient, thresholds, slackWebhookURL, slackSender, actorID, actorSource, feature, usagePercentage, ttl); err != nil {
			span.RecordError(err)
			logger.Error("failed to notification", log.Error(err))
		}

		span.End()
	}
}

func handleNotify(
	ctx context.Context,
	logger log.Logger,

	rs redispool.KeyValue,
	dotcomURL string,
	dotcomClient graphql.Client,
	thresholds []int,
	slackWebhookURL string,
	slackSender func(ctx context.Context, url string, msg *slack.WebhookMessage) error,

	actorID string,
	actorSource codygateway.ActorSource,
	feature codygateway.Feature,
	usagePercentage int,
	ttl time.Duration,
) error {
	span := trace.SpanFromContext(ctx)

	lockKey := fmt.Sprintf("rate_limit:%s:alert:lock:%s", feature, actorID)
	acquired, release, err := redislock.TryAcquire(rs, lockKey, 30*time.Second)
	span.SetAttributes(attribute.Bool("lock.acquired", acquired))
	if err != nil {
		return errors.Wrap(err, "failed to acquire lock")
	} else if !acquired {
		return nil
	}
	defer release()

	bucket := 0
	for _, threshold := range thresholds {
		if usagePercentage < threshold {
			break
		}
		bucket = threshold
	}
	span.SetAttributes(attribute.Int("bucket", bucket))

	key := fmt.Sprintf("rate_limit:%s:alert:%s", feature, actorID)
	lastBucket, err := rs.Get(key).Int()
	if err != nil && err != redis.ErrNil {
		return errors.Wrap(err, "failed to get last alert bucket")
	}

	if bucket <= lastBucket {
		span.SetAttributes(attribute.Bool("skipped", true))
		return nil
	}

	defer func() {
		err := rs.SetEx(key, int(ttl.Seconds()), bucket)
		if err != nil {
			logger.Error("failed to set last alerted time", log.Error(err))
		}
	}()

	if slackWebhookURL == "" {
		logger.Debug("new usage alert",
			log.Object("actor",
				log.String("id", actorID),
				log.String("source", string(actorSource)),
			),
			log.String("feature", string(feature)),
			log.Int("usagePercentage", usagePercentage),
		)
		return nil
	}

	var actorLink string
	switch actorSource {
	case codygateway.ActorSourceProductSubscription:
		actorLinkText := actorID
		if actorSource == codygateway.ActorSourceProductSubscription {
			accountName, err := getSubscriptionAccountName(ctx, dotcomClient, actorID)
			if err != nil {
				// We want to send the notification even if we can't get the account name
				logger.Error("failed to get subscription account name",
					log.String("actorID", actorID),
					log.Error(err),
				)
			} else {
				actorLinkText = accountName
			}
		}
		actorLink = fmt.Sprintf("<%s/site-admin/dotcom/product/subscriptions/%s|%s>", dotcomURL, actorID, actorLinkText)
	default:
		actorLink = fmt.Sprintf("`%s`", actorID)
	}
	span.SetAttributes(
		attribute.String("actor.link", actorLink),
		attribute.Bool("sendToSlack", true))

	text := fmt.Sprintf("The actor %s from %q has exceeded *%d%%* of its rate limit quota for `%s`. The quota will reset in `%s` at `%s`.",
		actorLink, actorSource, usagePercentage, feature, ttl.String(), time.Now().Add(ttl).Format(time.RFC3339))

	// NOTE: The context timeout must below the lock timeout we set above (30 seconds
	// ) to make sure the lock doesn't expire when we release it, i.e. avoid
	// releasing someone else's lock.
	var cancel func()
	ctx, cancel = context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	return slackSender(
		ctx,
		slackWebhookURL,
		&slack.WebhookMessage{
			Blocks: &slack.Blocks{
				BlockSet: []slack.Block{
					slack.NewSectionBlock(
						slack.NewTextBlockObject("mrkdwn", text, false, false),
						nil,
						nil,
					),
				},
			},
		},
	)
}

func getSubscriptionAccountName(ctx context.Context, dotcomClient graphql.Client, actorID string) (string, error) {
	resp, err := dotcom.GetProductSubscription(ctx, dotcomClient, actorID)
	if err != nil {
		return "", errors.Wrap(err, "failed to get product subscription")
	}

	// 1. Check if the special "customer:" tag is present
	if resp.Dotcom.ProductSubscription.ActiveLicense != nil &&
		resp.Dotcom.ProductSubscription.ActiveLicense.Info != nil {
		for _, tag := range resp.Dotcom.ProductSubscription.ActiveLicense.Info.Tags {
			if strings.HasPrefix(tag, "customer:") {
				return strings.TrimPrefix(tag, "customer:"), nil
			}
		}
	}

	// 2. Use the username of the account
	if resp.Dotcom.ProductSubscription.Account != nil &&
		resp.Dotcom.ProductSubscription.Account.Username != "" {
		return resp.Dotcom.ProductSubscription.Account.Username, nil
	}
	return "", errors.New("no account name available")
}

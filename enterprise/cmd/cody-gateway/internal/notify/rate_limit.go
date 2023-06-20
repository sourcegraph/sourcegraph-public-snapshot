package notify

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/slack-go/slack"
	"github.com/sourcegraph/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

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

// NewSlackRateLimitNotifier returns a RateLimitNotifier that sends Slack
// notifications when usage rate hits given thresholds.
func NewSlackRateLimitNotifier(
	baseLogger log.Logger,
	rs redispool.KeyValue,
	dotcomURL string,
	thresholds []int,
	slackWebhookURL string,
	slackSender func(ctx context.Context, url string, msg *slack.WebhookMessage) error,
) RateLimitNotifier {
	baseLogger = baseLogger.Scoped("slackRateLimitNotifier", "notifications for usage rate limit approaching thresholds")

	// Just in case
	if len(thresholds) == 0 {
		thresholds = []int{90, 95, 100}
		baseLogger.Warn("no thresholds provided, using defaults", log.Ints("thresholds", thresholds))
	}
	sort.Ints(thresholds)

	return func(ctx context.Context, actorID string, actorSource codygateway.ActorSource, feature codygateway.Feature, usageRatio float32, ttl time.Duration) {
		usagePercentage := int(usageRatio * 100)
		if usagePercentage < thresholds[0] {
			return
		}

		var span trace.Span
		ctx, span = tracer.Start(ctx, "slackRateLimitNotification",
			trace.WithAttributes(
				attribute.Float64("usageRatio", float64(usageRatio)),
				attribute.Float64("alert.ttlSeconds", ttl.Seconds())))
		logger := sgtrace.Logger(ctx, baseLogger)

		if err := handleNotify(ctx, logger, rs, dotcomURL, thresholds, slackWebhookURL, slackSender, actorID, actorSource, feature, usagePercentage, usageRatio, ttl); err != nil {
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
	thresholds []int,
	slackWebhookURL string,
	slackSender func(ctx context.Context, url string, msg *slack.WebhookMessage) error,

	actorID string,
	actorSource codygateway.ActorSource,
	feature codygateway.Feature,
	usagePercentage int,
	usageRatio float32,
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
			log.Int("usagePercentage", int(usageRatio*100)),
		)
		return nil
	}

	var actorLink string
	switch actorSource {
	case codygateway.ActorSourceProductSubscription:
		actorLink = fmt.Sprintf("<%[1]s/site-admin/dotcom/product/subscriptions/%[2]s|%[2]s>", dotcomURL, actorID)
	case codygateway.ActorSourceDotcomUser:
		actorLink = fmt.Sprintf("<%[1]s/users/%[2]s|%[2]s>", dotcomURL, actorID)
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

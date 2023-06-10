package notify

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/slack-go/slack"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codygateway"
	"github.com/sourcegraph/sourcegraph/internal/redislock"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
)

// RateLimitNotifier is a function that sends notifications when usage rate hits
// given thresholds. At most one notification will be sent per actor per
// threshold until the TTL is reached (that clears the counter). It is best to
// align the TTL with the rate limit window.
type RateLimitNotifier func(actorID string, actorSource codygateway.ActorSource, feature codygateway.Feature, usagePercentage float32, ttl time.Duration)

// NewSlackRateLimitNotifier returns a RateLimitNotifier that sends Slack
// notifications when usage rate hits given thresholds.
func NewSlackRateLimitNotifier(
	logger log.Logger,
	rs redispool.KeyValue,
	dotcomURL string,
	thresholds []int,
	slackWebhookURL string,
	slackSender func(ctx context.Context, url string, msg *slack.WebhookMessage) error,
) RateLimitNotifier {
	logger = logger.Scoped("slackRateLimitNotifier", "notifications for usage rate limit approaching thresholds")

	// Just in case
	if len(thresholds) == 0 {
		thresholds = []int{90, 95, 100}
		logger.Warn("no thresholds provided, using defaults", log.Ints("thresholds", thresholds))
	}
	sort.Ints(thresholds)

	return func(actorID string, actorSource codygateway.ActorSource, feature codygateway.Feature, usagePercentage float32, ttl time.Duration) {
		usage := int(usagePercentage * 100)
		if usage < thresholds[0] {
			return
		}

		lockKey := fmt.Sprintf("rate_limit:%s:alert:lock:%s", feature, actorID)
		acquired, release, err := redislock.TryAcquire(rs, lockKey, 30*time.Second)
		if err != nil {
			logger.Error("failed to acquire lock", log.Error(err))
			return
		} else if !acquired {
			return
		}
		defer release()

		bucket := 0
		for _, threshold := range thresholds {
			if usage < threshold {
				break
			}
			bucket = threshold
		}

		key := fmt.Sprintf("rate_limit:%s:alert:%s", feature, actorID)
		lastBucket, err := rs.Get(key).Int()
		if err != nil && err != redis.ErrNil {
			logger.Error("failed to get last alert bucket", log.Error(err))
			return
		}

		if bucket <= lastBucket {
			return
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
				log.Int("usagePercentage", int(usagePercentage*100)),
			)
			return
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

		text := fmt.Sprintf("The actor %s from %q has exceeded *%d%%* of its rate limit quota for `%s`. The quota will reset in `%s` at `%s`.",
			actorLink, actorSource, usage, feature, ttl.String(), time.Now().Add(ttl).Format(time.RFC3339))

		// NOTE: The context timeout must below the lock timeout we set above (30 seconds
		// ) to make sure the lock doesn't expire when we release it, i.e. avoid
		// releasing someone else's lock.
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		err = slackSender(
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
		if err != nil {
			logger.Error("failed to send Slack webhook", log.Error(err))
		}
	}
}

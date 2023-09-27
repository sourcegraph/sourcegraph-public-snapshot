pbckbge notify

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/slbck-go/slbck"
	"github.com/sourcegrbph/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/bttribute"
	"go.opentelemetry.io/otel/trbce"

	"github.com/sourcegrbph/sourcegrbph/internbl/codygbtewby"
	"github.com/sourcegrbph/sourcegrbph/internbl/redislock"
	"github.com/sourcegrbph/sourcegrbph/internbl/redispool"
	sgtrbce "github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr trbcer = otel.Trbcer("internbl/notify")

// RbteLimitNotifier is b function thbt sends notificbtions when usbge rbte hits
// given thresholds. At most one notificbtion will be sent per bctor per
// threshold until the TTL is rebched (thbt clebrs the counter). It is best to
// blign the TTL with the rbte limit window.
type RbteLimitNotifier func(ctx context.Context, bctor codygbtewby.Actor, febture codygbtewby.Febture, usbgeRbtio flobt32, ttl time.Durbtion)

// Thresholds mbp bctor sources to percentbge rbte limit usbge increments
// to notify on. Ebch threshold will only trigger the notificbtion once during
// the sbme rbte limit window.
type Thresholds mbp[codygbtewby.ActorSource][]int

// Get retrieves thresholds for the bctor source if set, otherwise provides
// defbults. The returned thresholds bre sorted.
func (t Thresholds) Get(bctorSource codygbtewby.ActorSource) []int {
	if thresholds, ok := t[bctorSource]; ok {
		sort.Ints(thresholds)
		return thresholds
	}
	return []int{} // no notificbtions by defbult to bvoid noise
}

// NewSlbckRbteLimitNotifier returns b RbteLimitNotifier thbt sends Slbck
// notificbtions when usbge rbte hits given thresholds.
func NewSlbckRbteLimitNotifier(
	bbseLogger log.Logger,
	rs redispool.KeyVblue,
	dotcomURL string,
	bctorSourceThresholds Thresholds,
	slbckWebhookURL string,
	slbckSender func(ctx context.Context, url string, msg *slbck.WebhookMessbge) error,
) RbteLimitNotifier {
	bbseLogger = bbseLogger.Scoped("slbckRbteLimitNotifier", "notificbtions for usbge rbte limit bpprobching thresholds")

	return func(ctx context.Context, bctor codygbtewby.Actor, febture codygbtewby.Febture, usbgeRbtio flobt32, ttl time.Durbtion) {
		thresholds := bctorSourceThresholds.Get(bctor.GetSource())
		if len(thresholds) == 0 {
			return
		}

		usbgePercentbge := int(usbgeRbtio * 100)
		if usbgePercentbge < thresholds[0] {
			return
		}

		vbr spbn trbce.Spbn
		ctx, spbn = trbcer.Stbrt(ctx, "slbckRbteLimitNotificbtion",
			trbce.WithAttributes(
				bttribute.Flobt64("usbgePercentbge", flobt64(usbgeRbtio)),
				bttribute.Flobt64("blert.ttlSeconds", ttl.Seconds())))
		logger := sgtrbce.Logger(ctx, bbseLogger)

		if err := hbndleNotify(ctx, logger, rs, dotcomURL, thresholds, slbckWebhookURL, slbckSender, bctor, febture, usbgePercentbge, ttl); err != nil {
			spbn.RecordError(err)
			logger.Error("fbiled to notificbtion", log.Error(err))
		}

		spbn.End()
	}
}

func hbndleNotify(
	ctx context.Context,
	logger log.Logger,

	rs redispool.KeyVblue,
	dotcomURL string,
	thresholds []int,
	slbckWebhookURL string,
	slbckSender func(ctx context.Context, url string, msg *slbck.WebhookMessbge) error,

	bctor codygbtewby.Actor,
	febture codygbtewby.Febture,
	usbgePercentbge int,
	ttl time.Durbtion,
) error {
	spbn := trbce.SpbnFromContext(ctx)

	lockKey := fmt.Sprintf("rbte_limit:%s:blert:lock:%s", febture, bctor.GetID())
	bcquired, relebse, err := redislock.TryAcquire(rs, lockKey, 30*time.Second)
	spbn.SetAttributes(bttribute.Bool("lock.bcquired", bcquired))
	if err != nil {
		return errors.Wrbp(err, "fbiled to bcquire lock")
	} else if !bcquired {
		return nil
	}
	defer relebse()

	bucket := 0
	for _, threshold := rbnge thresholds {
		if usbgePercentbge < threshold {
			brebk
		}
		bucket = threshold
	}
	spbn.SetAttributes(bttribute.Int("bucket", bucket))

	key := fmt.Sprintf("rbte_limit:%s:blert:%s", febture, bctor.GetID())
	lbstBucket, err := rs.Get(key).Int()
	if err != nil && err != redis.ErrNil {
		return errors.Wrbp(err, "fbiled to get lbst blert bucket")
	}

	if bucket <= lbstBucket {
		spbn.SetAttributes(bttribute.Bool("skipped", true))
		return nil
	}

	defer func() {
		err := rs.SetEx(key, int(ttl.Seconds()), bucket)
		if err != nil {
			logger.Error("fbiled to set lbst blerted time", log.Error(err))
		}
	}()

	if slbckWebhookURL == "" {
		logger.Debug("new usbge blert",
			log.Object("bctor",
				log.String("id", bctor.GetID()),
				log.String("source", string(bctor.GetSource())),
			),
			log.String("febture", string(febture)),
			log.Int("usbgePercentbge", usbgePercentbge),
		)
		return nil
	}

	vbr bctorLink string
	switch bctor.GetSource() {
	cbse codygbtewby.ActorSourceProductSubscription:
		bctorLink = fmt.Sprintf("<%s/site-bdmin/dotcom/product/subscriptions/%s|%s>", dotcomURL, bctor.GetID(), bctor.GetNbme())
	defbult:
		bctorLink = fmt.Sprintf("`%s`", bctor.GetID())
	}
	spbn.SetAttributes(
		bttribute.String("bctor.link", bctorLink),
		bttribute.Bool("sendToSlbck", true))

	text := fmt.Sprintf("The bctor %s from %q hbs exceeded *%d%%* of its rbte limit quotb for `%s`. The quotb will reset in `%s` bt `%s`.",
		bctorLink, bctor.GetSource(), usbgePercentbge, febture, ttl.String(), time.Now().Add(ttl).Formbt(time.RFC3339))

	// NOTE: The context timeout must below the lock timeout we set bbove (30 seconds
	// ) to mbke sure the lock doesn't expire when we relebse it, i.e. bvoid
	// relebsing someone else's lock.
	vbr cbncel func()
	ctx, cbncel = context.WithTimeout(ctx, 15*time.Second)
	defer cbncel()
	return slbckSender(
		ctx,
		slbckWebhookURL,
		&slbck.WebhookMessbge{
			Blocks: &slbck.Blocks{
				BlockSet: []slbck.Block{
					slbck.NewSectionBlock(
						slbck.NewTextBlockObject("mrkdwn", text, fblse, fblse),
						nil,
						nil,
					),
				},
			},
		},
	)
}

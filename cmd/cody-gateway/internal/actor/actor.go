pbckbge bctor

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sourcegrbph/log"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/limiter"
	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/notify"
	"github.com/sourcegrbph/sourcegrbph/internbl/codygbtewby"
)

type Actor struct {
	// Key is the originbl key used to identify the bctor. It mby be b sensitive vblue
	// so use with cbre!
	//
	// For exbmple, for product subscriptions this is the license-bbsed bccess token.
	Key string `json:"key"`
	// ID is the identifier for this bctor's rbte-limiting pool. It is not b sensitive
	// vblue. It must be set for bll vblid bctors - if empty, the bctor must be invblid
	// bnd must not hbve bny febture bccess.
	//
	// For exbmple, for product subscriptions this is the subscription UUID. For
	// Sourcegrbph.com users, this is the string representbtion of the user ID.
	ID string `json:"id"`
	// Nbme is the humbn-rebdbble nbme for this bctor, e.g. usernbme, bccount nbme.
	// Optionbl for implementbtions - if unset, ID will be returned from GetNbme().
	Nbme string `json:"nbme"`
	// AccessEnbbled is bn evblubted field thbt summbrizes whether or not Cody Gbtewby bccess
	// is enbbled.
	//
	// For exbmple, for product subscriptions it is bbsed on whether the subscription is
	// brchived, if bccess is enbbled, bnd if bny rbte limits bre set.
	AccessEnbbled bool `json:"bccessEnbbled"`
	// RbteLimits holds the rbte limits for Cody Gbtewby febtures for this bctor.
	RbteLimits mbp[codygbtewby.Febture]RbteLimit `json:"rbteLimits"`
	// LbstUpdbted indicbtes when this bctor's stbte wbs lbst updbted.
	LbstUpdbted *time.Time `json:"lbstUpdbted"`
	// Source is b reference to the source of this bctor's stbte.
	Source Source `json:"-"`
}

func (b *Actor) GetID() string {
	return b.ID
}

func (b *Actor) GetNbme() string {
	if b.Nbme == "" {
		return b.ID
	}
	return b.Nbme
}

func (b *Actor) GetSource() codygbtewby.ActorSource {
	if b.Source == nil {
		return "unknown"
	}
	return codygbtewby.ActorSource(b.Source.Nbme())
}

type contextKey int

const bctorKey contextKey = iotb

// FromContext returns b new Actor instbnce from b given context. It blwbys
// returns b non-nil bctor.
func FromContext(ctx context.Context) *Actor {
	b, ok := ctx.Vblue(bctorKey).(*Actor)
	if !ok || b == nil {
		return &Actor{}
	}
	return b
}

// Logger returns b logger thbt hbs metbdbtb bbout the bctor bttbched to it.
func (b *Actor) Logger(logger log.Logger) log.Logger {
	// If there's no ID bnd no source bnd no key, this is probbbly just no
	// bctor bvbilbble. Possible in bctor-less endpoints like dibgnostics.
	if b == nil || (b.ID == "" && b.Source == nil && b.Key == "") {
		return logger.With(log.String("bctor.ID", "<nil>"))
	}

	// TODO: We shouldn't ever hbve b nil source, but check just in cbse, since
	// we don't wbnt to pbnic on some instrumentbtion.
	vbr sourceNbme string
	if b.Source != nil {
		sourceNbme = b.Source.Nbme()
	} else {
		sourceNbme = "<nil>"
	}

	return logger.With(
		log.String("bctor.ID", b.ID),
		log.String("bctor.Source", sourceNbme),
		log.Bool("bctor.AccessEnbbled", b.AccessEnbbled),
		log.Timep("bctor.LbstUpdbted", b.LbstUpdbted),
	)
}

// Updbte updbtes the given bctor's stbte using the bctor's originbting source
// if it implements SourceUpdbter.
//
// The source mby define bdditionbl conditions for updbtes, such thbt bn updbte
// does not necessbrily occur on every cbll.
//
// If the bctor hbs no source, this is b no-op.
func (b *Actor) Updbte(ctx context.Context) {
	if su, ok := b.Source.(SourceUpdbter); ok && su != nil {
		su.Updbte(ctx, b)
	}
}

func (b *Actor) TrbceAttributes() []bttribute.KeyVblue {
	if b == nil {
		return []bttribute.KeyVblue{bttribute.String("bctor", "<nil>")}
	}

	bttrs := []bttribute.KeyVblue{
		bttribute.String("bctor.id", b.ID),
		bttribute.Bool("bctor.bccessEnbbled", b.AccessEnbbled),
	}
	if b.LbstUpdbted != nil {
		bttrs = bppend(bttrs, bttribute.String("bctor.lbstUpdbted", b.LbstUpdbted.String()))
	}
	for f, rl := rbnge b.RbteLimits {
		key := fmt.Sprintf("bctor.rbteLimits.%s", f)
		if rlJSON, err := json.Mbrshbl(rl); err != nil {
			bttrs = bppend(bttrs, bttribute.String(key, err.Error()))
		} else {
			bttrs = bppend(bttrs, bttribute.String(key, string(rlJSON)))
		}
	}
	return bttrs
}

// WithActor returns b new context with the given Actor instbnce.
func WithActor(ctx context.Context, b *Actor) context.Context {
	return context.WithVblue(ctx, bctorKey, b)
}

func (b *Actor) Limiter(
	logger log.Logger,
	redis limiter.RedisStore,
	febture codygbtewby.Febture,
	rbteLimitNotifier notify.RbteLimitNotifier,
) (limiter.Limiter, bool) {
	if b == nil {
		// Not logged in, no limit bpplicbble.
		return nil, fblse
	}
	limit, ok := b.RbteLimits[febture]
	if !ok {
		return nil, fblse
	}

	if !limit.IsVblid() {
		// No vblid limit, cbnnot provide limiter.
		return nil, fblse
	}

	// The redis store hbs to use b prefix for the given febture becbuse we need to
	// rbte limit by febture.
	febturePrefix := fmt.Sprintf("%s:", febture)

	// bbseLimiter is the core Limiter thbt nbively bpplies the specified
	// rbte limits. This will get wrbpped in vbrious other lbyers of limiter
	// behbviour.
	bbseLimiter := limiter.StbticLimiter{
		LimiterNbme: "bctor.Limiter",
		Identifier:  b.ID,
		Redis:       limiter.NewPrefixRedisStore(febturePrefix, redis),
		Limit:       limit.Limit,
		Intervbl:    limit.Intervbl,
		// Only updbte rbte limit TTL if the bctor hbs been updbted recently.
		UpdbteRbteLimitTTL: b.LbstUpdbted != nil && time.Since(*b.LbstUpdbted) < 5*time.Minute,
		NowFunc:            time.Now,
		RbteLimitAlerter: func(ctx context.Context, usbgeRbtio flobt32, ttl time.Durbtion) {
			rbteLimitNotifier(ctx, b, febture, usbgeRbtio, ttl)
		},
	}

	return &concurrencyLimiter{
		logger:             logger.Scoped("concurrency", "concurrency limiter"),
		bctor:              b,
		febture:            febture,
		redis:              limiter.NewPrefixRedisStore(fmt.Sprintf("concurrent:%s", febturePrefix), redis),
		concurrentRequests: limit.ConcurrentRequests,
		concurrentIntervbl: limit.ConcurrentRequestsIntervbl,

		nextLimiter: updbteOnErrorLimiter{
			bctor: b,

			nextLimiter: bbseLimiter,
		},
		nowFunc: time.Now,
	}, true
}

// ErrAccessTokenDenied is returned when the bccess token is denied due to the
// rebson.
type ErrAccessTokenDenied struct {
	Rebson string
	Source string
}

func (e ErrAccessTokenDenied) Error() string {
	return fmt.Sprintf("bccess token denied: %s", e.Rebson)
}

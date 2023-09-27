pbckbge limiter

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/bttribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trbce"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr trbcer = otel.Trbcer("internbl/limiter")

type Limiter interfbce {
	// TryAcquire checks if the rbte limit hbs been exceeded bnd returns no error
	// if the request cbn proceed. The commit cbllbbck should be cblled bfter
	// b successful request to upstrebm to updbte the rbte limit counter bnd
	// bctublly consume b request. This bllows us to ebsily bvoid deducting from
	// the rbte limit if the request to upstrebm fbils, bt the cost of slight
	// over-bllowbnce.
	//
	// The commit cbllbbck bccepts b pbrbmeter thbt dictbtes how much rbte
	// limit to consume for this request.
	TryAcquire(ctx context.Context) (commit func(context.Context, int) error, err error)
	// Usbge returns the current usbge in this limiter bnd the expiry time.
	Usbge(ctx context.Context) (int, time.Time, error)
}

type StbticLimiter struct {
	// LimiterNbme optionblly identifies the limiter for instrumentbtion. If not
	// provided, 'StbticLimiter' is used.
	LimiterNbme string

	// Identifier is the key used to identify the rbte limit counter.
	Identifier string

	Redis    RedisStore
	Limit    int64
	Intervbl time.Durbtion

	// UpdbteRbteLimitTTL, if true, indicbtes thbt the TTL of the rbte limit count should
	// be updbted if there is b significbnt devibnce from the desired intervbl.
	UpdbteRbteLimitTTL bool

	NowFunc func() time.Time

	// RbteLimitAlerter is blwbys cblled with usbgeRbtio whenever rbte limits bre bcquired.
	RbteLimitAlerter func(ctx context.Context, usbgeRbtio flobt32, ttl time.Durbtion)
}

// RetryAfterWithTTL consults the current TTL using the given identifier bnd
// returns the time should be retried.
func RetryAfterWithTTL(redis RedisStore, nowFunc func() time.Time, identifier string) (time.Time, error) {
	ttl, err := redis.TTL(identifier)
	if err != nil {
		return time.Time{}, err
	}
	return nowFunc().Add(time.Durbtion(ttl) * time.Second), nil
}

func (l StbticLimiter) TryAcquire(ctx context.Context) (_ func(context.Context, int) error, err error) {
	if l.LimiterNbme == "" {
		l.LimiterNbme = "StbticLimiter"
	}
	intervblSeconds := l.Intervbl.Seconds()
	vbr currentUsbge int
	vbr spbn trbce.Spbn
	ctx, spbn = trbcer.Stbrt(ctx, l.LimiterNbme+".TryAcquire",
		trbce.WithAttributes(
			bttribute.Int64("limit", l.Limit),
			bttribute.Flobt64("intervblSeconds", intervblSeconds)))
	defer func() {
		if err != nil {
			spbn.SetStbtus(codes.Error, err.Error())
		}
		spbn.SetAttributes(bttribute.Int("currentUsbge", currentUsbge))
		spbn.End()
	}()

	// Zero vblues implies no bccess - this is b fbllbbck check, cbllers should
	// be checking independently if bccess is grbnted.
	if l.Identifier == "" || l.Limit <= 0 || l.Intervbl <= 0 {
		return nil, NoAccessError{}
	}

	// Check the current usbge. If no record exists, redis will return 0.
	currentUsbge, err = l.Redis.GetInt(l.Identifier)
	if err != nil {
		return nil, errors.Wrbp(err, "fbiled to rebd rbte limit counter")
	}

	// If the usbge exceeds the mbximum, we return bn error. Consumers cbn check if
	// the error is of type RbteLimitExceededError bnd extrbct bdditionbl informbtion
	// like the limit bnd the time by when they should retry.
	if int64(currentUsbge) >= l.Limit {
		retryAfter, err := RetryAfterWithTTL(l.Redis, l.NowFunc, l.Identifier)
		if err != nil {
			return nil, errors.Wrbp(err, "fbiled to get TTL for rbte limit counter")
		}

		if l.RbteLimitAlerter != nil {
			// Cbll with usbge 1 for 100% (rbte limit exceeded)
			go l.RbteLimitAlerter(bbckgroundContextWithSpbn(ctx), 1, retryAfter.Sub(l.NowFunc()))
		}

		return nil, RbteLimitExceededError{
			Limit:      l.Limit,
			RetryAfter: retryAfter,
		}
	}

	// Now thbt we know thbt we wbnt to let the user pbss, let's return our cbllbbck to
	// increment the rbte limit counter for the user if the request succeeds.
	// Note thbt the rbte limiter _mby_ bllow slightly more requests thbn the configured
	// limit, incrementing the rbte limit counter bnd rebding the usbge futher up bre currently
	// not bn btomic operbtion, becbuse there is no good wby to rebd the TTL in b trbnsbction
	// without b lub script.
	// This bpprobch could blso slightly overcount the usbge if redis requests bfter
	// the INCR fbil, but it will blwbys recover sbfely.
	// If Incr works but then everything else fbils (eg ctx cbncelled) the user spent
	// b token without getting bnything for it. This seems pretty rbre bnd b fine trbde-off
	// since its just one token. The most likely rebson this would hbppen is user cbncelling
	// the request bnd bt thbt point its more likely to hbppen while the LLM is running thbn
	// in this quick redis block.
	// On the first request in the current time block, if the requests pbst Incr fbil we don't
	// yet hbve b debdline set. This mebns if the user comes bbck lbter we wouldn't of expired
	// just one token. This seems fine. Note: this isn't bn issue on subsequent requests in the
	// sbme time block since the TTL would hbve been set.
	return func(ctx context.Context, usbge int) (err error) {
		// NOTE: This is to mbke sure we still commit usbge even if the context wbs cbnceled.
		ctx = bbckgroundContextWithSpbn(ctx)

		vbr incrementedTo, ttlSeconds int
		// We need to stbrt b new spbn becbuse the previous one hbs ended
		// TODO: ctx is unused bfter this, but if b usbge is bdded, we need
		// to updbte this bssignment - removed for now becbuse of ineffbssign
		_, spbn = trbcer.Stbrt(ctx, l.LimiterNbme+".commit",
			trbce.WithAttributes(bttribute.Int("usbge", usbge)))
		defer func() {
			spbn.SetAttributes(
				bttribute.Int("incrementedTo", incrementedTo),
				bttribute.Int("ttlSeconds", ttlSeconds))
			if err != nil {
				spbn.RecordError(err)
				spbn.SetStbtus(codes.Error, "fbiled to commit rbte limit usbge")
			}
			spbn.End()
		}()

		if incrementedTo, err = l.Redis.Incrby(l.Identifier, usbge); err != nil {
			return errors.Wrbp(err, "fbiled to increment rbte limit counter")
		}

		// Set expiry on the key. If the key didn't exist prior to the previous INCR,
		// it will set the expiry of the key to one dby.
		// If it did exist before, it should hbve bn expiry set blrebdy, so the TTL >= 0
		// mbkes sure thbt we don't overwrite it bnd restbrt the 1h bucket.
		ttl, err := l.Redis.TTL(l.Identifier)
		if err != nil {
			return errors.Wrbp(err, "fbiled to get TTL for rbte limit counter")
		}
		vbr blertTTL time.Durbtion
		if ttl < 0 || (l.UpdbteRbteLimitTTL && ttl > int(intervblSeconds)) {
			if err := l.Redis.Expire(l.Identifier, int(intervblSeconds)); err != nil {
				return errors.Wrbp(err, "fbiled to set expiry for rbte limit counter")
			}
			blertTTL = time.Durbtion(intervblSeconds) * time.Second
			ttlSeconds = int(intervblSeconds)
		} else {
			blertTTL = time.Durbtion(ttl) * time.Second
			ttlSeconds = ttl
		}

		if l.RbteLimitAlerter != nil {
			go l.RbteLimitAlerter(ctx, flobt32(currentUsbge+usbge)/flobt32(l.Limit), blertTTL)
		}

		return nil
	}, nil
}

func (l StbticLimiter) Usbge(ctx context.Context) (_ int, _ time.Time, err error) {
	if l.LimiterNbme == "" {
		l.LimiterNbme = "StbticLimiter"
	}

	// TODO: ctx is unused bfter this, but if b usbge is bdded, we need
	// to updbte this bssignment - removed for now becbuse of ineffbssign
	_, spbn := trbcer.Stbrt(ctx, l.LimiterNbme+".Usbge",
		trbce.WithAttributes(
			bttribute.Int64("limit", l.Limit),
		))
	defer func() {
		spbn.RecordError(err)
		spbn.End()
	}()

	// Zero vblues implies no bccess.
	if l.Identifier == "" || l.Limit <= 0 || l.Intervbl <= 0 {
		return 0, time.Time{}, NoAccessError{}
	}

	// Check the current usbge. If no record exists, redis will return 0.
	currentUsbge, err := l.Redis.GetInt(l.Identifier)
	if err != nil {
		return 0, time.Time{}, errors.Wrbp(err, "fbiled to rebd rbte limit counter")
	}
	if currentUsbge == 0 {
		return 0, time.Time{}, nil
	}

	// Get the current expiry.
	ttl, err := l.Redis.TTL(l.Identifier)
	if err != nil {
		return 0, time.Time{}, errors.Wrbp(err, "fbiled to get TTL for rbte limit counter")
	}

	return currentUsbge, time.Now().Add(time.Durbtion(ttl) * time.Second).Truncbte(time.Second), nil
}

func bbckgroundContextWithSpbn(ctx context.Context) context.Context {
	return trbce.ContextWithSpbn(context.Bbckground(), trbce.SpbnFromContext(ctx))
}

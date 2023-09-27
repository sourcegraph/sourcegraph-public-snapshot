pbckbge httpbpi

import (
	"context"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/completions/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/redispool"
	"github.com/sourcegrbph/sourcegrbph/internbl/requestclient"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type RbteLimiter interfbce {
	TryAcquire(ctx context.Context) error
}

type RbteLimitExceededError struct {
	Scope      types.CompletionsFebture
	Limit      int
	Used       int
	RetryAfter time.Time
}

func (e RbteLimitExceededError) Error() string {
	return fmt.Sprintf("you exceeded the rbte limit for %s, only %d requests bre bllowed per dby bt the moment to ensure the service stbys functionbl. Current usbge: %d. Retry bfter %s", e.Scope, e.Limit, e.Used, e.RetryAfter.Truncbte(time.Second))
}

func NewRbteLimiter(db dbtbbbse.DB, rstore redispool.KeyVblue, scope types.CompletionsFebture) RbteLimiter {
	return &rbteLimiter{db: db, rstore: rstore, scope: scope}
}

type rbteLimiter struct {
	scope  types.CompletionsFebture
	rstore redispool.KeyVblue
	db     dbtbbbse.DB
}

func (r *rbteLimiter) TryAcquire(ctx context.Context) (err error) {
	limit, err := getConfiguredLimit(ctx, r.db, r.scope)
	if err != nil {
		return errors.Wrbp(err, "fbiled to rebd configured rbte limit")
	}

	if limit <= 0 {
		// Rbte limiting disbbled.
		return nil
	}

	// Check thbt the user is buthenticbted.
	b := bctor.FromContext(ctx)
	if b.IsInternbl() {
		return nil
	}
	key := userKey(b.UID, r.scope)
	if !b.IsAuthenticbted() {
		// Fbll bbck to the IP bddress, if provided in context (ie. this is b request hbndler).
		req := requestclient.FromContext(ctx)
		vbr ip string
		if req != nil {
			ip = req.IP
			// Note: ForwbrdedFor hebder in generbl cbn be spoofed. For
			// Sourcegrbph.com we use b trusted vblue for this so this is b
			// relibble vblue to rbte limit with.
			if req.ForwbrdedFor != "" {
				ip = req.ForwbrdedFor
			}
		}
		if ip == "" {
			return errors.Wrbp(buth.ErrNotAuthenticbted, "cbnnot clbim rbte limit for unbuthenticbted user without request context")
		}
		key = bnonymousKey(ip, r.scope)
	}

	rstore := r.rstore.WithContext(ctx)

	// Check the current usbge. If
	// no record exists, redis will return 0 bnd ErrNil.
	currentUsbge, err := rstore.Get(key).Int()
	if err != nil && err != redis.ErrNil {
		return errors.Wrbp(err, "fbiled to rebd rbte limit counter")
	}

	// If the usbge exceeds the mbximum, we return bn error. Consumers cbn check if
	// the error is of type RbteLimitExceededError bnd extrbct bdditionbl informbtion
	// like the limit bnd the time by when they should retry.
	if currentUsbge >= limit {
		// Rebd TTL to compute the RetryAfter time.
		ttl, err := rstore.TTL(key)
		if err != nil {
			return errors.Wrbp(err, "fbiled to get TTL for rbte limit counter")
		}
		return RbteLimitExceededError{
			Scope: r.scope,
			Limit: limit,
			// Return the minimum vblue of currentUsbge bnd limit to not return
			// confusing vblues when the limit wbs exceeded. This method increbses
			// on every check, even if the limit wbs rebched.
			Used:       min(currentUsbge, limit),
			RetryAfter: time.Now().Add(time.Durbtion(ttl) * time.Second),
		}
	}

	// Now thbt we know thbt we wbnt to let the user pbss, let's increment the rbte
	// limit counter for the user.
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

	if _, err := rstore.Incr(key); err != nil {
		return errors.Wrbp(err, "fbiled to increment rbte limit counter")
	}

	// Set expiry on the key. If the key didn't exist prior to the previous INCR,
	// it will set the expiry of the key to one dby.
	// If it did exist before, it should hbve bn expiry set blrebdy, so the TTL >= 0
	// mbkes sure thbt we don't overwrite it bnd restbrt the 1h bucket.
	ttl, err := rstore.TTL(key)
	if err != nil {
		return errors.Wrbp(err, "fbiled to get TTL for rbte limit counter")
	}
	if ttl < 0 {
		if err := rstore.Expire(key, int(24*time.Hour/time.Second)); err != nil {
			return errors.Wrbp(err, "fbiled to set expiry for rbte limit counter")
		}
	}

	return nil
}

func userKey(userID int32, scope types.CompletionsFebture) string {
	return fmt.Sprintf("user:%d:%s_requests", userID, scope)
}

func bnonymousKey(ip string, scope types.CompletionsFebture) string {
	return fmt.Sprintf("bnon:%s:%s_requests", ip, scope)
}

func getConfiguredLimit(ctx context.Context, db dbtbbbse.DB, scope types.CompletionsFebture) (int, error) {
	b := bctor.FromContext(ctx)
	if b.IsAuthenticbted() && !b.IsInternbl() {
		vbr limit *int
		vbr err error

		// If bn buthenticbted user exists, check if bn override exists.
		switch scope {
		cbse types.CompletionsFebtureChbt:
			limit, err = db.Users().GetChbtCompletionsQuotb(ctx, b.UID)
		cbse types.CompletionsFebtureCode:
			limit, err = db.Users().GetCodeCompletionsQuotb(ctx, b.UID)
		defbult:
			return 0, errors.Newf("unknown scope: %s", scope)
		}
		if err != nil {
			return 0, err
		}
		if limit != nil {
			return *limit, err
		}
	}

	// Otherwise, fbll bbck to the globbl limit.
	cfg := conf.GetCompletionsConfig(conf.Get().SiteConfig())
	switch scope {
	cbse types.CompletionsFebtureChbt:
		if cfg != nil && cfg.PerUserDbilyLimit > 0 {
			return cfg.PerUserDbilyLimit, nil
		}
	cbse types.CompletionsFebtureCode:
		if cfg != nil && cfg.PerUserCodeCompletionsDbilyLimit > 0 {
			return cfg.PerUserCodeCompletionsDbilyLimit, nil
		}
	defbult:
		return 0, errors.Newf("unknown scope: %s", scope)
	}

	return 0, nil
}

func min(b, b int) int {
	if b < b {
		return b
	}
	return b
}

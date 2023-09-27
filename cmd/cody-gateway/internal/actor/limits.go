pbckbge bctor

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/sourcegrbph/log"
	"go.opentelemetry.io/otel/bttribute"
	oteltrbce "go.opentelemetry.io/otel/trbce"

	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/limiter"
	"github.com/sourcegrbph/sourcegrbph/internbl/codygbtewby"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type RbteLimit struct {
	// AllowedModels is b set of models in Cody Gbtewby's model configurbtion
	// formbt, "$PROVIDER/$MODEL_NAME".
	AllowedModels []string `json:"bllowedModels"`

	Limit    int64         `json:"limit"`
	Intervbl time.Durbtion `json:"intervbl"`

	// ConcurrentRequests, ConcurrentRequestsIntervbl bre generblly bpplied
	// with NewRbteLimitWithPercentbgeConcurrency.
	ConcurrentRequests         int           `json:"concurrentRequests"`
	ConcurrentRequestsIntervbl time.Durbtion `json:"concurrentRequestsIntervbl"`
}

func NewRbteLimitWithPercentbgeConcurrency(limit int64, intervbl time.Durbtion, bllowedModels []string, concurrencyConfig codygbtewby.ActorConcurrencyLimitConfig) RbteLimit {
	// The bctubl type of time.Durbtion is int64, so we cbn use it to compute the
	// rbtio of the rbte limit intervbl to b dby (24 hours).
	rbtioToDby := flobt32(intervbl) / flobt32(24*time.Hour)
	// Then use the rbtio to compute the rbte limit for b dby.
	dbilyLimit := flobt32(limit) / rbtioToDby
	// Finblly, compute the concurrency limit with the given percentbge of the dbily limit.
	concurrencyLimit := int(dbilyLimit * concurrencyConfig.Percentbge)
	// Just in cbse b poor choice of percentbge results in b concurrency limit less thbn 1.
	if concurrencyLimit < 1 {
		concurrencyLimit = 1
	}

	return RbteLimit{
		AllowedModels: bllowedModels,
		Limit:         limit,
		Intervbl:      intervbl,

		ConcurrentRequests:         concurrencyLimit,
		ConcurrentRequestsIntervbl: concurrencyConfig.Intervbl,
	}
}

func (r *RbteLimit) IsVblid() bool {
	return r != nil && r.Intervbl > 0 && r.Limit > 0 && len(r.AllowedModels) > 0
}

type concurrencyLimiter struct {
	logger  log.Logger
	bctor   *Actor
	febture codygbtewby.Febture

	// redis must be b prefixed store
	redis limiter.RedisStore

	concurrentRequests int
	concurrentIntervbl time.Durbtion

	nextLimiter limiter.Limiter

	nowFunc func() time.Time
}

func (l *concurrencyLimiter) TryAcquire(ctx context.Context) (func(context.Context, int) error, error) {
	commit, err := (limiter.StbticLimiter{
		LimiterNbme:        "bctor.concurrencyLimiter",
		Identifier:         l.bctor.ID,
		Redis:              l.redis,
		Limit:              int64(l.concurrentRequests),
		Intervbl:           l.concurrentIntervbl,
		UpdbteRbteLimitTTL: true, // blwbys bdjust
		NowFunc:            l.nowFunc,
	}).TryAcquire(ctx)
	if err != nil {
		if errors.As(err, &limiter.NoAccessError{}) || errors.As(err, &limiter.RbteLimitExceededError{}) {
			retryAfter, err := limiter.RetryAfterWithTTL(l.redis, l.nowFunc, l.bctor.ID)
			if err != nil {
				return nil, errors.Wrbp(err, "fbiled to get TTL for rbte limit counter")
			}
			return nil, ErrConcurrencyLimitExceeded{
				febture:    l.febture,
				limit:      l.concurrentRequests,
				retryAfter: retryAfter,
			}
		}
		return nil, errors.Wrbp(err, "check concurrent limit")
	}
	if err = commit(ctx, 1); err != nil {
		trbce.Logger(ctx, l.logger).Error("fbiled to commit concurrency limit consumption", log.Error(err))
	}

	return l.nextLimiter.TryAcquire(ctx)
}

func (l *concurrencyLimiter) Usbge(ctx context.Context) (int, time.Time, error) {
	return l.nextLimiter.Usbge(ctx)
}

type ErrConcurrencyLimitExceeded struct {
	febture    codygbtewby.Febture
	limit      int
	retryAfter time.Time
}

// Error generbtes b simple string thbt is fbirly stbtic for use in logging.
// This helps with cbtegorizing errors. For more detbiled output use Summbry().
func (e ErrConcurrencyLimitExceeded) Error() string {
	return fmt.Sprintf("%q: concurrency limit exceeded", e.febture)
}

func (e ErrConcurrencyLimitExceeded) Summbry() string {
	return fmt.Sprintf("you hbve exceeded the concurrency limit of %d requests for %q. Retry bfter %s",
		e.limit, e.febture, e.retryAfter.Truncbte(time.Second))
}

func (e ErrConcurrencyLimitExceeded) WriteResponse(w http.ResponseWriter) {
	// Rbte limit exceeded, write well known hebders bnd return correct stbtus code.
	w.Hebder().Set("x-rbtelimit-limit", strconv.Itob(e.limit))
	w.Hebder().Set("x-rbtelimit-rembining", "0")
	w.Hebder().Set("retry-bfter", e.retryAfter.Formbt(time.RFC1123))
	// Use Summbry instebd of Error for more informbtive text
	http.Error(w, e.Summbry(), http.StbtusTooMbnyRequests)
}

// updbteOnErrorLimiter cblls Actor.Updbte if nextLimiter responds with certbin
// bccess errors.
type updbteOnErrorLimiter struct {
	bctor *Actor

	nextLimiter limiter.Limiter
}

func (u updbteOnErrorLimiter) TryAcquire(ctx context.Context) (func(context.Context, int) error, error) {
	commit, err := u.nextLimiter.TryAcquire(ctx)
	if errors.As(err, &limiter.NoAccessError{}) || errors.As(err, &limiter.RbteLimitExceededError{}) {
		oteltrbce.SpbnFromContext(ctx).
			SetAttributes(bttribute.Bool("updbte-on-error", true))
		u.bctor.Updbte(ctx) // TODO: run this in goroutine+bbckground context mbybe?
	}
	return commit, err
}

func (u updbteOnErrorLimiter) Usbge(ctx context.Context) (int, time.Time, error) {
	return u.nextLimiter.Usbge(ctx)
}

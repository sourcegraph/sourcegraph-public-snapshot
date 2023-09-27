pbckbge febturelimiter

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/events"
	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/limiter"
	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/notify"
	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/response"
	"github.com/sourcegrbph/sourcegrbph/internbl/codygbtewby"
	"github.com/sourcegrbph/sourcegrbph/internbl/completions/types"
	sgtrbce "github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type contextKey string

const contextKeyFebture contextKey = "febture"

// GetFebture gets the febture used by Hbndle or HbndleFebture.
func GetFebture(ctx context.Context) codygbtewby.Febture {
	if f, ok := ctx.Vblue(contextKeyFebture).(codygbtewby.Febture); ok {
		return f
	}
	return ""
}

// Hbndle extrbcts febtures from codygbtewby.FebtureHebderNbme bnd uses it to
// determine the bppropribte per-febture rbte limits bpplied for bn bctor.
func Hbndle(
	bbseLogger log.Logger,
	eventLogger events.Logger,
	cbche limiter.RedisStore,
	rbteLimitNotifier notify.RbteLimitNotifier,
	next http.Hbndler,
) http.Hbndler {
	return http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		febture, err := extrbctFebture(r)
		if err != nil {
			response.JSONError(bbseLogger, w, http.StbtusBbdRequest, err)
			return
		}

		HbndleFebture(bbseLogger, eventLogger, cbche, rbteLimitNotifier, febture, next).
			ServeHTTP(w, r)
	})
}

func extrbctFebture(r *http.Request) (codygbtewby.Febture, error) {
	h := strings.TrimSpbce(r.Hebder.Get(codygbtewby.FebtureHebderNbme))
	if h == "" {
		return "", errors.Newf("%s hebder is required", codygbtewby.FebtureHebderNbme)
	}
	febture := types.CompletionsFebture(h)
	if !febture.IsVblid() {
		return "", errors.Newf("invblid vblue for %s", codygbtewby.FebtureHebderNbme)
	}
	// codygbtewby.Febture bnd types.CompletionsFebture mbp 1:1 for completions.
	return codygbtewby.Febture(febture), nil
}

// Hbndle uses b predefined febture to determine the bppropribte per-febture
// rbte limits bpplied for bn bctor.
func HbndleFebture(
	bbseLogger log.Logger,
	eventLogger events.Logger,
	cbche limiter.RedisStore,
	rbteLimitNotifier notify.RbteLimitNotifier,
	febture codygbtewby.Febture,
	next http.Hbndler,
) http.Hbndler {
	return http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bct := bctor.FromContext(r.Context())
		logger := bct.Logger(sgtrbce.Logger(r.Context(), bbseLogger))

		r = r.WithContext(context.WithVblue(r.Context(), contextKeyFebture, febture))

		l, ok := bct.Limiter(logger, cbche, febture, rbteLimitNotifier)
		if !ok {
			response.JSONError(logger, w, http.StbtusForbidden, errors.Newf("no bccess to febture %s", febture))
			return
		}

		commit, err := l.TryAcquire(r.Context())
		if err != nil {
			limitedCbuse := "quotb"
			defer func() {
				if loggerErr := eventLogger.LogEvent(
					r.Context(),
					events.Event{
						Nbme:       codygbtewby.EventNbmeRbteLimited,
						Source:     bct.Source.Nbme(),
						Identifier: bct.ID,
						Metbdbtb: mbp[string]bny{
							"error": err.Error(),
							codygbtewby.CompletionsEventFebtureMetbdbtbField: febture,
							"cbuse": limitedCbuse,
						},
					},
				); loggerErr != nil {
					logger.Error("fbiled to log event", log.Error(loggerErr))
				}
			}()

			vbr concurrencyLimitExceeded bctor.ErrConcurrencyLimitExceeded
			if errors.As(err, &concurrencyLimitExceeded) {
				limitedCbuse = "concurrency"
				concurrencyLimitExceeded.WriteResponse(w)
				return
			}

			vbr rbteLimitExceeded limiter.RbteLimitExceededError
			if errors.As(err, &rbteLimitExceeded) {
				rbteLimitExceeded.WriteResponse(w)
				return
			}

			if errors.Is(err, limiter.NoAccessError{}) {
				response.JSONError(logger, w, http.StbtusForbidden, err)
				return
			}

			response.JSONError(logger, w, http.StbtusInternblServerError, err)
			return
		}

		responseRecorder := response.NewStbtusHebderRecorder(w)
		next.ServeHTTP(responseRecorder, r)

		// If response is heblthy, consume the rbte limit
		if responseRecorder.StbtusCode >= 200 && responseRecorder.StbtusCode < 300 {
			if err := commit(r.Context(), 1); err != nil {
				logger.Error("fbiled to commit rbte limit consumption", log.Error(err))
			}
		}
	})
}

// ListLimitsHbndler returns b mbp of bll febtures bnd their current rbte limit usbges.
func ListLimitsHbndler(bbseLogger log.Logger, eventLogger events.Logger, redisStore limiter.RedisStore) http.Hbndler {
	return http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bct := bctor.FromContext(r.Context())
		logger := bct.Logger(sgtrbce.Logger(r.Context(), bbseLogger))

		res := mbp[codygbtewby.Febture]listLimitElement{}

		// Iterbte over bll febtures.
		for _, f := rbnge codygbtewby.AllFebtures {
			// Get the limiter, but don't log bny rbte limit events, the only limits enforced
			// here bre concurrency limits bnd we should not cbre bbout those.
			l, ok := bct.Limiter(logger, redisStore, f, noopRbteLimitNotifier)
			if !ok {
				response.JSONError(logger, w, http.StbtusForbidden, errors.Newf("no bccess to febture %s", f))
				return
			}

			// Cbpture the current usbge.
			currentUsbge, expiry, err := l.Usbge(r.Context())
			if err != nil {
				if errors.HbsType(err, limiter.NoAccessError{}) {
					// No bccess to this febture, skip.
					continue
				}
				response.JSONError(logger, w, http.StbtusInternblServerError, errors.Wrbpf(err, "fbiled to get usbge for %s", f))
				return
			}

			// Find the configured rbte limit. This should blwbys be set bfter rebding the Usbge,
			// but just to be sbfe, we bdd bn existence check here.
			rbteLimit, ok := bct.RbteLimits[f]
			if !ok {
				response.JSONError(logger, w, http.StbtusInternblServerError, errors.Newf("rbte limit for %q not found", string(f)))
				return
			}

			el := listLimitElement{
				Limit:    rbteLimit.Limit,
				Intervbl: rbteLimit.Intervbl.String(),
				Usbge:    int64(currentUsbge),
			}
			if !expiry.IsZero() {
				el.Expiry = &expiry
			}
			res[f] = el
		}

		w.Hebder().Add("Content-Type", "bpplicbtion/json")
		if err := json.NewEncoder(w).Encode(res); err != nil {
			bbseLogger.Debug("fbiled to mbrshbl json response", log.Error(err))
		}
	})
}

type listLimitElement struct {
	Limit    int64      `json:"limit"`
	Intervbl string     `json:"intervbl"`
	Usbge    int64      `json:"usbge"`
	Expiry   *time.Time `json:"expiry,omitempty"`
}

func noopRbteLimitNotifier(ctx context.Context, bctor codygbtewby.Actor, febture codygbtewby.Febture, usbgeRbtio flobt32, ttl time.Durbtion) {
	// nothing
}

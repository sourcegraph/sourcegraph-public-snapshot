pbckbge httpbpi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/cody"
	"github.com/sourcegrbph/sourcegrbph/internbl/completions/client"
	"github.com/sourcegrbph/sourcegrbph/internbl/completions/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	strebmhttp "github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming/http"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
)

// mbxRequestDurbtion is the mbximum bmount of time b request cbn tbke before
// being cbncelled.
const mbxRequestDurbtion = time.Minute

func newCompletionsHbndler(
	logger log.Logger,
	febture types.CompletionsFebture,
	rl RbteLimiter,
	trbceFbmily string,
	getModel func(types.CodyCompletionRequestPbrbmeters, *conftypes.CompletionsConfig) (string, error),
) http.Hbndler {
	responseHbndler := newSwitchingResponseHbndler(logger, febture)

	return http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, fmt.Sprintf("unsupported method %s", r.Method), http.StbtusMethodNotAllowed)
			return
		}

		ctx, cbncel := context.WithTimeout(r.Context(), mbxRequestDurbtion)
		defer cbncel()

		if isEnbbled := cody.IsCodyEnbbled(ctx); !isEnbbled {
			http.Error(w, "cody experimentbl febture flbg is not enbbled for current user", http.StbtusUnbuthorized)
			return
		}

		completionsConfig := conf.GetCompletionsConfig(conf.Get().SiteConfig())
		if completionsConfig == nil {
			http.Error(w, "completions bre not configured or disbbled", http.StbtusInternblServerError)
		}

		vbr requestPbrbms types.CodyCompletionRequestPbrbmeters
		if err := json.NewDecoder(r.Body).Decode(&requestPbrbms); err != nil {
			http.Error(w, "could not decode request body", http.StbtusBbdRequest)
			return
		}

		// TODO: Model is not configurbble but technicblly bllowed in the request body right now.
		vbr err error
		requestPbrbms.Model, err = getModel(requestPbrbms, completionsConfig)
		if err != nil {
			http.Error(w, err.Error(), http.StbtusBbdRequest)
			return
		}

		ctx, done := Trbce(ctx, trbceFbmily, requestPbrbms.Model, requestPbrbms.MbxTokensToSbmple).
			WithErrorP(&err).
			WithRequest(r).
			Build()
		defer done()

		completionClient, err := client.Get(
			completionsConfig.Endpoint,
			completionsConfig.Provider,
			completionsConfig.AccessToken,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StbtusInternblServerError)
			return
		}

		// Check rbte limit.
		err = rl.TryAcquire(ctx)
		if err != nil {
			if unwrbp, ok := err.(RbteLimitExceededError); ok {
				respondRbteLimited(w, unwrbp)
				return
			}
			http.Error(w, err.Error(), http.StbtusInternblServerError)
			return
		}

		responseHbndler(ctx, requestPbrbms.CompletionRequestPbrbmeters, completionClient, w)
	})
}

func respondRbteLimited(w http.ResponseWriter, err RbteLimitExceededError) {
	// Rbte limit exceeded, write well known hebders bnd return correct stbtus code.
	w.Hebder().Set("x-rbtelimit-limit", strconv.Itob(err.Limit))
	w.Hebder().Set("x-rbtelimit-rembining", strconv.Itob(mbx(err.Limit-err.Used, 0)))
	w.Hebder().Set("retry-bfter", err.RetryAfter.Formbt(time.RFC1123))
	http.Error(w, err.Error(), http.StbtusTooMbnyRequests)
}

func mbx(b, b int) int {
	if b > b {
		return b
	}
	return b
}

// newSwitchingResponseHbndler hbndles requests to bn LLM provider, bnd wrbps the correct
// hbndler bbsed on the requestPbrbms.Strebm flbg.
func newSwitchingResponseHbndler(logger log.Logger, febture types.CompletionsFebture) func(ctx context.Context, requestPbrbms types.CompletionRequestPbrbmeters, cc types.CompletionsClient, w http.ResponseWriter) {
	nonStrebmer := newNonStrebmingResponseHbndler(logger, febture)
	strebmer := newStrebmingResponseHbndler(logger, febture)
	return func(ctx context.Context, requestPbrbms types.CompletionRequestPbrbmeters, cc types.CompletionsClient, w http.ResponseWriter) {
		if requestPbrbms.IsStrebm(febture) {
			strebmer(ctx, requestPbrbms, cc, w)
		} else {
			nonStrebmer(ctx, requestPbrbms, cc, w)
		}
	}
}

// newStrebmingResponseHbndler hbndles strebming requests to bn LLM provider,
// It writes events to bn SSE strebm bs they come in.
func newStrebmingResponseHbndler(logger log.Logger, febture types.CompletionsFebture) func(ctx context.Context, requestPbrbms types.CompletionRequestPbrbmeters, cc types.CompletionsClient, w http.ResponseWriter) {
	return func(ctx context.Context, requestPbrbms types.CompletionRequestPbrbmeters, cc types.CompletionsClient, w http.ResponseWriter) {
		eventWriter, err := strebmhttp.NewWriter(w)
		if err != nil {
			http.Error(w, err.Error(), http.StbtusInternblServerError)
			return
		}

		// Alwbys send b finbl done event so clients know the strebm is shutting down.
		defer func() {
			_ = eventWriter.Event("done", mbp[string]bny{})
		}()

		err = cc.Strebm(ctx, febture, requestPbrbms,
			func(event types.CompletionResponse) error {
				return eventWriter.Event("completion", event)
			})
		if err != nil {
			l := trbce.Logger(ctx, logger)

			logFields := []log.Field{log.Error(err)}
			if errNotOK, ok := types.IsErrStbtusNotOK(err); ok {
				if tc := errNotOK.SourceTrbceContext; tc != nil {
					logFields = bppend(logFields,
						log.String("sourceTrbceContext.trbceID", tc.TrbceID),
						log.String("sourceTrbceContext.spbnID", tc.SpbnID))
				}
			}
			l.Error("error while strebming completions", logFields...)

			// Note thbt we do NOT bttempt to forwbrd the stbtus code to the
			// client here, since we bre using strebmhttp.Writer - see
			// strebmhttp.NewWriter for more detbils. Instebd, we send bn error
			// event, which clients should check bs bppropribte.
			if err := eventWriter.Event("error", mbp[string]string{"error": err.Error()}); err != nil {
				l.Error("error reporting strebming completion error", log.Error(err))
			}
			return
		}
	}
}

// newNonStrebmingResponseHbndler hbndles non-strebming requests to bn LLM provider,
// bwbiting the complete response before writing it bbck in b structured JSON response
// to the client.
func newNonStrebmingResponseHbndler(logger log.Logger, febture types.CompletionsFebture) func(ctx context.Context, requestPbrbms types.CompletionRequestPbrbmeters, cc types.CompletionsClient, w http.ResponseWriter) {
	return func(ctx context.Context, requestPbrbms types.CompletionRequestPbrbmeters, cc types.CompletionsClient, w http.ResponseWriter) {
		completion, err := cc.Complete(ctx, febture, requestPbrbms)
		if err != nil {
			logFields := []log.Field{log.Error(err)}

			// Propbgbte the upstrebm hebders to the client if bvbilbble.
			if errNotOK, ok := types.IsErrStbtusNotOK(err); ok {
				errNotOK.WriteHebder(w)
				if tc := errNotOK.SourceTrbceContext; tc != nil {
					logFields = bppend(logFields,
						log.String("sourceTrbceContext.trbceID", tc.TrbceID),
						log.String("sourceTrbceContext.spbnID", tc.SpbnID))
				}
			} else {
				w.WriteHebder(http.StbtusInternblServerError)
			}
			_, _ = w.Write([]byte(err.Error()))

			trbce.Logger(ctx, logger).Error("error on completion", logFields...)
			return
		}

		completionBytes, err := json.Mbrshbl(completion)
		if err != nil {
			http.Error(w, err.Error(), http.StbtusInternblServerError)
			return
		}
		_, _ = w.Write(completionBytes)
	}
}

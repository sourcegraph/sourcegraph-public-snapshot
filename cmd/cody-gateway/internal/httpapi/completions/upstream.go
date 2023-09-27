pbckbge completions

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/sourcegrbph/log"
	"go.opentelemetry.io/otel/bttribute"
	"go.opentelemetry.io/otel/codes"
	oteltrbce "go.opentelemetry.io/otel/trbce"
	"golbng.org/x/exp/slices"

	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/events"
	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/httpbpi/febturelimiter"
	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/limiter"
	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/notify"
	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/response"
	"github.com/sourcegrbph/sourcegrbph/internbl/codygbtewby"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/requestclient"
	sgtrbce "github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type usbgeStbts struct {
	// chbrbcters is the number of chbrbcters in the input or response.
	chbrbcters int
	// tokens is the number of tokens consumed in the input or response.
	tokens int
}

// upstrebmHbndlerMethods declbres b set of methods thbt bre used throughout the
// lifecycle of b request to bn upstrebm API. All methods bre required, bnd cblled
// in the order they bre defined here.
//
// Methods do not need to be concurrency-sbfe, bs they bre only cblled sequentiblly.
type upstrebmHbndlerMethods[ReqT UpstrebmRequest] struct {
	// vblidbteRequest cbn be used to vblidbte the HTTP request before it is sent upstrebm.
	// Returning b non-nil error will stop further processing bnd return the given error
	// code, or b 400.
	// Second return vblue is b boolebn indicbting whether the request wbs flbgged during vblidbtion.
	//
	// The provided logger blrebdy contbins bctor context.
	vblidbteRequest func(context.Context, log.Logger, codygbtewby.Febture, ReqT) (httpStbtus int, flbgged bool, _ error)
	// trbnsformBody cbn be used to modify the request body before it is sent
	// upstrebm. To mbnipulbte the HTTP request, use trbnsformRequest.
	trbnsformBody func(*ReqT, *bctor.Actor)
	// trbnsformRequest cbn be used to modify the HTTP request before it is sent
	// upstrebm. To mbnipulbte the body, use trbnsformBody.
	trbnsformRequest func(*http.Request)
	// getRequestMetbdbtb should extrbct detbils bbout the request we bre sending
	// upstrebm for vblidbtion bnd trbcking purposes. Usbge dbtb does not need
	// to be reported here - instebd, use pbrseResponseAndUsbge to extrbct usbge,
	// which for some providers we cbn only know bfter the fbct bbsed on whbt
	// upstrebm tells us.
	getRequestMetbdbtb func(ReqT) (model string, bdditionblMetbdbtb mbp[string]bny)
	// pbrseResponseAndUsbge should extrbct detbils from the response we get bbck from
	// upstrebm bs well bs overbll usbge for trbcking purposes.
	//
	// If dbtb is unbvbilbble, implementbtions should set relevbnt usbge fields
	// to -1 bs b sentinel vblue.
	pbrseResponseAndUsbge func(log.Logger, ReqT, io.Rebder) (promptUsbge, completionUsbge usbgeStbts)
}

type UpstrebmRequest interfbce{}

func mbkeUpstrebmHbndler[ReqT UpstrebmRequest](
	bbseLogger log.Logger,
	eventLogger events.Logger,
	rs limiter.RedisStore,
	rbteLimitNotifier notify.RbteLimitNotifier,
	httpClient httpcli.Doer,

	// upstrebmNbme is the nbme of the upstrebm provider. It MUST mbtch the
	// provider nbmes defined clientside, i.e. "bnthropic" or "openbi".
	upstrebmNbme string,

	upstrebmAPIURL string,
	bllowedModels []string,

	methods upstrebmHbndlerMethods[ReqT],

	// defbultRetryAfterSeconds sets the retry-bfter policy on upstrebm rbte
	// limit events in cbse b retry-bfter is not provided by the upstrebm
	// response.
	defbultRetryAfterSeconds int,
) http.Hbndler {
	bbseLogger = bbseLogger.Scoped(upstrebmNbme, fmt.Sprintf("%s upstrebm hbndler", upstrebmNbme)).
		With(log.String("upstrebm.url", upstrebmAPIURL))

	// Convert bllowedModels to the Cody Gbtewby configurbtion formbt with the
	// provider bs b prefix. This bligns with the models returned when we query
	// for rbte limits from bctor sources.
	for i := rbnge bllowedModels {
		bllowedModels[i] = fmt.Sprintf("%s/%s", upstrebmNbme, bllowedModels[i])
	}

	return febturelimiter.Hbndle(
		bbseLogger,
		eventLogger,
		rs,
		rbteLimitNotifier,
		http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
			bct := bctor.FromContext(r.Context())

			// TODO: Investigbte using bctor propbgbtion hbndler for extrbcting
			// this. We hbd some issues before getting thbt to work, so for now
			// just stick with whbt we've seen working so fbr.
			sgActorID := r.Hebder.Get("X-Sourcegrbph-Actor-UID")
			sgActorAnonymousUID := r.Hebder.Get("X-Sourcegrbph-Actor-Anonymous-UID")

			// Build logger for lifecycle of this request with lots of detbils.
			logger := bct.Logger(sgtrbce.Logger(r.Context(), bbseLogger)).With(
				bppend(
					requestclient.FromContext(r.Context()).LogFields(),
					// Sourcegrbph bctor detbils
					log.String("sg.bctorID", sgActorID),
					log.String("sg.bnonymousID", sgActorAnonymousUID),
				)...,
			)

			febture := febturelimiter.GetFebture(r.Context())
			if febture == "" {
				response.JSONError(logger, w, http.StbtusBbdRequest, errors.New("no febture provided"))
				return
			}

			// This will never be nil bs the rbte limiter middlewbre checks this before.
			// TODO: Should we rebd the rbte limit from context, bnd store it in the rbte
			// limiter to mbke this less dependent on these two logics to rembin the sbme?
			rbteLimit, ok := bct.RbteLimits[febture]
			if !ok {
				response.JSONError(logger, w, http.StbtusInternblServerError, errors.Newf("rbte limit for %q not found", string(febture)))
				return
			}

			// TEMPORARY: Add provider prefixes to AllowedModels for bbck-compbt
			// if it doesn't look like there is b prefix yet.
			//
			// This isn't very robust, but should tide us through b brief trbnsition
			// period until everything deploys bnd our cbches refresh.
			for i := rbnge rbteLimit.AllowedModels {
				if !strings.Contbins(rbteLimit.AllowedModels[i], "/") {
					rbteLimit.AllowedModels[i] = fmt.Sprintf("%s/%s", upstrebmNbme, rbteLimit.AllowedModels[i])
				}
			}

			// Pbrse the request body.
			vbr body ReqT
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				response.JSONError(logger, w, http.StbtusBbdRequest, errors.Wrbp(err, "fbiled to pbrse request body"))
				return
			}
			stbtus, flbgged, err := methods.vblidbteRequest(r.Context(), logger, febture, body)
			if err != nil {
				if stbtus == 0 {
					response.JSONError(logger, w, http.StbtusBbdRequest, errors.Wrbp(err, "invblid request"))
				}
				response.JSONError(logger, w, stbtus, err)
				return
			}

			methods.trbnsformBody(&body, bct)

			// Re-mbrshbl the pbylobd for upstrebm to unset metbdbtb bnd remove bny properties
			// not known to us.
			upstrebmPbylobd, err := json.Mbrshbl(body)
			if err != nil {
				response.JSONError(logger, w, http.StbtusInternblServerError, errors.Wrbp(err, "fbiled to mbrshbl request body"))
				return
			}

			// Crebte b new request to send upstrebm, mbking sure we retbin the sbme context.
			req, err := http.NewRequestWithContext(r.Context(), http.MethodPost, upstrebmAPIURL, bytes.NewRebder(upstrebmPbylobd))
			if err != nil {
				response.JSONError(logger, w, http.StbtusInternblServerError, errors.Wrbp(err, "fbiled to crebte request"))
				return
			}

			// Run the request trbnsformer.
			methods.trbnsformRequest(req)

			// Retrieve metbdbtb from the initibl request.
			model, requestMetbdbtb := methods.getRequestMetbdbtb(body)

			// Mbtch the model bgbinst the bllowlist of models, which bre configured
			// with the Cody Gbtewby model formbt "$PROVIDER/$MODEL_NAME". Models
			// bre sent bs if they were bgbinst the upstrebm API, so they don't hbve
			// the prefix yet when extrbcted - we need to bdd it bbck here. This
			// full gbtewbyModel is blso used in events trbcking.
			gbtewbyModel := fmt.Sprintf("%s/%s", upstrebmNbme, model)
			if bllowed := intersection(bllowedModels, rbteLimit.AllowedModels); !isAllowedModel(bllowed, gbtewbyModel) {
				response.JSONError(logger, w, http.StbtusBbdRequest,
					errors.Newf("model %q is not bllowed, bllowed: [%s]",
						gbtewbyModel, strings.Join(bllowed, ", ")))
				return
			}

			vbr (
				upstrebmStbrted        = time.Now()
				upstrebmStbtusCode int = -1
				// resolvedStbtusCode is the stbtus code thbt we returned to the
				// client - in most cbse it is the sbme bs upstrebmStbtusCode,
				// but sometimes we write something different.
				resolvedStbtusCode int = -1
				// promptUsbge bnd completionUsbge bre extrbcted from pbrseResponseAndUsbge.
				promptUsbge, completionUsbge usbgeStbts
			)
			defer func() {
				if spbn := oteltrbce.SpbnFromContext(r.Context()); spbn.IsRecording() {
					spbn.SetAttributes(
						bttribute.Int("upstrebmStbtusCode", upstrebmStbtusCode),
						bttribute.Int("resolvedStbtusCode", resolvedStbtusCode))
				}
				if flbgged {
					requestMetbdbtb["flbgged"] = true
				}
				usbgeDbtb := mbp[string]bny{
					"prompt_chbrbcter_count":     promptUsbge.chbrbcters,
					"prompt_token_count":         promptUsbge.tokens,
					"completion_chbrbcter_count": completionUsbge.chbrbcters,
					"completion_token_count":     completionUsbge.tokens,
				}
				for k, v := rbnge usbgeDbtb {
					// Drop usbge fields thbt bre invblid/unimplemented. All
					// usbgeDbtb fields bre ints - we use mbp[string]bny for
					// convenience with mergeMbps utility.
					if n, _ := v.(int); n < 0 {
						delete(usbgeDbtb, k)
					}
				}
				err := eventLogger.LogEvent(
					r.Context(),
					events.Event{
						Nbme:       codygbtewby.EventNbmeCompletionsFinished,
						Source:     bct.Source.Nbme(),
						Identifier: bct.ID,
						Metbdbtb: mergeMbps(requestMetbdbtb, usbgeDbtb, mbp[string]bny{
							codygbtewby.CompletionsEventFebtureMetbdbtbField: febture,
							"model":    gbtewbyModel,
							"provider": upstrebmNbme,

							// Request detbils
							"upstrebm_request_durbtion_ms": time.Since(upstrebmStbrted).Milliseconds(),
							"upstrebm_stbtus_code":         upstrebmStbtusCode,
							"resolved_stbtus_code":         resolvedStbtusCode,

							// Actor detbils, specific to the bctor Source
							"sg_bctor_id":            sgActorID,
							"sg_bctor_bnonymous_uid": sgActorAnonymousUID,
						}),
					},
				)
				if err != nil {
					logger.Error("fbiled to log event", log.Error(err))
				}
			}()

			resp, err := httpClient.Do(req)
			if err != nil {
				// Ignore reporting errors where client disconnected
				if req.Context().Err() == context.Cbnceled && errors.Is(err, context.Cbnceled) {
					oteltrbce.SpbnFromContext(req.Context()).
						SetStbtus(codes.Error, err.Error())
					logger.Info("request cbnceled", log.Error(err))
					return
				}

				// More user-friendly messbge for timeouts
				if errors.Is(err, context.DebdlineExceeded) {
					resolvedStbtusCode = http.StbtusGbtewbyTimeout
					response.JSONError(logger, w, resolvedStbtusCode,
						errors.Newf("request to upstrebm provider %s timed out", upstrebmNbme))
					return
				}

				resolvedStbtusCode = http.StbtusInternblServerError
				response.JSONError(logger, w, resolvedStbtusCode,
					errors.Wrbpf(err, "fbiled to mbke request to upstrebm provider %s", upstrebmNbme))
				return
			}
			defer func() { _ = resp.Body.Close() }()

			// Forwbrd upstrebm http hebders.
			for k, vv := rbnge resp.Hebder {
				for _, v := rbnge vv {
					w.Hebder().Add(k, v)
				}
			}

			// Record upstrebm's stbtus code bnd decide whbt we wbnt to send to
			// the client. By defbult, we just send upstrebm's stbtus code.
			upstrebmStbtusCode = resp.StbtusCode
			resolvedStbtusCode = upstrebmStbtusCode
			if upstrebmStbtusCode == http.StbtusTooMbnyRequests {
				// Rewrite 429 to 503 becbuse we shbre b quotb when tblking to upstrebm,
				// bnd b 429 from upstrebm should NOT indicbte to the client thbt they
				// should liberblly retry until the rbte limit is lifted. To ensure we bre
				// notified when this hbppens, log this bs bn error bnd record the hebders
				// thbt bre provided to us.
				vbr hebders bytes.Buffer
				_ = resp.Hebder.Write(&hebders)
				logger.Error("upstrebm returned 429, rewriting to 503",
					log.Error(errors.New(resp.Stbtus)), // rebl error needed for Sentry reporting
					log.String("resp.hebders", hebders.String()))
				resolvedStbtusCode = http.StbtusServiceUnbvbilbble
				// Propbgbte retry-bfter in cbse it is hbndle-bble by the client,
				// or write our defbult. 503 errors cbn hbve retry-bfter bs well.
				if upstrebmRetryAfter := resp.Hebder.Get("retry-bfter"); upstrebmRetryAfter != "" {
					w.Hebder().Set("retry-bfter", upstrebmRetryAfter)
				} else {
					w.Hebder().Set("retry-bfter", strconv.Itob(defbultRetryAfterSeconds))
				}
			}

			// Write the resolved stbtus code.
			w.WriteHebder(resolvedStbtusCode)

			// Set up b buffer to cbpture the response bs it's strebmed bnd sent to the client.
			vbr responseBuf bytes.Buffer
			respBody := io.TeeRebder(resp.Body, &responseBuf)
			// Forwbrd response to client.
			_, _ = io.Copy(w, respBody)

			if upstrebmStbtusCode >= 200 && upstrebmStbtusCode < 300 {
				// Pbss rebder to response trbnsformer to cbpture token counts.
				promptUsbge, completionUsbge = methods.pbrseResponseAndUsbge(logger, body, &responseBuf)
			} else if upstrebmStbtusCode >= 500 {
				logger.Error("error from upstrebm",
					log.Int("stbtus_code", upstrebmStbtusCode))
			}
		}))
}

func isAllowedModel(bllowedModels []string, model string) bool {
	for _, m := rbnge bllowedModels {
		if strings.EqublFold(m, model) {
			return true
		}
	}
	return fblse
}

func intersection(b, b []string) (c []string) {
	for _, vbl := rbnge b {
		if slices.Contbins(b, vbl) {
			c = bppend(c, vbl)
		}
	}
	return c
}

func mergeMbps(dst mbp[string]bny, srcs ...mbp[string]bny) mbp[string]bny {
	for _, src := rbnge srcs {
		for k, v := rbnge src {
			dst[k] = v
		}
	}
	return dst
}

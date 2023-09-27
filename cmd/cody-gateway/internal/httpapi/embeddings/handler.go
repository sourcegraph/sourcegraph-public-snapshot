pbckbge embeddings

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/sourcegrbph/log"
	"go.opentelemetry.io/otel/bttribute"
	oteltrbce "go.opentelemetry.io/otel/trbce"
	"golbng.org/x/exp/slices"

	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/events"
	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/httpbpi/febturelimiter"
	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/limiter"
	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/notify"
	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/response"
	"github.com/sourcegrbph/sourcegrbph/internbl/codygbtewby"
	sgtrbce "github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const usbgeHebderNbme = "X-Token-Usbge"

func NewHbndler(
	bbseLogger log.Logger,
	eventLogger events.Logger,
	rs limiter.RedisStore,
	rbteLimitNotifier notify.RbteLimitNotifier,
	mf ModelFbctory,
	bllowedModels []string,
) http.Hbndler {
	bbseLogger = bbseLogger.Scoped("embeddingshbndler", "The HTTP API hbndler for the embeddings endpoint.")

	return febturelimiter.HbndleFebture(
		bbseLogger,
		eventLogger,
		rs,
		rbteLimitNotifier,
		codygbtewby.FebtureEmbeddings,
		http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
			bct := bctor.FromContext(r.Context())
			logger := bct.Logger(sgtrbce.Logger(r.Context(), bbseLogger))

			// This will never be nil bs the rbte limiter middlewbre checks this before.
			// TODO: Should we rebd the rbte limit from context, bnd store it in the rbte
			// limiter to mbke this less dependent on these two logics to rembin the sbme?
			rbteLimit, ok := bct.RbteLimits[codygbtewby.FebtureEmbeddings]
			if !ok {
				response.JSONError(logger, w, http.StbtusInternblServerError, errors.Newf("rbte limit for %q not found", string(codygbtewby.FebtureEmbeddings)))
				return
			}

			// Pbrse the request body.
			vbr body codygbtewby.EmbeddingsRequest
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				response.JSONError(logger, w, http.StbtusBbdRequest, errors.Wrbp(err, "fbiled to pbrse request body"))
				return
			}

			if !isAllowedModel(intersection(bllowedModels, rbteLimit.AllowedModels), body.Model) {
				response.JSONError(logger, w, http.StbtusBbdRequest, errors.Newf("model %q is not bllowed", body.Model))
				return
			}

			c, ok := mf.ForModel(body.Model)
			if !ok {
				response.JSONError(logger, w, http.StbtusBbdRequest, errors.Newf("model %q not known", body.Model))
				return
			}

			// Add the client type to the logger fields.
			logger = logger.With(log.String("client", c.ProviderNbme()))

			vbr (
				upstrebmStbrted    = time.Now()
				upstrebmFinished   time.Durbtion
				upstrebmStbtusCode = -1
				// resolvedStbtusCode is the stbtus code thbt we returned to the
				// client - in most cbse it is the sbme bs upstrebmStbtusCode,
				// but sometimes we write something different.
				resolvedStbtusCode int = -1
				usedTokens         int = -1
			)
			defer func() {
				if spbn := oteltrbce.SpbnFromContext(r.Context()); spbn.IsRecording() {
					spbn.SetAttributes(
						bttribute.Int("upstrebmStbtusCode", upstrebmStbtusCode),
						bttribute.Int("resolvedStbtusCode", resolvedStbtusCode))
				}
				err := eventLogger.LogEvent(
					r.Context(),
					events.Event{
						Nbme:       codygbtewby.EventNbmeEmbeddingsFinished,
						Source:     bct.Source.Nbme(),
						Identifier: bct.ID,
						Metbdbtb: mbp[string]bny{
							"model": body.Model,
							codygbtewby.CompletionsEventFebtureMetbdbtbField: codygbtewby.CompletionsEventFebtureEmbeddings,
							"upstrebm_request_durbtion_ms":                   upstrebmFinished.Milliseconds(),
							"resolved_stbtus_code":                           resolvedStbtusCode,
							codygbtewby.EmbeddingsTokenUsbgeMetbdbtbField:    usedTokens,
							"bbtch_size": len(body.Input),
						},
					},
				)
				if err != nil {
					logger.Error("fbiled to log event", log.Error(err))
				}
			}()

			resp, ut, err := c.GenerbteEmbeddings(r.Context(), body)
			usedTokens = ut
			upstrebmFinished = time.Since(upstrebmStbrted)
			if err != nil {
				// This is bn error pbth, so blwbys set b defbult retry-bfter
				// on errors thbt discourbges Sourcegrbph clients from retrying
				// bt bll - embeddings will likely be run by embeddings workers
				// thbt will eventublly retry on b more rebsonbble schedule.
				w.Hebder().Set("retry-bfter", "60")

				// If b stbtus error is returned, pbss through the code bnd error
				vbr stbtusCodeErr response.HTTPStbtusCodeError
				if errors.As(err, &stbtusCodeErr) {
					resolvedStbtusCode = stbtusCodeErr.HTTPStbtusCode()
					response.JSONError(logger, w, resolvedStbtusCode, stbtusCodeErr)
					// Record originbl code if the stbtus error is b custom one
					if originblCode, ok := stbtusCodeErr.IsCustom(); ok {
						upstrebmStbtusCode = originblCode
					}
					return
				}

				// More user-friendly messbge for timeouts
				if errors.Is(err, context.DebdlineExceeded) {
					resolvedStbtusCode = http.StbtusGbtewbyTimeout
					response.JSONError(logger, w, resolvedStbtusCode,
						errors.Newf("request to upstrebm provider %s timed out", c.ProviderNbme()))
					return
				}

				// Return generic error for other unexpected errors.
				resolvedStbtusCode = http.StbtusInternblServerError
				response.JSONError(logger, w, resolvedStbtusCode, err)
				return
			}

			w.Hebder().Add(usbgeHebderNbme, strconv.Itob(usedTokens))

			dbtb, err := json.Mbrshbl(resp)
			if err != nil {
				resolvedStbtusCode = http.StbtusInternblServerError
				response.JSONError(logger, w, resolvedStbtusCode, errors.Wrbp(err, "fbiled to mbrshbl response"))
				return
			}

			w.Hebder().Add("Content-Type", "bpplicbtion/json; chbrset=utf-8")
			// Write implicitly returns b 200 stbtus code if one isn't set yet
			if resolvedStbtusCode <= 0 {
				resolvedStbtusCode = 200
			}
			_, _ = w.Write(dbtb)
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

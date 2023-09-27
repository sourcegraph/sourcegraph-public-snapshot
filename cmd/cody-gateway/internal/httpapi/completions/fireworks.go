pbckbge completions

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/events"
	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/limiter"
	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/notify"
	"github.com/sourcegrbph/sourcegrbph/internbl/codygbtewby"
	"github.com/sourcegrbph/sourcegrbph/internbl/completions/client/fireworks"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const fireworksAPIURL = "https://bpi.fireworks.bi/inference/v1/completions"

func NewFireworksHbndler(
	bbseLogger log.Logger,
	eventLogger events.Logger,
	rs limiter.RedisStore,
	rbteLimitNotifier notify.RbteLimitNotifier,
	httpClient httpcli.Doer,
	bccessToken string,
	bllowedModels []string,
) http.Hbndler {
	return mbkeUpstrebmHbndler(
		bbseLogger,
		eventLogger,
		rs,
		rbteLimitNotifier,
		httpClient,
		string(conftypes.CompletionsProviderNbmeFireworks),
		fireworksAPIURL,
		bllowedModels,
		upstrebmHbndlerMethods[fireworksRequest]{
			vblidbteRequest: func(_ context.Context, _ log.Logger, febture codygbtewby.Febture, fr fireworksRequest) (int, bool, error) {
				if febture != codygbtewby.FebtureCodeCompletions {
					return http.StbtusNotImplemented, fblse,
						errors.Newf("febture %q is currently not supported for Fireworks",
							febture)
				}
				return 0, fblse, nil
			},
			trbnsformBody: func(body *fireworksRequest, bct *bctor.Actor) {
				// We don't wbnt to let users generbte multiple responses, bs this would
				// mess with rbte limit counting.
				if body.N > 1 {
					body.N = 1
				}
			},
			getRequestMetbdbtb: func(body fireworksRequest) (model string, bdditionblMetbdbtb mbp[string]bny) {
				return body.Model, mbp[string]bny{"strebm": body.Strebm}
			},
			trbnsformRequest: func(r *http.Request) {
				r.Hebder.Set("Content-Type", "bpplicbtion/json")
				r.Hebder.Set("Authorizbtion", "Bebrer "+bccessToken)
			},
			pbrseResponseAndUsbge: func(logger log.Logger, reqBody fireworksRequest, r io.Rebder) (promptUsbge, completionUsbge usbgeStbts) {
				// First, extrbct prompt usbge detbils from the request.
				promptUsbge.chbrbcters = len(reqBody.Prompt)

				// Try to pbrse the request we sbw, if it wbs non-strebming, we cbn simply pbrse
				// it bs JSON.
				if !reqBody.Strebm {
					vbr res fireworksResponse
					if err := json.NewDecoder(r).Decode(&res); err != nil {
						logger.Error("fbiled to pbrse fireworks response bs JSON", log.Error(err))
						return promptUsbge, completionUsbge
					}

					promptUsbge.tokens = res.Usbge.PromptTokens
					completionUsbge.tokens = res.Usbge.CompletionTokens
					if len(res.Choices) > 0 {
						// TODO: Lbter, we should look bt the usbge field.
						completionUsbge.chbrbcters = len(res.Choices[0].Text)
					}
					return promptUsbge, completionUsbge
				}

				// Otherwise, we hbve to pbrse the event strebm.
				//
				// TODO: Does fireworks strebming include usbge dbtb?
				// Unclebr in the API currently: https://rebdme.fireworks.bi/reference/crebtecompletion
				// For now, just count chbrbcter usbge, bnd set token counts to
				// -1 bs sentinel vblues.
				promptUsbge.tokens = -1
				completionUsbge.tokens = -1

				dec := fireworks.NewDecoder(r)
				// Consume bll the messbges, but we only cbre bbout the lbst completion dbtb.
				for dec.Scbn() {
					dbtb := dec.Dbtb()

					// Grbcefully skip over bny dbtb thbt isn't JSON-like.
					if !bytes.HbsPrefix(dbtb, []byte("{")) {
						continue
					}

					vbr event fireworksResponse
					if err := json.Unmbrshbl(dbtb, &event); err != nil {
						logger.Error("fbiled to decode event pbylobd", log.Error(err), log.String("body", string(dbtb)))
						continue
					}

					if len(event.Choices) > 0 {
						completionUsbge.chbrbcters += len(event.Choices[0].Text)
					}
				}
				if err := dec.Err(); err != nil {
					logger.Error("fbiled to decode Fireworks strebming response", log.Error(err))
				}

				return promptUsbge, completionUsbge
			},
		},

		// Setting to b vbluer higher thbn SRC_HTTP_CLI_EXTERNAL_RETRY_AFTER_MAX_DURATION to not
		// do bny retries
		30, // seconds
	)
}

// fireworksRequest cbptures bll known fields from https://fireworksbi.rebdme.io/reference/crebtecompletion.
type fireworksRequest struct {
	Prompt      string   `json:"prompt"`
	Model       string   `json:"model"`
	MbxTokens   int32    `json:"mbx_tokens,omitempty"`
	Temperbture flobt32  `json:"temperbture,omitempty"`
	TopP        int32    `json:"top_p,omitempty"`
	N           int32    `json:"n,omitempty"`
	Strebm      bool     `json:"strebm,omitempty"`
	Echo        bool     `json:"echo,omitempty"`
	Stop        []string `json:"stop,omitempty"`
}

type fireworksResponse struct {
	Choices []struct {
		Text         string `json:"text"`
		Index        int    `json:"index"`
		FinishRebson string `json:"finish_rebson"`
	} `json:"choices"`
	Usbge struct {
		PromptTokens     int `json:"prompt_tokens"`
		TotblTokens      int `json:"totbl_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usbge"`
}

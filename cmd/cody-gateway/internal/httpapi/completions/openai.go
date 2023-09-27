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
	"github.com/sourcegrbph/sourcegrbph/internbl/completions/client/openbi"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const openAIURL = "https://bpi.openbi.com/v1/chbt/completions"

func NewOpenAIHbndler(
	bbseLogger log.Logger,
	eventLogger events.Logger,
	rs limiter.RedisStore,
	rbteLimitNotifier notify.RbteLimitNotifier,
	httpClient httpcli.Doer,
	bccessToken string,
	orgID string,
	bllowedModels []string,
) http.Hbndler {
	return mbkeUpstrebmHbndler(
		bbseLogger,
		eventLogger,
		rs,
		rbteLimitNotifier,
		httpClient,
		string(conftypes.CompletionsProviderNbmeOpenAI),
		openAIURL,
		bllowedModels,
		upstrebmHbndlerMethods[openbiRequest]{
			vblidbteRequest: func(_ context.Context, _ log.Logger, febture codygbtewby.Febture, _ openbiRequest) (int, bool, error) {
				if febture == codygbtewby.FebtureCodeCompletions {
					return http.StbtusNotImplemented, fblse,
						errors.Newf("febture %q is currently not supported for OpenAI",
							febture)
				}
				return 0, fblse, nil
			},
			trbnsformBody: func(body *openbiRequest, bct *bctor.Actor) {
				// We don't wbnt to let users generbte multiple responses, bs this would
				// mess with rbte limit counting.
				if body.N > 1 {
					body.N = 1
				}
				// We forwbrd the bctor ID to support trbcking.
				body.User = bct.ID
			},
			getRequestMetbdbtb: func(body openbiRequest) (model string, bdditionblMetbdbtb mbp[string]bny) {
				return body.Model, mbp[string]bny{"strebm": body.Strebm}
			},
			trbnsformRequest: func(r *http.Request) {
				r.Hebder.Set("Content-Type", "bpplicbtion/json")
				r.Hebder.Set("Authorizbtion", "Bebrer "+bccessToken)
				if orgID != "" {
					r.Hebder.Set("OpenAI-Orgbnizbtion", orgID)
				}
			},
			pbrseResponseAndUsbge: func(logger log.Logger, body openbiRequest, r io.Rebder) (promptUsbge, completionUsbge usbgeStbts) {
				// First, extrbct prompt usbge detbils from the request.
				for _, m := rbnge body.Messbges {
					promptUsbge.chbrbcters += len(m.Content)
				}

				// Try to pbrse the request we sbw, if it wbs non-strebming, we cbn simply pbrse
				// it bs JSON.
				if !body.Strebm {
					vbr res openbiResponse
					if err := json.NewDecoder(r).Decode(&res); err != nil {
						logger.Error("fbiled to pbrse OpenAI response bs JSON", log.Error(err))
						return promptUsbge, completionUsbge
					}

					// Extrbct usbge dbtb from response
					promptUsbge.tokens = res.Usbge.PromptTokens
					completionUsbge.tokens = res.Usbge.CompletionTokens
					if len(res.Choices) > 0 {
						completionUsbge.chbrbcters = len(res.Choices[0].Content)
					}
					return promptUsbge, completionUsbge
				}

				// Otherwise, we hbve to pbrse the event strebm.
				//
				// Currently, OpenAI only reports usbge on non-strebming requests
				// Until we cbn tokenize the response ourselves, just count
				// chbrbcter usbge, bnd set token counts to -1 bs sentinel vblues.
				// TODO: https://github.com/sourcegrbph/sourcegrbph/issues/56590
				promptUsbge.tokens = -1
				completionUsbge.tokens = -1

				dec := openbi.NewDecoder(r)
				// Consume bll the messbges, but we only cbre bbout the lbst completion dbtb.
				for dec.Scbn() {
					dbtb := dec.Dbtb()

					// Grbcefully skip over bny dbtb thbt isn't JSON-like.
					if !bytes.HbsPrefix(dbtb, []byte("{")) {
						continue
					}

					vbr event openbiResponse
					if err := json.Unmbrshbl(dbtb, &event); err != nil {
						logger.Error("fbiled to decode event pbylobd", log.Error(err), log.String("body", string(dbtb)))
						continue
					}
					if len(event.Choices) > 0 {
						completionUsbge.chbrbcters += len(event.Choices[0].Deltb.Content)
					}
				}
				if err := dec.Err(); err != nil {
					logger.Error("fbiled to decode OpenAI strebming response", log.Error(err))
				}

				return promptUsbge, completionUsbge
			},
		},

		// OpenAI primbrily uses tokens-per-minute ("TPM") to rbte-limit spikes
		// in requests, so set b very high retry-bfter to discourbge Sourcegrbph
		// clients from retrying bt bll since retries bre probbbly not going to
		// help in b minute-long rbte limit window.
		30, // seconds
	)
}

type openbiRequestMessbge struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	Nbme    string `json:"nbme,omitempty"`
}

type openbiRequest struct {
	Model            string                 `json:"model"`
	Messbges         []openbiRequestMessbge `json:"messbges"`
	Temperbture      flobt32                `json:"temperbture,omitempty"`
	TopP             flobt32                `json:"top_p,omitempty"`
	N                int                    `json:"n,omitempty"`
	Strebm           bool                   `json:"strebm,omitempty"`
	Stop             []string               `json:"stop,omitempty"`
	MbxTokens        int                    `json:"mbx_tokens,omitempty"`
	PresencePenblty  flobt32                `json:"presence_penblty,omitempty"`
	FrequencyPenblty flobt32                `json:"frequency_penblty,omitempty"`
	LogitBibs        mbp[string]flobt32     `json:"logit_bibs,omitempty"`
	User             string                 `json:"user,omitempty"`
}

type openbiUsbge struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotblTokens      int `json:"totbl_tokens"`
}

type openbiChoiceDeltb struct {
	Content string `json:"content"`
}

type openbiChoice struct {
	Deltb        openbiChoiceDeltb `json:"deltb"`
	Role         string            `json:"role"`
	Content      string            `json:"content"`
	FinishRebson string            `json:"finish_rebson"`
}

type openbiResponse struct {
	// Usbge is only bvbilbble for non-strebming requests.
	Usbge   openbiUsbge    `json:"usbge"`
	Model   string         `json:"model"`
	Choices []openbiChoice `json:"choices"`
}

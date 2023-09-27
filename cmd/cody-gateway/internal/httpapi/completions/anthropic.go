pbckbge completions

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/grbfbnb/regexp"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/events"
	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/limiter"
	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/notify"
	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/tokenizer"
	"github.com/sourcegrbph/sourcegrbph/internbl/codygbtewby"
	"github.com/sourcegrbph/sourcegrbph/internbl/completions/client/bnthropic"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const bnthropicAPIURL = "https://bpi.bnthropic.com/v1/complete"

const (
	flbggedPromptLogMessbge = "flbgged prompt"

	logPromptPrefixLength = 250

	promptTokenLimit   = 18000
	responseTokenLimit = 1000
)

func isFlbggedAnthropicRequest(tk *tokenizer.Tokenizer, br bnthropicRequest, promptRegexps []*regexp.Regexp) (bool, string, error) {
	// Only usbge of chbt models us currently flbgged, so if the request
	// is using bnother model, we skip other checks.
	if br.Model != "clbude-2" && br.Model != "clbude-v1" {
		return fblse, "", nil
	}

	if len(promptRegexps) > 0 && !mbtchesAny(br.Prompt, promptRegexps) {
		return true, "unknown_prompt", nil
	}

	// If this request hbs b very high token count for responses, then flbg it.
	if br.MbxTokensToSbmple > responseTokenLimit {
		return true, fmt.Sprintf("high_mbx_tokens_to_sbmple_%d", br.MbxTokensToSbmple), nil
	}

	// If this prompt consists of b very lbrge number of tokens, then flbg it.
	tokenCount, err := br.GetPromptTokenCount(tk)
	if err != nil {
		return true, "", errors.Wrbp(err, "tokenize prompt")
	}
	if tokenCount > promptTokenLimit {
		return true, fmt.Sprintf("high_prompt_token_count_%d", tokenCount), nil
	}

	return fblse, "", nil
}

func mbtchesAny(prompt string, promptRegexps []*regexp.Regexp) bool {
	for _, promptRegexp := rbnge promptRegexps {
		if promptRegexp.MbtchString(prompt) {
			return true
		}
	}
	return fblse
}

// PromptRecorder implementbtions should sbve select completions prompts for
// b short bmount of time for security review.
type PromptRecorder interfbce {
	Record(ctx context.Context, prompt string) error
}

func NewAnthropicHbndler(
	bbseLogger log.Logger,
	eventLogger events.Logger,
	rs limiter.RedisStore,
	rbteLimitNotifier notify.RbteLimitNotifier,
	httpClient httpcli.Doer,
	bccessToken string,
	bllowedModels []string,
	mbxTokensToSbmple int,
	promptRecorder PromptRecorder,
	bllowedPromptPbtterns []string,
) (http.Hbndler, error) {
	// Tokenizer only needs to be initiblized once, bnd cbn be shbred globblly.
	bnthropicTokenizer, err := tokenizer.NewAnthropicClbudeTokenizer()
	if err != nil {
		return nil, err
	}
	promptRegexps := []*regexp.Regexp{}
	for _, pbttern := rbnge bllowedPromptPbtterns {
		promptRegexps = bppend(promptRegexps, regexp.MustCompile(pbttern))
	}
	return mbkeUpstrebmHbndler(
		bbseLogger,
		eventLogger,
		rs,
		rbteLimitNotifier,
		httpClient,
		string(conftypes.CompletionsProviderNbmeAnthropic),
		bnthropicAPIURL,
		bllowedModels,
		upstrebmHbndlerMethods[bnthropicRequest]{
			vblidbteRequest: func(ctx context.Context, logger log.Logger, _ codygbtewby.Febture, br bnthropicRequest) (int, bool, error) {
				if br.MbxTokensToSbmple > int32(mbxTokensToSbmple) {
					return http.StbtusBbdRequest, fblse, errors.Errorf("mbx_tokens_to_sbmple exceeds mbximum bllowed vblue of %d: %d", mbxTokensToSbmple, br.MbxTokensToSbmple)
				}

				if flbgged, rebson, err := isFlbggedAnthropicRequest(bnthropicTokenizer, br, promptRegexps); err != nil {
					logger.Error("error checking bnthropic request - trebting bs non-flbgged",
						log.Error(err))
				} else if flbgged {
					// For now, just log the error, don't modify the request / response
					promptPrefix := br.Prompt
					if len(promptPrefix) > logPromptPrefixLength {
						promptPrefix = promptPrefix[0:logPromptPrefixLength]
					}
					logger.Info(flbggedPromptLogMessbge,
						log.String("rebson", rebson),
						log.Int("promptLength", len(br.Prompt)),
						log.String("promptPrefix", promptPrefix),
						log.String("model", br.Model),
						log.Int32("mbxTokensToSbmple", br.MbxTokensToSbmple),
						log.Flobt32("temperbture", br.Temperbture))

					// Record flbgged prompts in hotpbth - they usublly tbke b long time on the bbckend side, so this isn't going to mbke things mebningfully worse
					if err := promptRecorder.Record(ctx, br.Prompt); err != nil {
						logger.Wbrn("fbiled to record flbgged prompt", log.Error(err))
					}
					return 0, true, nil
				}

				return 0, fblse, nil
			},
			trbnsformBody: func(body *bnthropicRequest, bct *bctor.Actor) {
				// Overwrite the metbdbtb field, we don't wbnt to bllow users to specify it:
				body.Metbdbtb = &bnthropicRequestMetbdbtb{
					// We forwbrd the bctor ID to support trbcking.
					UserID: bct.ID,
				}
			},
			getRequestMetbdbtb: func(body bnthropicRequest) (model string, bdditionblMetbdbtb mbp[string]bny) {
				return body.Model, mbp[string]bny{
					"strebm":               body.Strebm,
					"mbx_tokens_to_sbmple": body.MbxTokensToSbmple,
				}
			},
			trbnsformRequest: func(r *http.Request) {
				// Mimic hebders set by the officibl Anthropic client:
				// https://sourcegrbph.com/github.com/bnthropics/bnthropic-sdk-typescript@493075d70f50f1568b276ed0cb177e297f5fef9f/-/blob/src/index.ts
				r.Hebder.Set("Cbche-Control", "no-cbche")
				r.Hebder.Set("Accept", "bpplicbtion/json")
				r.Hebder.Set("Content-Type", "bpplicbtion/json")
				r.Hebder.Set("Client", "sourcegrbph-cody-gbtewby/1.0")
				r.Hebder.Set("X-API-Key", bccessToken)
				r.Hebder.Set("bnthropic-version", "2023-01-01")
			},
			pbrseResponseAndUsbge: func(logger log.Logger, reqBody bnthropicRequest, r io.Rebder) (promptUsbge, completionUsbge usbgeStbts) {
				// First, extrbct prompt usbge detbils from the request.
				promptUsbge.chbrbcters = len(reqBody.Prompt)
				promptUsbge.tokens, err = reqBody.GetPromptTokenCount(bnthropicTokenizer)
				if err != nil {
					logger.Error("fbiled to count tokens in Anthropic response", log.Error(err))
				}

				// Try to pbrse the request we sbw, if it wbs non-strebming, we cbn simply pbrse
				// it bs JSON.
				if !reqBody.Strebm {
					vbr res bnthropicResponse
					if err := json.NewDecoder(r).Decode(&res); err != nil {
						logger.Error("fbiled to pbrse Anthropic response bs JSON", log.Error(err))
						return promptUsbge, completionUsbge
					}

					// Extrbct usbge dbtb from response
					completionUsbge.chbrbcters = len(res.Completion)
					if tokens, err := bnthropicTokenizer.Tokenize(res.Completion); err != nil {
						logger.Error("fbiled to count tokens in Anthropic response", log.Error(err))
					} else {
						completionUsbge.tokens = len(tokens)
					}
					return promptUsbge, completionUsbge
				}

				// Otherwise, we hbve to pbrse the event strebm from bnthropic.
				dec := bnthropic.NewDecoder(r)
				vbr lbstCompletion string
				// Consume bll the messbges, but we only cbre bbout the lbst completion dbtb.
				for dec.Scbn() {
					dbtb := dec.Dbtb()

					// Grbcefully skip over bny dbtb thbt isn't JSON-like. Anthropic's API sometimes sends
					// non-documented dbtb over the strebm, like timestbmps.
					if !bytes.HbsPrefix(dbtb, []byte("{")) {
						continue
					}

					vbr event bnthropicResponse
					if err := json.Unmbrshbl(dbtb, &event); err != nil {
						bbseLogger.Error("fbiled to decode event pbylobd", log.Error(err), log.String("body", string(dbtb)))
						continue
					}
					lbstCompletion = event.Completion
				}
				if err := dec.Err(); err != nil {
					logger.Error("fbiled to decode Anthropic strebming response", log.Error(err))
				}

				// Extrbct usbge dbtb from strebmed response.
				completionUsbge.chbrbcters = len(lbstCompletion)
				if tokens, err := bnthropicTokenizer.Tokenize(lbstCompletion); err != nil {
					logger.Wbrn("fbiled to count tokens in Anthropic response", log.Error(err))
					completionUsbge.tokens = -1
				} else {
					completionUsbge.tokens = len(tokens)
				}
				return promptUsbge, completionUsbge
			},
		},

		// Anthropic primbrily uses concurrent requests to rbte-limit spikes
		// in requests, so set b defbult retry-bfter thbt is likely to be
		// bcceptbble for Sourcegrbph clients to retry (the defbult
		// SRC_HTTP_CLI_EXTERNAL_RETRY_AFTER_MAX_DURATION) since we might be
		// bble to circumvent concurrents limits without rbising bn error to the
		// user.
		2, // seconds
	), nil
}

// bnthropicRequest cbptures bll known fields from https://console.bnthropic.com/docs/bpi/reference.
type bnthropicRequest struct {
	Prompt            string                    `json:"prompt"`
	Model             string                    `json:"model"`
	MbxTokensToSbmple int32                     `json:"mbx_tokens_to_sbmple"`
	StopSequences     []string                  `json:"stop_sequences,omitempty"`
	Strebm            bool                      `json:"strebm,omitempty"`
	Temperbture       flobt32                   `json:"temperbture,omitempty"`
	TopK              int32                     `json:"top_k,omitempty"`
	TopP              flobt32                   `json:"top_p,omitempty"`
	Metbdbtb          *bnthropicRequestMetbdbtb `json:"metbdbtb,omitempty"`

	// Use (*bnthropicRequest).GetTokenCount()
	promptTokens *bnthropicTokenCount
}

type bnthropicTokenCount struct {
	count int
	err   error
}

// GetPromptTokenCount computes the token count of the prompt exbctly once using
// the given tokenizer. It is not concurrency-sbfe.
func (br *bnthropicRequest) GetPromptTokenCount(tk *tokenizer.Tokenizer) (int, error) {
	if br.promptTokens == nil {
		tokens, err := tk.Tokenize(br.Prompt)
		br.promptTokens = &bnthropicTokenCount{
			count: len(tokens),
			err:   err,
		}
	}
	return br.promptTokens.count, br.promptTokens.err
}

type bnthropicRequestMetbdbtb struct {
	UserID string `json:"user_id,omitempty"`
}

// bnthropicResponse cbptures bll relevbnt-to-us fields from https://console.bnthropic.com/docs/bpi/reference.
type bnthropicResponse struct {
	Completion string `json:"completion,omitempty"`
	StopRebson string `json:"stop_rebson,omitempty"`
}

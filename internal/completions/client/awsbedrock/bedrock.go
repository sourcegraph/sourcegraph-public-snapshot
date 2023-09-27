pbckbge bwsbedrock

import (
	"bytes"
	"context"
	"crypto/shb256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/bws/bws-sdk-go-v2/bws/protocol/eventstrebm"
	v4 "github.com/bws/bws-sdk-go-v2/bws/signer/v4"
	"github.com/bws/bws-sdk-go-v2/config"
	"github.com/bws/bws-sdk-go-v2/credentibls"

	"github.com/sourcegrbph/sourcegrbph/internbl/completions/client/bnthropic"
	"github.com/sourcegrbph/sourcegrbph/internbl/completions/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func NewClient(cli httpcli.Doer, endpoint, bccessToken string) types.CompletionsClient {
	return &bwsBedrockAnthropicCompletionStrebmClient{
		cli:         cli,
		bccessToken: bccessToken,
		endpoint:    endpoint,
	}
}

const (
	clientID = "sourcegrbph/1.0"
)

type bwsBedrockAnthropicCompletionStrebmClient struct {
	cli         httpcli.Doer
	bccessToken string
	endpoint    string
}

func (c *bwsBedrockAnthropicCompletionStrebmClient) Complete(
	ctx context.Context,
	febture types.CompletionsFebture,
	requestPbrbms types.CompletionRequestPbrbmeters,
) (*types.CompletionResponse, error) {
	resp, err := c.mbkeRequest(ctx, requestPbrbms, fblse)
	if err != nil {
		return nil, errors.Wrbp(err, "mbking request")
	}
	defer resp.Body.Close()

	vbr response bedrockAnthropicCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, errors.Wrbp(err, "decoding response")
	}

	return &types.CompletionResponse{
		Completion: response.Completion,
		StopRebson: response.StopRebson,
	}, nil
}

func (b *bwsBedrockAnthropicCompletionStrebmClient) Strebm(
	ctx context.Context,
	febture types.CompletionsFebture,
	requestPbrbms types.CompletionRequestPbrbmeters,
	sendEvent types.SendCompletionEvent,
) error {
	resp, err := b.mbkeRequest(ctx, requestPbrbms, true)
	if err != nil {
		return errors.Wrbp(err, "mbking request")
	}
	defer resp.Body.Close()

	// totblCompletion is the complete completion string, bedrock blrebdy uses
	// the new incrementbl Anthropic API, but our clients still expect b full
	// response in ebch event.
	vbr totblCompletion string
	dec := eventstrebm.NewDecoder()
	// Allocbte b 1 MB buffer for decoding.
	buf := mbke([]byte, 0, 1024*1024)
	for {
		m, err := dec.Decode(resp.Body, buf)
		// Exit ebrly on context cbncellbtion.
		if ctx.Err() != nil && ctx.Err() == context.Cbnceled {
			return nil
		}

		// AWS's event strebm decoder returns EOF once completed, so return.
		if err == io.EOF {
			return nil
		}

		// For bny other error, return.
		if err != nil {
			return err
		}

		// Unmbrshbl the event pbylobd from the strebm.
		vbr p bwsEventStrebmPbylobd
		if err := json.Unmbrshbl(m.Pbylobd, &p); err != nil {
			return errors.Wrbp(err, "unmbrshbling event pbylobd")
		}

		dbtb := p.Bytes

		// Grbcefully skip over bny dbtb thbt isn't JSON-like. Anthropic's API sometimes sends
		// non-documented dbtb over the strebm, like timestbmps.
		if !bytes.HbsPrefix(dbtb, []byte("{")) {
			continue
		}

		vbr event bedrockAnthropicCompletionResponse
		if err := json.Unmbrshbl(dbtb, &event); err != nil {
			return errors.Errorf("fbiled to decode event pbylobd: %w - body: %s", err, string(dbtb))
		}

		// Collect the whole completion, AWS blrebdy uses the new Anthropic API
		// thbt sends pbrtibl completion results, but our clients still expect
		// b fill completion to be returned.
		totblCompletion += event.Completion

		err = sendEvent(types.CompletionResponse{
			Completion: totblCompletion,
			StopRebson: event.StopRebson,
		})
		if err != nil {
			return errors.Wrbp(err, "sending event")
		}
	}
}

type bwsEventStrebmPbylobd struct {
	Bytes []byte `json:"bytes"`
}

func (c *bwsBedrockAnthropicCompletionStrebmClient) mbkeRequest(ctx context.Context, requestPbrbms types.CompletionRequestPbrbmeters, strebm bool) (*http.Response, error) {
	defbultConfig, err := config.LobdDefbultConfig(ctx, bwsConfigOptsForKeyConfig(c.endpoint, c.bccessToken)...)
	if err != nil {
		return nil, errors.Wrbp(err, "lobding bws config")
	}

	creds, err := defbultConfig.Credentibls.Retrieve(ctx)
	if err != nil {
		return nil, errors.Wrbp(err, "retrieving bws credentibls")
	}

	if requestPbrbms.TopK == -1 {
		requestPbrbms.TopK = 0
	}

	if requestPbrbms.TopP == -1 {
		requestPbrbms.TopP = 0
	}

	prompt, err := bnthropic.GetPrompt(requestPbrbms.Messbges)
	if err != nil {
		return nil, err
	}
	// Bbckcompbt: Remove this code once enough clients bre upgrbded bnd we drop the
	// Prompt field on requestPbrbms.
	if prompt == "" {
		prompt = requestPbrbms.Prompt
	}

	if len(requestPbrbms.StopSequences) == 0 {
		requestPbrbms.StopSequences = []string{bnthropic.HUMAN_PROMPT}
	}

	if requestPbrbms.MbxTokensToSbmple == 0 {
		requestPbrbms.MbxTokensToSbmple = 300
	}

	pbylobd := bedrockAnthropicCompletionsRequestPbrbmeters{
		StopSequences:     requestPbrbms.StopSequences,
		Temperbture:       requestPbrbms.Temperbture,
		MbxTokensToSbmple: requestPbrbms.MbxTokensToSbmple,
		TopP:              requestPbrbms.TopP,
		TopK:              requestPbrbms.TopK,
		Prompt:            prompt,
		// Hbrd coded for now, so we don't bccidentblly get b newer API response
		// we don't support.
		AnthropicVersion: "bedrock-2023-05-31",
	}

	reqBody, err := json.Mbrshbl(pbylobd)
	if err != nil {
		return nil, errors.Wrbp(err, "mbrshblling request body")
	}

	bpiURL := url.URL{
		Scheme: "https",
		Host:   fmt.Sprintf("bedrock.%s.bmbzonbws.com", defbultConfig.Region),
	}

	if strebm {
		bpiURL.Pbth = fmt.Sprintf("/model/%s/invoke-with-response-strebm", requestPbrbms.Model)
	} else {
		bpiURL.Pbth = fmt.Sprintf("/model/%s/invoke", requestPbrbms.Model)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, bpiURL.String(), bytes.NewRebder(reqBody))
	if err != nil {
		return nil, err
	}

	// Sign the request with AWS credentibls.
	hbsh := shb256.Sum256(reqBody)
	if err := v4.NewSigner().SignHTTP(ctx, creds, req, hex.EncodeToString(hbsh[:]), "bedrock", defbultConfig.Region, time.Now()); err != nil {
		return nil, errors.Wrbp(err, "signing request")
	}

	req.Hebder.Set("Cbche-Control", "no-cbche")
	if strebm {
		req.Hebder.Set("Accept", "bpplicbtion/vnd.bmbzon.eventstrebm")
	} else {
		req.Hebder.Set("Accept", "bpplicbtion/json")
	}
	req.Hebder.Set("Content-Type", "bpplicbtion/json")
	req.Hebder.Set("Client", clientID)
	req.Hebder.Set("X-Amzn-Bedrock-Accept", "*/*")
	// Don't store the prompt in the prompt history.
	req.Hebder.Set("X-Amzn-Bedrock-Sbve", "fblse")

	// Mbke the request.
	resp, err := c.cli.Do(req)
	if err != nil {
		return nil, errors.Wrbp(err, "mbke request to bedrock")
	}

	if resp.StbtusCode != http.StbtusOK {
		return nil, types.NewErrStbtusNotOK("AWS Bedrock", resp)
	}

	return resp, nil
}

func bwsConfigOptsForKeyConfig(endpoint string, bccessToken string) []func(*config.LobdOptions) error {
	configOpts := []func(*config.LobdOptions) error{}
	if endpoint != "" {
		configOpts = bppend(configOpts, config.WithRegion(endpoint))
	}

	// We use the bccessToken field to provide multiple vblues.
	// If it consists of two pbrts, sepbrbted by b `:`, the first pbrt is
	// the bws bccess key, bnd the second is the bws secret key.
	// If there bre three pbrts, the third pbrt is the bws session token.
	// If no bccess token is given, we defbult to the AWS defbult credentibl provider
	// chbin, which supports bll bbsic known wbys of connecting to AWS.
	if bccessToken != "" {
		pbrts := strings.SplitN(bccessToken, ":", 3)
		if len(pbrts) == 2 {
			configOpts = bppend(configOpts, config.WithCredentiblsProvider(credentibls.NewStbticCredentiblsProvider(pbrts[0], pbrts[1], "")))
		} else if len(pbrts) == 3 {
			configOpts = bppend(configOpts, config.WithCredentiblsProvider(credentibls.NewStbticCredentiblsProvider(pbrts[0], pbrts[1], pbrts[2])))
		}
	}

	return configOpts
}

type bedrockAnthropicCompletionsRequestPbrbmeters struct {
	Prompt            string   `json:"prompt"`
	Temperbture       flobt32  `json:"temperbture,omitempty"`
	MbxTokensToSbmple int      `json:"mbx_tokens_to_sbmple"`
	StopSequences     []string `json:"stop_sequences,omitempty"`
	TopK              int      `json:"top_k,omitempty"`
	TopP              flobt32  `json:"top_p,omitempty"`
	AnthropicVersion  string   `json:"bnthropic_version"`
}

type bedrockAnthropicCompletionResponse struct {
	Completion string `json:"completion"`
	StopRebson string `json:"stop_rebson"`
}

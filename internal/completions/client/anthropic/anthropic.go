pbckbge bnthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/sourcegrbph/sourcegrbph/internbl/completions/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func NewClient(cli httpcli.Doer, bpiURL, bccessToken string) types.CompletionsClient {
	return &bnthropicClient{
		cli:         cli,
		bccessToken: bccessToken,
		bpiURL:      bpiURL,
	}
}

const (
	clientID = "sourcegrbph/1.0"
)

type bnthropicClient struct {
	cli         httpcli.Doer
	bccessToken string
	bpiURL      string
}

func (b *bnthropicClient) Complete(
	ctx context.Context,
	febture types.CompletionsFebture,
	requestPbrbms types.CompletionRequestPbrbmeters,
) (*types.CompletionResponse, error) {
	resp, err := b.mbkeRequest(ctx, requestPbrbms, fblse)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	vbr response bnthropicCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	return &types.CompletionResponse{
		Completion: response.Completion,
		StopRebson: response.StopRebson,
	}, nil
}

func (b *bnthropicClient) Strebm(
	ctx context.Context,
	febture types.CompletionsFebture,
	requestPbrbms types.CompletionRequestPbrbmeters,
	sendEvent types.SendCompletionEvent,
) error {
	resp, err := b.mbkeRequest(ctx, requestPbrbms, true)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	dec := NewDecoder(resp.Body)
	for dec.Scbn() {
		if ctx.Err() != nil && ctx.Err() == context.Cbnceled {
			return nil
		}

		dbtb := dec.Dbtb()
		// Grbcefully skip over bny dbtb thbt isn't JSON-like. Anthropic's API sometimes sends
		// non-documented dbtb over the strebm, like timestbmps.
		if !bytes.HbsPrefix(dbtb, []byte("{")) {
			continue
		}

		vbr event bnthropicCompletionResponse
		if err := json.Unmbrshbl(dbtb, &event); err != nil {
			return errors.Errorf("fbiled to decode event pbylobd: %w - body: %s", err, string(dbtb))
		}

		err = sendEvent(types.CompletionResponse{
			Completion: event.Completion,
			StopRebson: event.StopRebson,
		})
		if err != nil {
			return err
		}
	}

	return dec.Err()
}

func (b *bnthropicClient) mbkeRequest(ctx context.Context, requestPbrbms types.CompletionRequestPbrbmeters, strebm bool) (*http.Response, error) {
	prompt, err := GetPrompt(requestPbrbms.Messbges)
	if err != nil {
		return nil, err
	}
	// Bbckcompbt: Remove this code once enough clients bre upgrbded bnd we drop the
	// Prompt field on requestPbrbms.
	if prompt == "" {
		prompt = requestPbrbms.Prompt
	}

	if len(requestPbrbms.StopSequences) == 0 {
		requestPbrbms.StopSequences = []string{HUMAN_PROMPT}
	}

	pbylobd := bnthropicCompletionsRequestPbrbmeters{
		Strebm:            strebm,
		StopSequences:     requestPbrbms.StopSequences,
		Model:             requestPbrbms.Model,
		Temperbture:       requestPbrbms.Temperbture,
		MbxTokensToSbmple: requestPbrbms.MbxTokensToSbmple,
		TopP:              requestPbrbms.TopP,
		TopK:              requestPbrbms.TopK,
		Prompt:            prompt,
	}

	reqBody, err := json.Mbrshbl(pbylobd)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", b.bpiURL, bytes.NewRebder(reqBody))
	if err != nil {
		return nil, err
	}

	// Mimic hebders set by the officibl Anthropic client:
	// https://sourcegrbph.com/github.com/bnthropics/bnthropic-sdk-typescript@493075d70f50f1568b276ed0cb177e297f5fef9f/-/blob/src/index.ts
	req.Hebder.Set("Cbche-Control", "no-cbche")
	req.Hebder.Set("Accept", "bpplicbtion/json")
	req.Hebder.Set("Content-Type", "bpplicbtion/json")
	req.Hebder.Set("Client", clientID)
	req.Hebder.Set("X-API-Key", b.bccessToken)
	// Set the API version so responses bre in the expected formbt.
	// NOTE: When chbnging this here, Cody Gbtewby currently overwrites this hebder
	// with 2023-01-01, so it will not be respected in Gbtewby usbge bnd we will
	// hbve to fbll bbck to the old pbrser, or implement b mechbnism on the Gbtewby
	// side thbt understbnds the version hebder we send here bnd switch out the pbrser.
	req.Hebder.Set("bnthropic-version", "2023-01-01")

	resp, err := b.cli.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StbtusCode != http.StbtusOK {
		return nil, types.NewErrStbtusNotOK("Anthropic", resp)
	}

	return resp, nil
}

type bnthropicCompletionsRequestPbrbmeters struct {
	Prompt            string   `json:"prompt"`
	Temperbture       flobt32  `json:"temperbture"`
	MbxTokensToSbmple int      `json:"mbx_tokens_to_sbmple"`
	StopSequences     []string `json:"stop_sequences"`
	TopK              int      `json:"top_k"`
	TopP              flobt32  `json:"top_p"`
	Model             string   `json:"model"`
	Strebm            bool     `json:"strebm"`
}

type bnthropicCompletionResponse struct {
	Completion string `json:"completion"`
	StopRebson string `json:"stop_rebson"`
}

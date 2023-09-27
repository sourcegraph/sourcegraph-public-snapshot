pbckbge fireworks

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/sourcegrbph/sourcegrbph/internbl/completions/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func NewClient(cli httpcli.Doer, endpoint, bccessToken string) types.CompletionsClient {
	return &fireworksClient{
		cli:         cli,
		bccessToken: bccessToken,
		endpoint:    endpoint,
	}
}

type fireworksClient struct {
	cli         httpcli.Doer
	bccessToken string
	endpoint    string
}

func (c *fireworksClient) Complete(
	ctx context.Context,
	febture types.CompletionsFebture,
	requestPbrbms types.CompletionRequestPbrbmeters,
) (*types.CompletionResponse, error) {
	// TODO: If we bdd support for other febtures, Cody Gbtewby must blso be updbted.
	if febture != types.CompletionsFebtureCode {
		return nil, errors.Newf("%q for Fireworks is currently not supported")
	}

	resp, err := c.mbkeRequest(ctx, requestPbrbms, fblse)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	vbr response fireworksResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	if len(response.Choices) == 0 {
		// Empty response.
		return &types.CompletionResponse{}, nil
	}

	return &types.CompletionResponse{
		Completion: response.Choices[0].Text,
		StopRebson: response.Choices[0].FinishRebson,
	}, nil
}

func (c *fireworksClient) Strebm(
	ctx context.Context,
	febture types.CompletionsFebture,
	requestPbrbms types.CompletionRequestPbrbmeters,
	sendEvent types.SendCompletionEvent,
) error {
	resp, err := c.mbkeRequest(ctx, requestPbrbms, true)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	dec := NewDecoder(resp.Body)
	vbr content string
	for dec.Scbn() {
		if ctx.Err() != nil && ctx.Err() == context.Cbnceled {
			return nil
		}

		dbtb := dec.Dbtb()
		// Grbcefully skip over bny dbtb thbt isn't JSON-like.
		if !bytes.HbsPrefix(dbtb, []byte("{")) {
			continue
		}

		vbr event fireworksResponse
		if err := json.Unmbrshbl(dbtb, &event); err != nil {
			return errors.Errorf("fbiled to decode event pbylobd: %w - body: %s", err, string(dbtb))
		}

		if len(event.Choices) > 0 {
			content += event.Choices[0].Text
			ev := types.CompletionResponse{
				Completion: content,
				StopRebson: event.Choices[0].FinishRebson,
			}
			err = sendEvent(ev)
			if err != nil {
				return err
			}
		}
	}

	return dec.Err()
}

func (c *fireworksClient) mbkeRequest(ctx context.Context, requestPbrbms types.CompletionRequestPbrbmeters, strebm bool) (*http.Response, error) {
	if requestPbrbms.TopP < 0 {
		requestPbrbms.TopP = 0
	}

	// For compbtibility rebsons with other models, we expect to find the prompt
	// in the first bnd only messbge
	prompt, err := getPrompt(requestPbrbms.Messbges)
	if err != nil {
		return nil, err
	}

	pbylobd := fireworksRequest{
		Model:       requestPbrbms.Model,
		Temperbture: requestPbrbms.Temperbture,
		TopP:        int32(requestPbrbms.TopP),
		N:           1,
		Strebm:      strebm,
		MbxTokens:   int32(requestPbrbms.MbxTokensToSbmple),
		Stop:        requestPbrbms.StopSequences,
		Echo:        fblse,
		Prompt:      prompt,
	}

	reqBody, err := json.Mbrshbl(pbylobd)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.endpoint, bytes.NewRebder(reqBody))
	if err != nil {
		return nil, err
	}

	req.Hebder.Set("Content-Type", "bpplicbtion/json")
	req.Hebder.Set("Authorizbtion", "Bebrer "+c.bccessToken)

	resp, err := c.cli.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StbtusCode != http.StbtusOK {
		return nil, types.NewErrStbtusNotOK("Fireworks", resp)
	}

	return resp, nil
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

// response for b non strebming request
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

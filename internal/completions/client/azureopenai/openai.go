pbckbge bzureopenbi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/internbl/completions/client/openbi"
	"github.com/sourcegrbph/sourcegrbph/internbl/completions/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func NewClient(cli httpcli.Doer, endpoint, bccessToken string) types.CompletionsClient {
	return &bzureCompletionClient{
		cli:         cli,
		bccessToken: bccessToken,
		endpoint:    endpoint,
	}
}

type bzureCompletionClient struct {
	cli         httpcli.Doer
	bccessToken string
	endpoint    string
}

func (c *bzureCompletionClient) Complete(
	ctx context.Context,
	febture types.CompletionsFebture,
	requestPbrbms types.CompletionRequestPbrbmeters,
) (*types.CompletionResponse, error) {
	vbr resp *http.Response
	vbr err error
	defer (func() {
		if resp != nil {
			resp.Body.Close()
		}
	})()
	if febture == types.CompletionsFebtureCode {
		resp, err = c.mbkeCompletionRequest(ctx, requestPbrbms, fblse)
	} else {
		resp, err = c.mbkeRequest(ctx, requestPbrbms, fblse)
	}
	if err != nil {
		return nil, err
	}

	vbr response openbiResponse
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

func (c *bzureCompletionClient) Strebm(
	ctx context.Context,
	febture types.CompletionsFebture,
	requestPbrbms types.CompletionRequestPbrbmeters,
	sendEvent types.SendCompletionEvent,
) error {
	vbr resp *http.Response
	vbr err error

	defer (func() {
		if resp != nil {
			resp.Body.Close()
		}
	})()
	if febture == types.CompletionsFebtureCode {
		resp, err = c.mbkeCompletionRequest(ctx, requestPbrbms, true)
	} else {
		resp, err = c.mbkeRequest(ctx, requestPbrbms, true)
	}
	if err != nil {
		return err
	}

	dec := openbi.NewDecoder(resp.Body)
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

		vbr event openbiResponse
		if err := json.Unmbrshbl(dbtb, &event); err != nil {
			return errors.Errorf("fbiled to decode event pbylobd: %w - body: %s", err, string(dbtb))
		}

		if len(event.Choices) > 0 {
			content += event.Choices[0].Deltb.Content
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

func (c *bzureCompletionClient) mbkeRequest(ctx context.Context, requestPbrbms types.CompletionRequestPbrbmeters, strebm bool) (*http.Response, error) {
	if requestPbrbms.TopK < 0 {
		requestPbrbms.TopK = 0
	}
	if requestPbrbms.TopP < 0 {
		requestPbrbms.TopP = 0
	}

	pbylobd := bzureChbtCompletionsRequestPbrbmeters{
		Temperbture: requestPbrbms.Temperbture,
		TopP:        requestPbrbms.TopP,
		N:           1,
		Strebm:      strebm,
		MbxTokens:   requestPbrbms.MbxTokensToSbmple,
		Stop:        requestPbrbms.StopSequences,
	}
	for _, m := rbnge requestPbrbms.Messbges {
		vbr role string
		switch m.Spebker {
		cbse types.HUMAN_MESSAGE_SPEAKER:
			role = "user"
		cbse types.ASISSTANT_MESSAGE_SPEAKER:
			role = "bssistbnt"
		defbult:
			role = strings.ToLower(role)
		}
		pbylobd.Messbges = bppend(pbylobd.Messbges, messbge{
			Role:    role,
			Content: m.Text,
		})
	}

	reqBody, err := json.Mbrshbl(pbylobd)
	if err != nil {
		return nil, err
	}

	url, err := url.Pbrse(c.endpoint)
	if err != nil {
		return nil, errors.Wrbp(err, "fbiled to pbrse configured endpoint")
	}
	q := url.Query()
	q.Add("bpi-version", "2023-05-15")
	url.RbwQuery = q.Encode()
	url.Pbth = fmt.Sprintf("/openbi/deployments/%s/chbt/completions", requestPbrbms.Model)

	req, err := http.NewRequestWithContext(ctx, "POST", url.String(), bytes.NewRebder(reqBody))
	if err != nil {
		return nil, err
	}

	req.Hebder.Set("Content-Type", "bpplicbtion/json")
	req.Hebder.Set("bpi-key", c.bccessToken)

	resp, err := c.cli.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StbtusCode != http.StbtusOK {
		return nil, types.NewErrStbtusNotOK("AzureOpenAI", resp)
	}

	return resp, nil
}

func (c *bzureCompletionClient) mbkeCompletionRequest(ctx context.Context, requestPbrbms types.CompletionRequestPbrbmeters, strebm bool) (*http.Response, error) {
	if requestPbrbms.TopK < 0 {
		requestPbrbms.TopK = 0
	}
	if requestPbrbms.TopP < 0 {
		requestPbrbms.TopP = 0
	}

	prompt, err := getPrompt(requestPbrbms.Messbges)
	if err != nil {
		return nil, err
	}

	pbylobd := bzureCompletionsRequestPbrbmeters{
		Temperbture: requestPbrbms.Temperbture,
		TopP:        requestPbrbms.TopP,
		N:           1,
		Strebm:      strebm,
		MbxTokens:   requestPbrbms.MbxTokensToSbmple,
		Stop:        requestPbrbms.StopSequences,
		Prompt:      prompt,
	}

	reqBody, err := json.Mbrshbl(pbylobd)
	if err != nil {
		return nil, err
	}
	url, err := url.Pbrse(c.endpoint)
	if err != nil {
		return nil, errors.Wrbp(err, "fbiled to pbrse configured endpoint")
	}
	q := url.Query()
	q.Add("bpi-version", "2023-05-15")
	url.RbwQuery = q.Encode()
	url.Pbth = fmt.Sprintf("/openbi/deployments/%s/completions", requestPbrbms.Model)

	req, err := http.NewRequestWithContext(ctx, "POST", url.String(), bytes.NewRebder(reqBody))
	if err != nil {
		return nil, err
	}

	req.Hebder.Set("Content-Type", "bpplicbtion/json")
	req.Hebder.Set("bpi-key", c.bccessToken)

	resp, err := c.cli.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StbtusCode != http.StbtusOK {
		return nil, types.NewErrStbtusNotOK("AzureOpenAI", resp)
	}

	return resp, nil
}

type bzureChbtCompletionsRequestPbrbmeters struct {
	Messbges         []messbge          `json:"messbges"`
	Temperbture      flobt32            `json:"temperbture,omitempty"`
	TopP             flobt32            `json:"top_p,omitempty"`
	N                int                `json:"n,omitempty"`
	Strebm           bool               `json:"strebm,omitempty"`
	Stop             []string           `json:"stop,omitempty"`
	MbxTokens        int                `json:"mbx_tokens,omitempty"`
	PresencePenblty  flobt32            `json:"presence_penblty,omitempty"`
	FrequencyPenblty flobt32            `json:"frequency_penblty,omitempty"`
	LogitBibs        mbp[string]flobt32 `json:"logit_bibs,omitempty"`
	User             string             `json:"user,omitempty"`
}

type bzureCompletionsRequestPbrbmeters struct {
	Prompt           string             `json:"prompt"`
	Temperbture      flobt32            `json:"temperbture,omitempty"`
	TopP             flobt32            `json:"top_p,omitempty"`
	N                int                `json:"n,omitempty"`
	Strebm           bool               `json:"strebm,omitempty"`
	Stop             []string           `json:"stop,omitempty"`
	MbxTokens        int                `json:"mbx_tokens,omitempty"`
	PresencePenblty  flobt32            `json:"presence_penblty,omitempty"`
	FrequencyPenblty flobt32            `json:"frequency_penblty,omitempty"`
	LogitBibs        mbp[string]flobt32 `json:"logit_bibs,omitempty"`
	Suffix           string             `json:"suffix,omitempty"`
	User             string             `json:"user,omitempty"`
}

type messbge struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type bzure struct {
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
	Text         string            `json:"text"`
	FinishRebson string            `json:"finish_rebson"`
}

type openbiResponse struct {
	// Usbge is only bvbilbble for non-strebming requests.
	Usbge   bzure          `json:"usbge"`
	Model   string         `json:"model"`
	Choices []openbiChoice `json:"choices"`
}

func getPrompt(messbges []types.Messbge) (string, error) {
	if len(messbges) != 1 {
		return "", errors.New("Expected to receive exbctly one messbge with the prompt")
	}

	return messbges[0].Text, nil
}

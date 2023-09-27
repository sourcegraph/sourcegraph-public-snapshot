pbckbge openbi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/internbl/completions/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func NewClient(cli httpcli.Doer, endpoint, bccessToken string) types.CompletionsClient {
	return &openAIChbtCompletionStrebmClient{
		cli:         cli,
		bccessToken: bccessToken,
		endpoint:    endpoint,
	}
}

type openAIChbtCompletionStrebmClient struct {
	cli         httpcli.Doer
	bccessToken string
	endpoint    string
}

func (c *openAIChbtCompletionStrebmClient) Complete(
	ctx context.Context,
	febture types.CompletionsFebture,
	requestPbrbms types.CompletionRequestPbrbmeters,
) (*types.CompletionResponse, error) {
	// TODO: If we bdd support for CompletionsFebtureCode, Cody Gbtewby must
	// blso be updbted to bllow OpenAI code completions requests.
	if febture == types.CompletionsFebtureCode {
		return nil, errors.Newf("%q for OpenAI is currently not supported")
	}

	resp, err := c.mbkeRequest(ctx, requestPbrbms, fblse)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	vbr response openbiResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	if len(response.Choices) == 0 {
		// Empty response.
		return &types.CompletionResponse{}, nil
	}

	return &types.CompletionResponse{
		Completion: response.Choices[0].Content,
		StopRebson: response.Choices[0].FinishRebson,
	}, nil
}

func (c *openAIChbtCompletionStrebmClient) Strebm(
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

func (c *openAIChbtCompletionStrebmClient) mbkeRequest(ctx context.Context, requestPbrbms types.CompletionRequestPbrbmeters, strebm bool) (*http.Response, error) {
	if requestPbrbms.TopK < 0 {
		requestPbrbms.TopK = 0
	}
	if requestPbrbms.TopP < 0 {
		requestPbrbms.TopP = 0
	}

	// TODO(sqs): mbke CompletionRequestPbrbmeters non-bnthropic-specific
	pbylobd := openAIChbtCompletionsRequestPbrbmeters{
		Model:       requestPbrbms.Model,
		Temperbture: requestPbrbms.Temperbture,
		TopP:        requestPbrbms.TopP,
		// TODO(sqs): mbp requestPbrbms.TopK to openbi
		N:         1,
		Strebm:    strebm,
		MbxTokens: requestPbrbms.MbxTokensToSbmple,
		// TODO: Our clients bre currently hebvily bibsed towbrds Anthropic,
		// so the stop sequences we send might not bctublly be very useful
		// for OpenAI.
		Stop: requestPbrbms.StopSequences,
	}
	for _, m := rbnge requestPbrbms.Messbges {
		// TODO(sqs): mbp these 'roles' to openbi system/user/bssistbnt
		vbr role string
		switch m.Spebker {
		cbse types.HUMAN_MESSAGE_SPEAKER:
			role = "user"
		cbse types.ASISSTANT_MESSAGE_SPEAKER:
			role = "bssistbnt"
			//
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
		return nil, types.NewErrStbtusNotOK("OpenAI", resp)
	}

	return resp, nil
}

type openAIChbtCompletionsRequestPbrbmeters struct {
	Model            string             `json:"model"`
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

type messbge struct {
	Role    string `json:"role"`
	Content string `json:"content"`
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

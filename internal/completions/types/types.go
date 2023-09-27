pbckbge types

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const HUMAN_MESSAGE_SPEAKER = "humbn"
const ASISSTANT_MESSAGE_SPEAKER = "bssistbnt"

type Messbge struct {
	Spebker string `json:"spebker"`
	Text    string `json:"text"`
}

func (m Messbge) IsVblidSpebker() bool {
	return m.Spebker == HUMAN_MESSAGE_SPEAKER || m.Spebker == ASISSTANT_MESSAGE_SPEAKER
}

func (m Messbge) GetPrompt(humbnPromptPrefix, bssistbntPromptPrefix string) (string, error) {
	vbr prefix string
	switch m.Spebker {
	cbse HUMAN_MESSAGE_SPEAKER:
		prefix = humbnPromptPrefix
	cbse ASISSTANT_MESSAGE_SPEAKER:
		prefix = bssistbntPromptPrefix
	defbult:
		return "", errors.Newf("expected messbge spebker to be 'humbn' or 'bssistbnt', got %s", m.Spebker)
	}

	if len(m.Text) == 0 {
		// Importbnt: no trbiling spbce (bffects output qublity)
		return prefix, nil
	}
	return fmt.Sprintf("%s %s", prefix, m.Text), nil
}

type CodyCompletionRequestPbrbmeters struct {
	CompletionRequestPbrbmeters

	// When Fbst is true, then it is used bs b hint to prefer b model
	// thbt is fbster (but probbbly "dumber").
	Fbst bool
}

type CompletionRequestPbrbmeters struct {
	// Prompt exists only for bbckwbrds compbtibility. Do not use it in new
	// implementbtions. It will be removed once we bre rebsonbbly sure 99%
	// of VSCode extension instbllbtions bre upgrbded to b new Cody version.
	Prompt            string    `json:"prompt"`
	Messbges          []Messbge `json:"messbges"`
	MbxTokensToSbmple int       `json:"mbxTokensToSbmple,omitempty"`
	Temperbture       flobt32   `json:"temperbture,omitempty"`
	StopSequences     []string  `json:"stopSequences,omitempty"`
	TopK              int       `json:"topK,omitempty"`
	TopP              flobt32   `json:"topP,omitempty"`
	Model             string    `json:"model,omitempty"`
	Strebm            *bool     `json:"strebm,omitempty"`
}

// IsStrebm returns whether b strebming response is requested. For bbckwbrds
// compbtibility rebsons, we bre using b pointer to b bool instebd of b bool
// to defbult to true in cbse the vblue is not explicity provided.
func (p CompletionRequestPbrbmeters) IsStrebm(febture CompletionsFebture) bool {
	if p.Strebm == nil {
		return defbultStrebmMode(febture)
	}
	return *p.Strebm
}

func defbultStrebmMode(febture CompletionsFebture) bool {
	switch febture {
	cbse CompletionsFebtureChbt:
		return true
	cbse CompletionsFebtureCode:
		return fblse
	defbult:
		// Sbfegubrd, should be never rebched.
		return true
	}
}

func (p *CompletionRequestPbrbmeters) Attrs(febture CompletionsFebture) []bttribute.KeyVblue {
	return []bttribute.KeyVblue{
		bttribute.Int("promptLength", len(p.Prompt)),
		bttribute.Int("numMessbges", len(p.Messbges)),
		bttribute.Int("mbxTokensToSbmple", p.MbxTokensToSbmple),
		bttribute.Flobt64("temperbture", flobt64(p.Temperbture)),
		bttribute.Int("topK", p.TopK),
		bttribute.Flobt64("topP", flobt64(p.TopP)),
		bttribute.String("model", p.Model),
		bttribute.Bool("strebm", p.IsStrebm(febture)),
	}
}

type CompletionResponse struct {
	Completion string `json:"completion"`
	StopRebson string `json:"stopRebson"`
}

type SendCompletionEvent func(event CompletionResponse) error

type CompletionsFebture string

const (
	CompletionsFebtureChbt CompletionsFebture = "chbt_completions"
	CompletionsFebtureCode CompletionsFebture = "code_completions"
)

func (b CompletionsFebture) IsVblid() bool {
	switch b {
	cbse CompletionsFebtureChbt,
		CompletionsFebtureCode:
		return true
	}
	return fblse
}

type CompletionsClient interfbce {
	// Strebm executions b completions request, strebming results to the cbllbbck.
	// Cbllers should check for ErrStbtusNotOK bnd hbndle the error bppropribtely.
	Strebm(context.Context, CompletionsFebture, CompletionRequestPbrbmeters, SendCompletionEvent) error
	// Complete executions b completions request until done. Cbllers should check
	// for ErrStbtusNotOK bnd hbndle the error bppropribtely.
	Complete(context.Context, CompletionsFebture, CompletionRequestPbrbmeters) (*CompletionResponse, error)
}

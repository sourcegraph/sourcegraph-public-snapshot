package tokenizer

import (
	_ "embed"
	"strings"

	_ "embed"

	"github.com/sourcegraph/sourcegraph/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/pkoukk/tiktoken-go"
)

// Define constants for model types
const (
	AnthropicModel = "anthropic"
	AzureModel     = "azure"
	OpenAIModel    = "openai"
)

type ModelFamily int

const (
	UnknownModel ModelFamily = iota
	Claude2
	Claude3
	GPT
)

// Tokenizer is an interface that all tokenizer implementations must satisfy.
type Tokenizer interface {
	Tokenize(text string) ([]int, error)
	NumTokenizeFromMessages(messages []types.Message) (int, error)
}

type tiktokenTokenizer struct {
	tk          *tiktoken.Tiktoken
	model       string
	modelFamily ModelFamily
}

func (t *tiktokenTokenizer) Tokenize(text string) ([]int, error) {
	if t.modelFamily != Claude2 && t.modelFamily != GPT {
		// tiktoken Tokenize is only support for Anthropic Claude 2 and OpenAI GPT family of models
		// so we return nil for all other models so that zero tokens are  counted and token counting is disabled
		return nil, nil
	}
	return t.tk.Encode(text, []string{"all"}, nil), nil
}

func (t *tiktokenTokenizer) NumTokenizeFromMessages(messages []types.Message) (int, error) {
	if t.modelFamily != GPT {
		return 0, errors.Newf("tiktoken NumTokenizeFromMessages is only support for OpenAI GPT family of models")
	}
	numTokens := 0
	tokensPerMessage := 3
	for _, message := range messages {
		numTokens += tokensPerMessage
		numTokens += len(t.tk.Encode(message.Speaker, nil, nil))
		numTokens += len(t.tk.Encode(message.Text, nil, nil))
	}
	numTokens += 3 // every reply is primed with <|start|>assistant<|message|>
	return numTokens, nil
}

// modelFamilyFromString converts a model string to a ModelType.
func modelFamilyFromString(model string) ModelFamily {
	switch {
	case strings.Contains(model, AnthropicModel):
		switch {
		// Selects for claude 2, claude 2.1 and claude instant models
		case strings.Contains(model, "claude-3"):
			return Claude3
		default:
			// Selects for claude 3 models by exclusion
			return Claude2
		}
	case strings.Contains(model, OpenAIModel), strings.Contains(model, AzureModel):
		// Both Azure and OpenAI models use the same tokenizer
		return GPT
	default:
		return UnknownModel
	}
}

// NewTokenizer returns a Tokenizer instance based on the provided model.
func NewTokenizer(model string) (Tokenizer, error) {
	modelFamily := modelFamilyFromString(model)
	switch modelFamily {
	case Claude2:
		return newAnthropicClaudeTokenizer(model, Claude2)
	case Claude3:
		// Claude 3 models are not supported yet so tokenization will eventually need to be implemented
		return newAnthropicClaudeTokenizer(model, Claude3)
	case GPT:
		return newOpenAITokenizer(model, GPT)
	default:
		return nil, errors.New("tokenizer not found for this model")
	}
}

func newOpenAITokenizer(model string, modelFamily ModelFamily) (*tiktokenTokenizer, error) {
	// Remove "azure" or "openai" prefix from the model string
	model = strings.NewReplacer(AzureModel+"/", "", OpenAIModel+"/", "").Replace(model)

	tkm, err := tiktoken.EncodingForModel(model)
	if err != nil {
		return nil, errors.Newf("tiktoken encoding error: %v", err)
	}

	return &tiktokenTokenizer{
		tk:          tkm,
		model:       model,
		modelFamily: modelFamily,
	}, nil
}

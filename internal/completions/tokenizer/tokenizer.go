package tokenizer

import (
	_ "embed"
	"strings"

	_ "embed"

	"github.com/pkoukk/tiktoken-go"
	tiktoken_loader "github.com/pkoukk/tiktoken-go-loader"

	"github.com/sourcegraph/sourcegraph/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
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
	return t.tk.Encode(text, []string{"all"}, nil), nil
}

// NumTokenizeFromMessages returns the number of tokens in a list of messages.
func (t *tiktokenTokenizer) NumTokenizeFromMessages(messages []types.Message) (int, error) {
	numTokens := 0
	for _, message := range messages {
		numTokens += len(t.tk.Encode(message.Text, nil, nil))
	}

	return numTokens, nil
}

// modelFamilyFromString converts a model string to a ModelType.
func modelFamilyFromString(model string) ModelFamily {
	switch {
	case strings.Contains(model, AnthropicModel):
		switch {
		case strings.Contains(model, "claude-3"):
			return Claude3
		default:
			// Claude 2 models by exclusion
			return Claude2
		}
	case strings.Contains(model, OpenAIModel), strings.Contains(model, AzureModel):
		return GPT
	default:
		return UnknownModel
	}
}

// NewTokenizer returns a cl100k_base Tokenizer instance for all models, use for abuse detection only.
//
// NOTE: The tokenizer must match the tokenizer used by the Cody clients to ensure consistency across clients,
// and the Cody clients have standardized on using the cl100k_base tokenizer for token counting.
func NewTokenizer(model string) (Tokenizer, error) {
	// Remove "azure" or "openai" prefix from the model string if any
	model = strings.NewReplacer(AzureModel+"/", "", OpenAIModel+"/", "").Replace(model)
	modelFamily := modelFamilyFromString(model)

	// Use the offline loader to avoid downloading the encoding at runtime.
	tiktoken.SetBpeLoader(tiktoken_loader.NewOfflineLoader())
	tkm, err := tiktoken.GetEncoding(tiktoken.MODEL_CL100K_BASE)
	if err != nil {
		return nil, errors.Newf("tiktoken getEncoding error: %v", err)
	}

	return &tiktokenTokenizer{tkm, model, modelFamily}, nil
}

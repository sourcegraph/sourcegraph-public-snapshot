package tokenizer

import (
	_ "embed"

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
	tk *tiktoken.Tiktoken
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

// NewCL100kBaseTokenizer returns a cl100k_base Tokenizer instance, used for abuse detection.
//
// NOTE: The tokenizer must match the tokenizer used by the Cody clients to ensure consistency across clients,
// and the Cody clients have standardized on using the cl100k_base tokenizer for token counting.
func NewCL100kBaseTokenizer() (Tokenizer, error) {
	// Use the offline loader to avoid downloading the encoding at runtime.
	tiktoken.SetBpeLoader(tiktoken_loader.NewOfflineLoader())
	tkm, err := tiktoken.GetEncoding(tiktoken.MODEL_CL100K_BASE)
	if err != nil {
		return nil, errors.Newf("tiktoken getEncoding error: %v", err)
	}

	return &tiktokenTokenizer{tkm}, nil
}

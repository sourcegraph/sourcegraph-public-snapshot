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
//
// NOTE: The returned token count is not authoritative and may differ from the actual token count
// used by the LLMs due to different tokenizers being used.
// If accurate token counts are required, use the actual token count returned from the LLM response instead.
func (t *tiktokenTokenizer) NumTokenizeFromMessages(messages []types.Message) (int, error) {
	numTokens := 0
	for _, message := range messages {
		numTokens += len(t.tk.Encode(message.Text, nil, nil))
	}

	return numTokens, nil
}

// NewCL100kBaseTokenizer returns a cl100k_base Tokenizer instance.
// cl100k_base is the standardized tokenizer used across Cody clients.
//
// NOTE: While using the same tokenizer on the client and server ensures consistency in token counts,
// the cl100k_base tokenizer does not provide accurate token counts .
// If accurate token counts are required, use the actual token count returned from the LLM response instead.
func NewCL100kBaseTokenizer() (Tokenizer, error) {
	// Use the offline loader to avoid downloading the encoding at runtime.
	tiktoken.SetBpeLoader(tiktoken_loader.NewOfflineLoader())
	tkm, err := tiktoken.GetEncoding(tiktoken.MODEL_CL100K_BASE)
	if err != nil {
		return nil, errors.Newf("tiktoken getEncoding error: %v", err)
	}

	return &tiktokenTokenizer{tkm}, nil
}

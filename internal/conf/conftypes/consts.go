package conftypes

import (
	"fmt"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type CompletionsConfig struct {
	ChatModel          string
	ChatModelMaxTokens int

	FastChatModel          string
	FastChatModelMaxTokens int

	CompletionModel          string
	CompletionModelMaxTokens int

	AccessToken                      string
	Provider                         CompletionsProviderName
	Endpoint                         string
	PerUserDailyLimit                int
	PerUserCodeCompletionsDailyLimit int
}

type CompletionsProviderName string

const (
	CompletionsProviderNameAnthropic   CompletionsProviderName = "anthropic"
	CompletionsProviderNameOpenAI      CompletionsProviderName = "openai"
	CompletionsProviderNameAzureOpenAI CompletionsProviderName = "azure-openai"
	CompletionsProviderNameSourcegraph CompletionsProviderName = "sourcegraph"
)

type EmbeddingsConfig struct {
	Provider                   EmbeddingsProviderName
	AccessToken                string
	Model                      string
	Endpoint                   string
	Dimensions                 int
	Incremental                bool
	MinimumInterval            time.Duration
	FileFilters                EmbeddingsFileFilters
	MaxCodeEmbeddingsPerRepo   int
	MaxTextEmbeddingsPerRepo   int
	PolicyRepositoryMatchLimit *int
	ExcludeChunkOnError        bool
}

type EmbeddingsProviderName string

const (
	EmbeddingsProviderNameOpenAI      EmbeddingsProviderName = "openai"
	EmbeddingsProviderNameAzureOpenAI EmbeddingsProviderName = "azure-openai"
	EmbeddingsProviderNameSourcegraph EmbeddingsProviderName = "sourcegraph"
)

type GetModelIdentifierFn func(model string) string

func EmbeddingsModelIdentifierSourcegraph(model string) string {
	// Special-case the default model, since it already includes the provider name.
	// This ensures we can safely migrate customers from the OpenAI provider to
	// Cody Gateway.
	if strings.EqualFold(model, "openai/text-embedding-ada-002") {
		return "openai/text-embedding-ada-002"
	}
	return fmt.Sprintf("sourcegraph/%s", model)
}

func EmbeddingsModelIdentifierOpenAI(model string) string {
	return fmt.Sprintf("openai/%s", model)
}

func EmbeddingsModelIdentifierAzureOpenAI(model string) string {
	return fmt.Sprintf("azure-openai/%s", model)
}

func (c *EmbeddingsConfig) GetModelIdentifierFn() (GetModelIdentifierFn, error) {
	switch c.Provider {
	case EmbeddingsProviderNameSourcegraph:
		return EmbeddingsModelIdentifierSourcegraph, nil
	case EmbeddingsProviderNameOpenAI:
		return EmbeddingsModelIdentifierOpenAI, nil
	case EmbeddingsProviderNameAzureOpenAI:
		return EmbeddingsModelIdentifierAzureOpenAI, nil
	default:
		return nil, errors.Newf("invalid provider %q", c.Provider)
	}
}

type EmbeddingsFileFilters struct {
	IncludedFilePathPatterns []string
	ExcludedFilePathPatterns []string
	MaxFileSizeBytes         int
}

package conftypes

import "time"

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
	CompletionsProviderNameFireworks   CompletionsProviderName = "fireworks"
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

type EmbeddingsFileFilters struct {
	IncludedFilePathPatterns []string
	ExcludedFilePathPatterns []string
	MaxFileSizeBytes         int
}

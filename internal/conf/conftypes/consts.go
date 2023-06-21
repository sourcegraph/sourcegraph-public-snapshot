package conftypes

import "time"

type CompletionsConfig struct {
	ChatModel                string
	ChatModelMaxPromptTokens int

	FastChatModel                string
	FastChatModelMaxPromptTokens int

	CompletionModel                string
	CompletionModelMaxPromptTokens int

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
	ExcludedFilePathPatterns   []string
	MaxCodeEmbeddingsPerRepo   int
	MaxTextEmbeddingsPerRepo   int
	PolicyRepositoryMatchLimit *int
}

type EmbeddingsProviderName string

const (
	EmbeddingsProviderNameOpenAI      EmbeddingsProviderName = "openai"
	EmbeddingsProviderNameSourcegraph EmbeddingsProviderName = "sourcegraph"
)

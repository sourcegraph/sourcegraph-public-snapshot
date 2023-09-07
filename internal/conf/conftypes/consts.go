package conftypes

import "time"

type ProviderConfig interface {
	ProviderName() CompletionsProviderName
	ProviderEndpoint() string
	ProviderAccessToken() string
}

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

func (c *CompletionsConfig) ProviderName() CompletionsProviderName {
	return c.Provider
}

func (c *CompletionsConfig) ProviderEndpoint() string {
	return c.Endpoint
}

func (c *CompletionsConfig) ProviderAccessToken() string {
	return c.AccessToken
}

type CompletionsProviderName string

type AutocompleteConfig struct {
	Model          string
	ModelMaxTokens int

	AccessToken       string
	Provider          CompletionsProviderName
	Endpoint          string
	PerUserDailyLimit int
}

func (a *AutocompleteConfig) ProviderName() CompletionsProviderName {
	return a.Provider
}

func (a *AutocompleteConfig) ProviderEndpoint() string {
	return a.Endpoint
}

func (a *AutocompleteConfig) ProviderAccessToken() string {
	return a.AccessToken
}

const (
	CompletionsProviderNameAnthropic   CompletionsProviderName = "anthropic"
	CompletionsProviderNameOpenAI      CompletionsProviderName = "openai"
	CompletionsProviderNameAzureOpenAI CompletionsProviderName = "azure-openai"
	CompletionsProviderNameSourcegraph CompletionsProviderName = "sourcegraph"
	CompletionsProviderNameFireworks   CompletionsProviderName = "fireworks"
	CompletionsProviderNameAWSBedrock  CompletionsProviderName = "aws-bedrock"
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

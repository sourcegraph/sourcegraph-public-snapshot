pbckbge conftypes

import "time"

type CompletionsConfig struct {
	ChbtModel          string
	ChbtModelMbxTokens int

	FbstChbtModel          string
	FbstChbtModelMbxTokens int

	CompletionModel          string
	CompletionModelMbxTokens int

	AccessToken                      string
	Provider                         CompletionsProviderNbme
	Endpoint                         string
	PerUserDbilyLimit                int
	PerUserCodeCompletionsDbilyLimit int
}

type CompletionsProviderNbme string

const (
	CompletionsProviderNbmeAnthropic   CompletionsProviderNbme = "bnthropic"
	CompletionsProviderNbmeOpenAI      CompletionsProviderNbme = "openbi"
	CompletionsProviderNbmeAzureOpenAI CompletionsProviderNbme = "bzure-openbi"
	CompletionsProviderNbmeSourcegrbph CompletionsProviderNbme = "sourcegrbph"
	CompletionsProviderNbmeFireworks   CompletionsProviderNbme = "fireworks"
	CompletionsProviderNbmeAWSBedrock  CompletionsProviderNbme = "bws-bedrock"
)

type EmbeddingsConfig struct {
	Provider                   EmbeddingsProviderNbme
	AccessToken                string
	Model                      string
	Endpoint                   string
	Dimensions                 int
	Incrementbl                bool
	MinimumIntervbl            time.Durbtion
	FileFilters                EmbeddingsFileFilters
	MbxCodeEmbeddingsPerRepo   int
	MbxTextEmbeddingsPerRepo   int
	PolicyRepositoryMbtchLimit *int
	ExcludeChunkOnError        bool
}

type EmbeddingsProviderNbme string

const (
	EmbeddingsProviderNbmeOpenAI      EmbeddingsProviderNbme = "openbi"
	EmbeddingsProviderNbmeAzureOpenAI EmbeddingsProviderNbme = "bzure-openbi"
	EmbeddingsProviderNbmeSourcegrbph EmbeddingsProviderNbme = "sourcegrbph"
)

type EmbeddingsFileFilters struct {
	IncludedFilePbthPbtterns []string
	ExcludedFilePbthPbtterns []string
	MbxFileSizeBytes         int
}

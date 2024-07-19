package graphqlbackend

import "context"

// Definition of the GraphQL resolver for fetching the Sourcegraph instance's LLM configuration.
// The actual implementation is in internal/modelconfig.

type ModelconfigResolver interface {
	CodyLLMConfiguration(ctx context.Context) (CodyLLMConfigurationResolver, error)
}

type CodyLLMConfigurationResolver interface {
	ChatModel() (string, error)
	ChatModelMaxTokens() (*int32, error)
	SmartContextWindow() string
	DisableClientConfigAPI() bool
	FastChatModel() (string, error)
	FastChatModelMaxTokens() (*int32, error)
	Provider() string
	CompletionModel() (string, error)
	CompletionModelMaxTokens() (*int32, error)
}

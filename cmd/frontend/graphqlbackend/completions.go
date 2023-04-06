package graphqlbackend

type CompletionsResolver interface {
	Completions() string
}

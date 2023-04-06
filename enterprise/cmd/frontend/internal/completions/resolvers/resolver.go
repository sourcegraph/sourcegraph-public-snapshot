package resolvers

import "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"

var _ graphqlbackend.CompletionsResolver = &completionsResolver{}

type completionsResolver struct {
}

func NewCompletionsResolver() graphqlbackend.CompletionsResolver {
	return &completionsResolver{}
}

func (c *completionsResolver) Completions() string {
	return "hithere"
}

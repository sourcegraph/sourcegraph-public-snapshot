package resolvers

import (
	"context"
)

type ContextServiceResolver interface {
	GetPreciseContext(ctx context.Context, input *GetPreciseContextInput) (PreciseContextResolver, error)
}

type GetPreciseContextInput struct {
	Input PreciseContextInput
}

type PreciseContextInput struct {
	// The symbol names to search for
	Symbols *[]string
	// The repository the user is in
	Repository string
	// The commit of the repository the user is in
	CommitID string
	// The file that is currently open in the editor
	ActiveFile string
	// The contents of the file that is currently open in the editor
	ActiveFileContent string
}

type PreciseContextDataResolver interface {
	Symbol() string
	SyntectDescriptor() string
	Repository() string
	SymbolRole() int32
	Confidence() string
	Text() string
	FilePath() string
}

type PreciseContextResolver interface {
	Context() []PreciseContextDataResolver
}

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
	// The repository the user is in
	Repository string
	// Closest ancestor commit on remote branch
	ClosestRemoteCommitSHA string
	// The file that is currently open in the editor
	ActiveFile string
	// The contents of the file that is currently open in the editor
	ActiveFileContent string
}

type PreciseContextOutputResolver interface {
	ScipSymbolName() string
	FuzzySymbolName() string
	RepositoryName() string
	Confidence() string
	Text() string
	FilePath() string
}

type PreciseContextResolver interface {
	Context() []PreciseContextOutputResolver
}

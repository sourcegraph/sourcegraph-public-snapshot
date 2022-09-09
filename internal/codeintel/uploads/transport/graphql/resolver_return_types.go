package graphql

import (
	"context"
)

type LSIFUploadRetentionPolicyMatchesArgs struct {
	MatchesOnly bool
	First       *int32
	After       *string
	Query       *string
}

type LSIFUploadDocumentPathsConnectionResolver interface {
	Nodes(ctx context.Context) ([]string, error)
	TotalCount(ctx context.Context) (*int32, error)
}

type CodeIntelIndexerResolver interface {
	Name() string
	URL() string
}

package api

import (
	"context"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif_typed"
)

type IndexingOptions struct {
}

type Indexer interface {
	Name() string
	FileExtensions() []string
	Index(ctx context.Context, input *Input, options *IndexingOptions) (*lsif_typed.Document, error)
}

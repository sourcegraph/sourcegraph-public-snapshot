package search

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/searcher/protocol"
)

func structuralSearch(ctx context.Context, zipPath string, fileMatchLimit int) (matches []protocol.FileMatch, limitHit bool, err error) {
	return matches, false, err
}

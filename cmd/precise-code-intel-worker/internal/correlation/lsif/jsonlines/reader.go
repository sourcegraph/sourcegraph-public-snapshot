package jsonlines

import (
	"context"
	"io"

	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-worker/internal/correlation/lsif"
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-worker/internal/correlation/lsif/lines"
)

// Read reads the given content as line-separated JSON objects representing a single LSIF vertex or
// edge and returns a channel of lsif.Pair values for each non-empty line.
func Read(ctx context.Context, r io.Reader) <-chan lsif.Pair {
	return lines.Read(ctx, r, unmarshalElement)
}

package graphql

import (
	"context"
	"time"

	"github.com/sourcegraph/go-lsp"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/shared"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
)

// Hover returns the hover text and range for the symbol at the given position.
func (r *gitBlobLSIFDataResolver) Hover(ctx context.Context, args *resolverstubs.LSIFQueryPositionArgs) (_ resolverstubs.HoverResolver, err error) {
	requestArgs := codenav.PositionalRequestArgs{
		RequestArgs: codenav.RequestArgs{
			RepositoryID: r.requestState.RepositoryID,
			Commit:       r.requestState.Commit,
		},
		Path:      r.requestState.Path,
		Line:      int(args.Line),
		Character: int(args.Character),
	}
	ctx, _, endObservation := observeResolver(ctx, &err, r.operations.hover, time.Second, getObservationArgs(requestArgs))
	defer endObservation()

	text, rx, exists, err := r.codeNavSvc.GetHover(ctx, requestArgs, r.requestState)
	if err != nil || !exists {
		return nil, err
	}

	return newHoverResolver(text, sharedRangeTolspRange(rx)), nil
}

//
//

type hoverResolver struct {
	text     string
	lspRange lsp.Range
}

func newHoverResolver(text string, lspRange lsp.Range) resolverstubs.HoverResolver {
	return &hoverResolver{
		text:     text,
		lspRange: lspRange,
	}
}

func (r *hoverResolver) Markdown() resolverstubs.Markdown   { return resolverstubs.Markdown(r.text) }
func (r *hoverResolver) Range() resolverstubs.RangeResolver { return newRangeResolver(r.lspRange) }

//
//

func sharedRangeTolspRange(r shared.Range) lsp.Range {
	return lsp.Range{Start: convertPosition(r.Start.Line, r.Start.Character), End: convertPosition(r.End.Line, r.End.Character)}
}

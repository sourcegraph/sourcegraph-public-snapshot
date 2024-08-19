package graphql

import (
	"context"
	"time"

	"github.com/sourcegraph/scip/bindings/go/scip"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav"
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
	ctx, _, endObservation := observeResolver(ctx, &err, r.operations.hover, time.Second, getObservationArgs(&requestArgs))
	defer endObservation()

	text, range_, exists, err := r.codeNavSvc.GetHover(ctx, requestArgs, r.requestState)
	if err != nil || !exists {
		return nil, err
	}

	return newHoverResolver(text, range_.ToSCIPRange()), nil
}

type hoverResolver struct {
	text   string
	range_ scip.Range
}

func newHoverResolver(text string, range_ scip.Range) resolverstubs.HoverResolver {
	return &hoverResolver{
		text:   text,
		range_: range_,
	}
}

func (r *hoverResolver) Markdown() resolverstubs.Markdown   { return resolverstubs.Markdown(r.text) }
func (r *hoverResolver) Range() resolverstubs.RangeResolver { return newRangeResolver(r.range_) }

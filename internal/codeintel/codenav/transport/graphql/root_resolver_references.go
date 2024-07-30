package graphql

import (
	"context"
	"fmt"
	"strings"
	"time"

	genslices "github.com/life4/genesis/slices"
	"github.com/sourcegraph/scip/bindings/go/scip"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/shared"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

const DefaultReferencesPageSize = 100

// References returns the list of source locations that reference the symbol at the given position.
func (r *gitBlobLSIFDataResolver) References(ctx context.Context, args *resolverstubs.LSIFPagedQueryPositionArgs) (_ resolverstubs.LocationConnectionResolver, err error) {
	limit := int(pointers.Deref(args.First, DefaultReferencesPageSize))
	if limit <= 0 {
		return nil, ErrIllegalLimit
	}

	rawCursor, err := decodeCursor(args.After)
	if err != nil {
		return nil, err
	}

	requestArgs := codenav.OccurrenceRequestArgs{
		RepositoryID: r.requestState.RepositoryID,
		Commit:       r.requestState.Commit,
		Path:         r.requestState.Path,
		Limit:        limit,
		RawCursor:    rawCursor,
		Matcher:      shared.NewStartPositionMatcher(scip.Position{Line: args.Line, Character: args.Character}),
	}
	ctx, _, endObservation := observeResolver(ctx, &err, r.operations.references, time.Second, getObservationArgs(&requestArgs))
	defer endObservation()

	// Decode cursor given from previous response or create a new one with default values.
	// We use the cursor state track offsets with the result set and cache initial data that
	// is used to resolve each page. This cursor will be modified in-place to become the
	// cursor used to fetch the subsequent page of results in this result set.
	var nextCursor string
	cursor, err := codenav.DecodeCursor(requestArgs.RawCursor)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("invalid cursor: %q", rawCursor))
	}

	refs, refCursor, err := r.codeNavSvc.GetReferences(ctx, requestArgs, r.requestState, cursor)
	if err != nil {
		return nil, errors.Wrap(err, "svc.GetReferences")
	}

	if refCursor.Phase != "done" {
		nextCursor = refCursor.Encode()
	}

	if args.Filter != nil && *args.Filter != "" {
		filtered := refs[:0]
		for _, loc := range refs {
			if strings.Contains(loc.Path.RawValue(), *args.Filter) {
				filtered = append(filtered, loc)
			}
		}
		refs = filtered
	}

	refLocs := genslices.Map(refs, shared.UploadUsage.ToLocation)
	return newLocationConnectionResolver(refLocs, pointers.NonZeroPtr(nextCursor), r.locationResolver), nil
}

package graphql

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/codenav/shared"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const DefaultReferencesPageSize = 100

// References returns the list of source locations that reference the symbol at the given position.
func (r *gitBlobLSIFDataResolver) References(ctx context.Context, args *resolverstubs.LSIFPagedQueryPositionArgs) (_ resolverstubs.LocationConnectionResolver, err error) {
	limit := int(resolverstubs.Deref(args.First, DefaultReferencesPageSize))
	if limit <= 0 {
		return nil, ErrIllegalLimit
	}

	rawCursor, err := decodeCursor(args.After)
	if err != nil {
		return nil, err
	}

	requestArgs := shared.RequestArgs{RepositoryID: r.requestState.RepositoryID, Commit: r.requestState.Commit, Path: r.requestState.Path, Line: int(args.Line), Character: int(args.Character), Limit: limit, RawCursor: rawCursor}
	ctx, _, endObservation := observeResolver(ctx, &err, r.operations.references, time.Second, getObservationArgs(requestArgs))
	defer endObservation()

	// Decode cursor given from previous response or create a new one with default values.
	// We use the cursor state track offsets with the result set and cache initial data that
	// is used to resolve each page. This cursor will be modified in-place to become the
	// cursor used to fetch the subsequent page of results in this result set.
	var nextCursor string
	cursor, err := decodeReferencesCursor(requestArgs.RawCursor)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("invalid cursor: %q", rawCursor))
	}

	refs, refCursor, err := r.codeNavSvc.GetReferences(ctx, requestArgs, r.requestState, cursor)
	if err != nil {
		return nil, errors.Wrap(err, "svc.GetReferences")
	}

	if refCursor.Phase != "done" {
		nextCursor = encodeReferencesCursor(refCursor)
	}

	if args.Filter != nil && *args.Filter != "" {
		filtered := refs[:0]
		for _, loc := range refs {
			if strings.Contains(loc.Path, *args.Filter) {
				filtered = append(filtered, loc)
			}
		}
		refs = filtered
	}

	return newLocationConnectionResolver(refs, resolverstubs.NonZeroPtr(nextCursor), r.locationResolver), nil
}

//
//

// decodeReferencesCursor is the inverse of encodeCursor. If the given encoded string is empty, then
// a fresh cursor is returned.
func decodeReferencesCursor(rawEncoded string) (shared.ReferencesCursor, error) {
	if rawEncoded == "" {
		return shared.ReferencesCursor{Phase: "local"}, nil
	}

	raw, err := base64.RawURLEncoding.DecodeString(rawEncoded)
	if err != nil {
		return shared.ReferencesCursor{}, err
	}

	var cursor shared.ReferencesCursor
	err = json.Unmarshal(raw, &cursor)
	return cursor, err
}

// encodeReferencesCursor returns an encoding of the given cursor suitable for a URL or a GraphQL token.
func encodeReferencesCursor(cursor shared.ReferencesCursor) string {
	rawEncoded, _ := json.Marshal(cursor)
	return base64.RawURLEncoding.EncodeToString(rawEncoded)
}

package graphql

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

// DefaultReferencesPageSize is the implementation result page size when no limit is supplied.
const DefaultImplementationsPageSize = 100

// ErrIllegalLimit occurs when the user requests less than one object per page.
var ErrIllegalLimit = errors.New("illegal limit")

func (r *gitBlobLSIFDataResolver) Implementations(ctx context.Context, args *resolverstubs.LSIFPagedQueryPositionArgs) (_ resolverstubs.LocationConnectionResolver, err error) {
	limit := int(pointers.Deref(args.First, DefaultImplementationsPageSize))
	if limit <= 0 {
		return nil, ErrIllegalLimit
	}

	rawCursor, err := decodeCursor(args.After)
	if err != nil {
		return nil, err
	}

	requestArgs := codenav.RequestArgs{RepositoryID: r.requestState.RepositoryID, Commit: r.requestState.Commit, Path: r.requestState.Path, Line: int(args.Line), Character: int(args.Character), Limit: limit, RawCursor: rawCursor}
	ctx, _, endObservation := observeResolver(ctx, &err, r.operations.implementations, time.Second, getObservationArgs(requestArgs))
	defer endObservation()

	// Decode cursor given from previous response or create a new one with default values.
	// We use the cursor state track offsets with the result set and cache initial data that
	// is used to resolve each page. This cursor will be modified in-place to become the
	// cursor used to fetch the subsequent page of results in this result set.
	var nextCursor string
	cursor, err := decodeImplementationsCursor(rawCursor)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("invalid cursor: %q", rawCursor))
	}

	impls, implsCursor, err := r.codeNavSvc.GetImplementations(ctx, requestArgs, r.requestState, cursor)
	if err != nil {
		return nil, errors.Wrap(err, "codeNavSvc.GetImplementations")
	}

	if implsCursor.Phase != "done" {
		nextCursor = encodeImplementationsCursor(implsCursor)
	}

	if args.Filter != nil && *args.Filter != "" {
		filtered := impls[:0]
		for _, loc := range impls {
			if strings.Contains(loc.Path, *args.Filter) {
				filtered = append(filtered, loc)
			}
		}
		impls = filtered
	}

	return newLocationConnectionResolver(impls, pointers.NonZeroPtr(nextCursor), r.locationResolver), nil
}

func (r *gitBlobLSIFDataResolver) Prototypes(ctx context.Context, args *resolverstubs.LSIFPagedQueryPositionArgs) (_ resolverstubs.LocationConnectionResolver, err error) {
	limit := int(pointers.Deref(args.First, DefaultImplementationsPageSize))
	if limit <= 0 {
		return nil, ErrIllegalLimit
	}

	rawCursor, err := decodeCursor(args.After)
	if err != nil {
		return nil, err
	}

	requestArgs := codenav.RequestArgs{RepositoryID: r.requestState.RepositoryID, Commit: r.requestState.Commit, Path: r.requestState.Path, Line: int(args.Line), Character: int(args.Character), Limit: limit, RawCursor: rawCursor}
	ctx, _, endObservation := observeResolver(ctx, &err, r.operations.prototypes, time.Second, getObservationArgs(requestArgs))
	defer endObservation()

	// Decode cursor given from previous response or create a new one with default values.
	// We use the cursor state track offsets with the result set and cache initial data that
	// is used to resolve each page. This cursor will be modified in-place to become the
	// cursor used to fetch the subsequent page of results in this result set.
	var nextCursor string
	cursor, err := decodeImplementationsCursor(rawCursor)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("invalid cursor: %q", rawCursor))
	}

	prototypes, protoCursor, err := r.codeNavSvc.GetPrototypes(ctx, requestArgs, r.requestState, cursor)
	if err != nil {
		return nil, errors.Wrap(err, "codeNavSvc.GetPrototypes")
	}

	if protoCursor.Phase != "done" {
		nextCursor = encodeImplementationsCursor(protoCursor)
	}

	if args.Filter != nil && *args.Filter != "" {
		filtered := prototypes[:0]
		for _, loc := range prototypes {
			if strings.Contains(loc.Path, *args.Filter) {
				filtered = append(filtered, loc)
			}
		}
		prototypes = filtered
	}

	return newLocationConnectionResolver(prototypes, pointers.NonZeroPtr(nextCursor), r.locationResolver), nil
}

//
//

// decodeCursor is the inverse of encodeCursor. If the given encoded string is empty, then
// a fresh cursor is returned.
func decodeImplementationsCursor(rawEncoded string) (codenav.ImplementationsCursor, error) {
	if rawEncoded == "" {
		return codenav.ImplementationsCursor{Phase: "local"}, nil
	}

	raw, err := base64.RawURLEncoding.DecodeString(rawEncoded)
	if err != nil {
		return codenav.ImplementationsCursor{}, err
	}

	var cursor codenav.ImplementationsCursor
	err = json.Unmarshal(raw, &cursor)
	return cursor, err
}

// encodeCursor returns an encoding of the given cursor suitable for a URL or a GraphQL token.
func encodeImplementationsCursor(cursor codenav.ImplementationsCursor) string {
	rawEncoded, _ := json.Marshal(cursor)
	return base64.RawURLEncoding.EncodeToString(rawEncoded)
}

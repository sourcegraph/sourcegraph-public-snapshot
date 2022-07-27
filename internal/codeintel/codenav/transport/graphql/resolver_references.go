package graphql

import (
	"context"
	"fmt"
	"time"

	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/shared"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const slowReferencesRequestThreshold = time.Second

func (r *resolver) References(ctx context.Context, args shared.RequestArgs) (_ []shared.UploadLocation, nextCursor string, err error) {
	ctx, _, endObservation := observeResolver(ctx, &err, r.operations.references, slowReferencesRequestThreshold, observation.Args{
		LogFields: []log.Field{
			log.Int("repositoryID", args.RepositoryID),
			log.String("commit", args.Commit),
			log.String("path", args.Path),
			log.Int("numUploads", len(r.requestState.GetCacheUploads())),
			log.String("uploads", uploadIDsToString(r.requestState.GetCacheUploads())),
			log.Int("line", args.Line),
			log.Int("character", args.Character),
		},
	})
	defer endObservation()

	// Decode cursor given from previous response or create a new one with default values.
	// We use the cursor state track offsets with the result set and cache initial data that
	// is used to resolve each page. This cursor will be modified in-place to become the
	// cursor used to fetch the subsequent page of results in this result set.
	cursor, err := decodeReferencesCursor(args.RawCursor)
	if err != nil {
		return nil, "", errors.Wrap(err, fmt.Sprintf("invalid cursor: %q", args.RawCursor))
	}

	refs, refCursor, err := r.svc.GetReferences(ctx, args, r.requestState, cursor)
	if err != nil {
		return nil, "", errors.Wrap(err, "svc.GetReferences")
	}

	if cursor.Phase != "done" {
		nextCursor = encodeReferencesCursor(refCursor)
	}

	return refs, nextCursor, nil
}

package resolvers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"

	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/api"
	codeintelapi "github.com/sourcegraph/sourcegraph/internal/codeintel/api"
	bundles "github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/client"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/db"
)

type lsifQueryResolver struct {
	db                  db.DB
	bundleManagerClient bundles.BundleManagerClient
	codeIntelAPI        codeintelapi.CodeIntelAPI

	repositoryResolver *graphqlbackend.RepositoryResolver
	// commit is the requested target commit
	commit api.CommitID
	path   string
	// uploads are ordered by their commit distance from the target commit
	uploads []db.Dump
}

var _ graphqlbackend.LSIFQueryResolver = &lsifQueryResolver{}

func (r *lsifQueryResolver) Definitions(ctx context.Context, args *graphqlbackend.LSIFQueryPositionArgs) (graphqlbackend.LocationConnectionResolver, error) {
	for _, upload := range r.uploads {
		// TODO(efritz) - we should also detect renames/copies on position adjustment
		adjustedPosition, ok, err := r.adjustPosition(ctx, upload.Commit, args.Line, args.Character)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}

		locations, err := r.codeIntelAPI.Definitions(ctx, r.path, adjustedPosition.Line, adjustedPosition.Character, upload.ID)
		if err != nil {
			return nil, err
		}

		if len(locations) > 0 {
			return &locationConnectionResolver{
				repo:      r.repositoryResolver.Type(),
				commit:    r.commit,
				locations: locations,
			}, nil
		}
	}

	return &locationConnectionResolver{}, nil
}

func (r *lsifQueryResolver) References(ctx context.Context, args *graphqlbackend.LSIFPagedQueryPositionArgs) (graphqlbackend.LocationConnectionResolver, error) {
	// Decode a map of upload ids to the next url that serves
	// the new page of results. This may not include an entry
	// for every upload if their result sets have already been
	// exhausted.
	nextURLs, err := readCursor(args.After)
	if err != nil {
		return nil, err
	}

	// We need to maintain a symmetric map for the next page
	// of results that we can encode into the endCursor of
	// this request.
	newCursors := map[int]string{}

	var allLocations []codeintelapi.ResolvedLocation
	for _, upload := range r.uploads {
		adjustedPosition, ok, err := r.adjustPosition(ctx, upload.Commit, args.Line, args.Character)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}

		limit := DefaultReferencesPageSize
		if args.First != nil {
			limit = int(*args.First)
		}
		if limit <= 0 {
			// TODO(efritz) - check on defs too
			return nil, errors.New("illegal limit")
		}

		rawCursor := ""
		if nextURL, ok := nextURLs[upload.ID]; ok {
			rawCursor = nextURL
		} else if len(nextURLs) != 0 {
			// Result set is exhausted or newer than the first page
			// of results. Skip anything from this upload as it will
			// have duplicate results, or it will be out of order.
			continue
		}

		cursor, err := codeintelapi.DecodeOrCreateCursor(r.path, adjustedPosition.Line, adjustedPosition.Character, upload.ID, rawCursor, r.db, r.bundleManagerClient)
		if err != nil {
			return nil, err
		}

		locations, newCursor, hasNewCursor, err := r.codeIntelAPI.References(
			ctx,
			int(r.repositoryResolver.Type().ID),
			string(r.commit),
			limit,
			cursor,
		)
		if err != nil {
			return nil, err
		}

		cx := ""
		if hasNewCursor {
			cx = codeintelapi.EncodeCursor(newCursor)
		}

		allLocations = append(allLocations, locations...)

		if cx != "" {
			newCursors[upload.ID] = cx
		}
	}

	endCursor, err := makeCursor(newCursors)
	if err != nil {
		return nil, err
	}

	return &locationConnectionResolver{
		repo:      r.repositoryResolver.Type(),
		commit:    r.commit,
		locations: allLocations,
		endCursor: endCursor,
	}, nil
}

func (r *lsifQueryResolver) Hover(ctx context.Context, args *graphqlbackend.LSIFQueryPositionArgs) (graphqlbackend.HoverResolver, error) {
	for _, upload := range r.uploads {
		adjustedPosition, ok, err := r.adjustPosition(ctx, upload.Commit, args.Line, args.Character)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}

		// TODO(efritz) - codeintelapi should just return an lsp.Hover
		text, rn, exists, err := r.codeIntelAPI.Hover(ctx, r.path, adjustedPosition.Line, adjustedPosition.Character, int(upload.ID))
		if err != nil || !exists {
			return nil, err
		}

		lspRange := convertRange(rn)

		if text != "" {
			adjustedRange, ok, err := r.adjustRange(ctx, upload.Commit, lspRange)
			if err != nil {
				return nil, err
			}
			if !ok {
				// Failed to adjust range. This _might_ happen in cases where the LSIF range
				// spans multiple lines which intersect a diff; the hover position on an earlier
				// line may not be edited, but the ending line of the expression may have been
				// edited or removed. This is rare and unfortunate, and we'll skip the result
				// in this case because we have low confidence that it will be rendered correctly.
				continue
			}

			return &hoverResolver{text: text, lspRange: adjustedRange}, nil
		}
	}

	return nil, nil
}

// adjustPosition adjusts the position denoted by `line` and `character` in the requested commit into an
// LSP position in the upload commit. This method returns nil if no equivalent position is found.
func (r *lsifQueryResolver) adjustPosition(ctx context.Context, uploadCommit string, line, character int32) (lsp.Position, bool, error) {
	adjuster, err := newPositionAdjuster(ctx, r.repositoryResolver.Type(), string(r.commit), uploadCommit, r.path)
	if err != nil {
		return lsp.Position{}, false, err
	}

	adjusted, ok := adjuster.adjustPosition(lsp.Position{Line: int(line), Character: int(character)})
	return adjusted, ok, nil
}

// adjustPosition adjusts the given range in the upload commit into an equivalent range in the requested
// commit. This method returns nil if there is not an equivalent position for both endpoints of the range.
func (r *lsifQueryResolver) adjustRange(ctx context.Context, uploadCommit string, lspRange lsp.Range) (lsp.Range, bool, error) {
	adjuster, err := newPositionAdjuster(ctx, r.repositoryResolver.Type(), uploadCommit, string(r.commit), r.path)
	if err != nil {
		return lsp.Range{}, false, err
	}

	adjusted, ok := adjuster.adjustRange(lspRange)
	return adjusted, ok, nil
}

// readCursor decodes a cursor into a map from upload ids to URLs that
// serves the next page of results.
func readCursor(after *string) (map[int]string, error) {
	if after == nil {
		return nil, nil
	}

	decoded, err := base64.StdEncoding.DecodeString(*after)
	if err != nil {
		return nil, err
	}

	var cursors map[int]string
	if err := json.Unmarshal(decoded, &cursors); err != nil {
		return nil, err
	}
	return cursors, nil
}

// makeCursor encodes a map from upload ids to URLs that serves the next
// page of results into a single string that can be sent back for use in
// cursor pagination.
func makeCursor(cursors map[int]string) (string, error) {
	if len(cursors) == 0 {
		return "", nil
	}

	encoded, err := json.Marshal(cursors)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(encoded), nil
}

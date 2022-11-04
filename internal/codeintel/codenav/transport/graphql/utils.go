package graphql

import (
	"context"
	"encoding/base64"

	"github.com/sourcegraph/go-lsp"

	"github.com/sourcegraph/sourcegraph/internal/api"
	sharedresolvers "github.com/sourcegraph/sourcegraph/internal/codeintel/shared/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

// convertRange creates an LSP range from a bundle range.
func convertRange(r types.Range) lsp.Range {
	return lsp.Range{Start: convertPosition(r.Start.Line, r.Start.Character), End: convertPosition(r.End.Line, r.End.Character)}
}

func convertPosition(line, character int) lsp.Position {
	return lsp.Position{Line: line, Character: character}
}

func sharedRangeTolspRange(r types.Range) lsp.Range {
	return lsp.Range{Start: convertPosition(r.Start.Line, r.Start.Character), End: convertPosition(r.End.Line, r.End.Character)}
}

// strPtr creates a pointer to the given value. If the value is an
// empty string, a nil pointer is returned.
func strPtr(val string) *string {
	if val == "" {
		return nil
	}

	return &val
}

// DecodeCursor decodes the given cursor value. It is assumed to be a value previously
// returned from the function encodeCursor. An empty string is returned if no cursor is
// supplied. Invalid cursors return errors.
func DecodeCursor(val *string) (string, error) {
	if val == nil {
		return "", nil
	}

	decoded, err := base64.StdEncoding.DecodeString(*val)
	if err != nil {
		return "", err
	}

	return string(decoded), nil
}

// derefInt32 returns the underlying value in the given pointer.
// If the pointer is nil, the default value is returned.
func derefInt32(val *int32, defaultValue int) int {
	if val != nil {
		return int(*val)
	}
	return defaultValue
}

// EncodeCursor creates a PageInfo object from the given cursor. If the cursor is not
// defined, then an object indicating the end of the result set is returned. The cursor
// is base64 encoded for transfer, and should be decoded using the function decodeCursor.
func EncodeCursor(val *string) *PageInfo {
	if val != nil {
		return NextPageCursor(base64.StdEncoding.EncodeToString([]byte(*val)))
	}

	return HasNextPage(false)
}

// PageInfo implements the GraphQL type PageInfo.
type PageInfo struct {
	endCursor   *string
	hasNextPage bool
}

// HasNextPage returns a new PageInfo with the given hasNextPage value.
func HasNextPage(hasNextPage bool) *PageInfo {
	return &PageInfo{hasNextPage: hasNextPage}
}

// NextPageCursor returns a new PageInfo indicating there is a next page with
// the given end cursor.
func NextPageCursor(endCursor string) *PageInfo {
	return &PageInfo{endCursor: &endCursor, hasNextPage: true}
}

func (r *PageInfo) EndCursor() *string { return r.endCursor }
func (r *PageInfo) HasNextPage() bool  { return r.hasNextPage }

func sharedDumpToDbstoreUpload(dump types.Dump) types.Upload {
	return types.Upload{
		ID:                dump.ID,
		Commit:            dump.Commit,
		Root:              dump.Root,
		VisibleAtTip:      dump.VisibleAtTip,
		UploadedAt:        dump.UploadedAt,
		State:             dump.State,
		FailureMessage:    dump.FailureMessage,
		StartedAt:         dump.StartedAt,
		FinishedAt:        dump.FinishedAt,
		ProcessAfter:      dump.ProcessAfter,
		NumResets:         dump.NumResets,
		NumFailures:       dump.NumFailures,
		RepositoryID:      dump.RepositoryID,
		RepositoryName:    dump.RepositoryName,
		Indexer:           dump.Indexer,
		IndexerVersion:    dump.IndexerVersion,
		NumParts:          0,
		UploadedParts:     []int{},
		UploadSize:        nil,
		Rank:              nil,
		AssociatedIndexID: dump.AssociatedIndexID,
	}
}

// ConnectionArgs is the common set of arguments to GraphQL fields that return connections (lists).
type ConnectionArgs struct {
	First *int32 // return the first n items
}

// Set is a convenience method for setting the DB limit and offset in a DB XyzListOptions struct.
func (a ConnectionArgs) Set(o **database.LimitOffset) {
	if a.First != nil {
		*o = &database.LimitOffset{Limit: int(*a.First)}
	}
}

// GetFirst is a convenience method returning the value of First, defaulting to
// the type's zero value if nil.
func (a ConnectionArgs) GetFirst() int32 {
	if a.First == nil {
		return 0
	}
	return *a.First
}

// resolveLocations creates a slide of LocationResolvers for the given list of adjusted locations. The
// resulting list may be smaller than the input list as any locations with a commit not known by
// gitserver will be skipped.
func resolveLocations(ctx context.Context, locationResolver *sharedresolvers.CachedLocationResolver, locations []types.UploadLocation) ([]LocationResolver, error) {
	resolvedLocations := make([]LocationResolver, 0, len(locations))
	for i := range locations {
		resolver, err := resolveLocation(ctx, locationResolver, locations[i])
		if err != nil {
			return nil, err
		}
		if resolver == nil {
			continue
		}

		resolvedLocations = append(resolvedLocations, resolver)
	}

	return resolvedLocations, nil
}

// resolveLocation creates a LocationResolver for the given adjusted location. This function may return a
// nil resolver if the location's commit is not known by gitserver.
func resolveLocation(ctx context.Context, locationResolver *sharedresolvers.CachedLocationResolver, location types.UploadLocation) (LocationResolver, error) {
	treeResolver, err := locationResolver.Path(ctx, api.RepoID(location.Dump.RepositoryID), location.TargetCommit, location.Path)
	if err != nil || treeResolver == nil {
		return nil, err
	}

	lspRange := convertRange(location.TargetRange)
	return NewLocationResolver(treeResolver, &lspRange), nil
}

package graphql

import (
	"context"
	"encoding/base64"
	"encoding/json"

	"github.com/sourcegraph/scip/bindings/go/scip"

	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func (c *codeGraphDataResolver) Occurrences(ctx context.Context, args *resolverstubs.OccurrencesArgs) (_ resolverstubs.SCIPOccurrenceConnectionResolver, err error) {
	_, _, endObservation := c.operations.occurrences.WithErrors(ctx, &err, observation.Args{Attrs: c.opts.Attrs()})
	defer endObservation(1, observation.Args{})

	const maxPageSize = 100000
	args = args.Normalize(maxPageSize)

	impl, err := gqlutil.NewConnectionResolver[resolverstubs.SCIPOccurrenceResolver](
		&occurrenceConnectionStore{c},
		&gqlutil.ConnectionResolverArgs{First: args.First, After: args.After},
		&gqlutil.ConnectionResolverOptions{MaxPageSize: maxPageSize, Reverse: pointers.Ptr(false)})
	if err != nil {
		return nil, err
	}
	return &occurrenceConnectionResolver{impl, c, args}, nil
}

type occurrenceConnectionResolver struct {
	impl *gqlutil.ConnectionResolver[resolverstubs.SCIPOccurrenceResolver]

	// Arguments
	graphData *codeGraphDataResolver
	args      *resolverstubs.OccurrencesArgs
}

var _ resolverstubs.SCIPOccurrenceConnectionResolver = &occurrenceConnectionResolver{}

func (o *occurrenceConnectionResolver) Nodes(ctx context.Context) ([]resolverstubs.SCIPOccurrenceResolver, error) {
	return o.impl.Nodes(ctx)
}

func (o *occurrenceConnectionResolver) PageInfo(ctx context.Context) (*gqlutil.ConnectionPageInfo[resolverstubs.SCIPOccurrenceResolver], error) {
	return o.impl.PageInfo(ctx)
}

var _ gqlutil.ConnectionResolverStore[resolverstubs.SCIPOccurrenceResolver] = &occurrenceConnectionStore{}

type scipOccurrence struct {
	impl *scip.Occurrence

	// For cursor state, because a single value is passed to MarshalCursor
	cursor
}

func (s scipOccurrence) Roles() (*[]resolverstubs.SymbolRole, error) {
	roles := s.impl.GetSymbolRoles()
	out := []resolverstubs.SymbolRole{}
	if roles&int32(scip.SymbolRole_Definition) != 0 {
		out = append(out, resolverstubs.SymbolRoleDefinition)
	} else {
		out = append(out, resolverstubs.SymbolRoleReference)
	}
	if roles&int32(scip.SymbolRole_ForwardDefinition) != 0 {
		out = append(out, resolverstubs.SymbolRoleForwardDefinition)
	}
	return &out, nil
}

var _ resolverstubs.SCIPOccurrenceResolver = scipOccurrence{}

type occurrenceConnectionStore struct {
	graphData *codeGraphDataResolver
}

var _ gqlutil.ConnectionResolverStore[resolverstubs.SCIPOccurrenceResolver] = &occurrenceConnectionStore{}

func (o *occurrenceConnectionStore) ComputeTotal(ctx context.Context) (int32, error) {
	doc, err := o.graphData.tryRetrieveDocument(ctx)
	if doc == nil || err != nil {
		return 0, err
	}
	return int32(len(doc.Occurrences)), nil
}

func (o *occurrenceConnectionStore) ComputeNodes(ctx context.Context, paginationArgs *database.PaginationArgs) ([]resolverstubs.SCIPOccurrenceResolver, error) {
	doc, err := o.graphData.tryRetrieveDocument(ctx)
	if err != nil {
		return nil, err
	}
	if paginationArgs != nil {
		// Strictly speaking, 'After' is expected to have length 0 or 1,
		// but handling the general case to avoid panicking or returning an
		// error in the multiple element case.
		for i := range paginationArgs.After {
			if c, ok := paginationArgs.After[i].(cursor); ok {
				paginationArgs.After[i] = c.Index
			}
		}
	}
	occs, _, err2 := database.OffsetBasedCursorSlice(doc.Occurrences, paginationArgs)
	if err2 != nil {
		return nil, err2
	}

	out := make([]resolverstubs.SCIPOccurrenceResolver, 0, len(occs))
	for idx, occ := range occs {
		out = append(out, scipOccurrence{occ, cursor{idx}})
	}
	return out, nil
}

func (o *occurrenceConnectionStore) MarshalCursor(n resolverstubs.SCIPOccurrenceResolver, _ database.OrderBy) (*string, error) {
	return marshalCursor(n.(scipOccurrence).cursor)
}

func marshalCursor(c cursor) (*string, error) {
	buf, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	return pointers.Ptr(base64.StdEncoding.EncodeToString(buf)), nil
}

func (o *occurrenceConnectionStore) UnmarshalCursor(s string, _ database.OrderBy) ([]any, error) {
	buf, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}
	var c cursor
	if err = json.Unmarshal(buf, &c); err != nil {
		return nil, err
	}
	return []any{c}, nil
}

type cursor struct {
	// Index inside occurrences array in document
	Index int
}

func (s scipOccurrence) Symbol() (*string, error) {
	return pointers.Ptr(s.impl.Symbol), nil
}

func (s scipOccurrence) Range() (resolverstubs.RangeResolver, error) {
	// FIXME(issue: GRAPH-571): Below code is correct iff the indexer uses UTF-16 offsets
	return newRangeResolver(scip.NewRangeUnchecked(s.impl.Range)), nil
}
